package avitohandler

import (
	"context"
	"encoding/json"
	"fmt"
	"hr-tools-backend/db"
	applicanthistoryhandler "hr-tools-backend/lib/applicant-history"
	applicantstore "hr-tools-backend/lib/applicant/store"
	externalservices "hr-tools-backend/lib/external-services"
	avitoclient "hr-tools-backend/lib/external-services/avito/client"
	extservicestore "hr-tools-backend/lib/external-services/store"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	"hr-tools-backend/lib/utils/helpers"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	"hr-tools-backend/models"
	avitoapimodels "hr-tools-backend/models/api/avito"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var Instance externalservices.JobSiteProvider

func NewHandler() {
	Instance = &impl{
		client:             avitoclient.Instance,
		extStore:           extservicestore.NewInstance(db.DB),
		spaceUserStore:     spaceusersstore.NewInstance(db.DB),
		vacancyStore:       vacancystore.NewInstance(db.DB),
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
		applicantStore:     applicantstore.NewInstance(db.DB),
		applicantHistory:   applicanthistoryhandler.Instance,
		tokenMap:           sync.Map{},
	}
}

type impl struct {
	client             avitoclient.Provider
	extStore           extservicestore.Provider
	spaceUserStore     spaceusersstore.Provider
	vacancyStore       vacancystore.Provider
	spaceSettingsStore spacesettingsstore.Provider
	applicantStore     applicantstore.Provider
	applicantHistory   applicanthistoryhandler.Provider
	tokenMap           sync.Map
}

const (
	TokenCode              = "AVITO_TOKEN"
	LastApplicationDateTpl = "AVITO_LAST_APPL_DATE:%v"
)

func (i *impl) getLogger(spaceID, vacancyID string) *log.Entry {
	logger := log.WithField("integration", "Avito")
	if spaceID != "" {
		logger = logger.WithField("space_id", spaceID)
	}
	if vacancyID != "" {
		logger = logger.WithField("vacancy_id", vacancyID)
	}
	return logger
}

func (i *impl) GetConnectUri(spaceID string) (uri string, err error) {
	clientID, err := i.getValue(spaceID, models.AvitoClientIDSetting)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения настройки ClientID для Avito")
	}
	_, err = i.getValue(spaceID, models.AvitoClientSecretSetting)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения настройки ClientSecret для Avito")
	}
	return i.client.GetLoginUri(clientID, spaceID)
}

func (i *impl) RequestToken(spaceID, code string) {
	logger := i.getLogger(spaceID, "")
	clientID, err := i.getValue(spaceID, models.AvitoClientIDSetting)
	if err != nil {
		logger.WithError(err).Error("ошибка получения настройки ClientID для Avito")
		return
	}

	clientSecret, err := i.getValue(spaceID, models.AvitoClientSecretSetting)
	if err != nil {
		logger.WithError(err).Error("ошибка получения настройки ClientSecret для Avito")
		return
	}

	req := avitoapimodels.RequestToken{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Code:         code,
	}
	token, err := i.client.RequestToken(context.TODO(), req)
	if err != nil {
		logger.WithError(err).Error("ошибка получения токена Avito")
		return
	}
	err = i.storeToken(spaceID, *token, true)
	if err != nil {
		logger.Error(err)
	}
}

func (i *impl) CheckConnected(spaceID string) bool {
	_, ok := i.tokenMap.Load(spaceID)
	if ok {
		return true
	}
	data, ok, err := i.extStore.Get(spaceID, TokenCode)
	if err != nil || !ok {
		return false
	}
	token := avitoapimodels.ResponseToken{}
	err = json.Unmarshal(data, &token)
	if err != nil {
		return false
	}
	i.storeToken(spaceID, token, false)
	return true
}

func (i *impl) VacancyPublish(ctx context.Context, spaceID, vacancyID string) error {
	logger := i.getLogger(spaceID, vacancyID)

	accessToken, err := i.getToken(ctx, spaceID)
	if err != nil {
		return err
	}

	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("вакансия не найдена")
	}

	if models.VacancyStatusOpened != rec.Status {
		return errors.Errorf("неподходящей статус вакансии %v, для публикации в НН", rec.Status)
	}

	if rec.AvitoID != 0 || rec.AvitoPublishID != "" {
		if rec.AvitoStatus != models.VacancyPubStatusNone && rec.AvitoStatus != models.VacancyPubStatusClosed {
			return errors.New("вакансия уже размещена")
		}
	}

	request, err := i.fillVacancyData(rec)
	if err != nil {
		return err
	}

	id, err := i.client.VacancyPublish(ctx, accessToken, *request)
	if err != nil {
		return err
	}
	updMap := map[string]interface{}{
		"avito_publish_id": id,
		"avito_id":         nil,
		"avito_uri":        nil,
		"avito_reasons":    nil,
		"avito_status":     models.VacancyPubStatusModeration,
	}
	err = i.vacancyStore.Update(spaceID, vacancyID, updMap)
	if err != nil {
		errMsg := errors.Errorf("не удалось сохранить идентификатор опубликованной вакансии (%v)", id)
		logger.WithError(err).Error(errMsg)
		return errMsg
	}
	return nil
}

func (i *impl) VacancyUpdate(ctx context.Context, spaceID, vacancyID string) error {
	logger := i.getLogger(spaceID, vacancyID)
	accessToken, err := i.getToken(ctx, spaceID)
	if err != nil {
		return err
	}

	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return err
	}
	err = allowChange(rec, true)
	if err != nil {
		return err
	}

	request, err := i.fillVacancyData(rec)
	if err != nil {
		return err
	}

	id, err := i.client.VacancyUpdate(ctx, accessToken, rec.AvitoPublishID, rec.AvitoID, *request)
	if err != nil {
		return err
	}
	updMap := map[string]interface{}{
		"avito_publish_id": id,
		"avito_reasons":    nil,
		"avito_status":     models.VacancyPubStatusModeration,
	}
	err = i.vacancyStore.Update(spaceID, vacancyID, updMap)
	if err != nil {
		errMsg := errors.Errorf("не удалось сохранить идентификатор опубликованной вакансии (%v)", id)
		logger.WithError(err).Error(errMsg)
		return errMsg
	}
	return nil
}

func (i *impl) VacancyClose(ctx context.Context, spaceID, vacancyID string) error {
	accessToken, err := i.getToken(ctx, spaceID)
	if err != nil {
		return err
	}

	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return err
	}
	err = allowChange(rec, false)
	if err != nil {
		return err
	}
	return i.client.VacancyClose(ctx, accessToken, rec.AvitoID)
}

func (i *impl) VacancyAttach(ctx context.Context, spaceID, vacancyID string, extID string) error {
	avitoID, err := strconv.Atoi(extID)
	if err != nil {
		return errors.New("указане некорректный идентификатор вакансии")
	}
	accessToken, err := i.getToken(ctx, spaceID)
	if err != nil {
		return err
	}
	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("вакансия не найдена")
	}
	if rec.AvitoID != 0 {
		return errors.New("ссылка на вакансию уже добавлена")
	}
	if models.VacancyStatusOpened != rec.Status {
		return errors.Errorf("неподходящей статус вакансии: %v", rec.Status)
	}
	data, err := i.client.GetVacancy(ctx, accessToken, avitoID)
	if err != nil {
		return err
	}
	if !data.IsActive {
		return errors.New("указанная вакансия уже не активна")
	}
	updMap := map[string]interface{}{
		"avito_id":         data.ID,
		"avito_uri":        data.Url,
		"avito_publish_id": nil,
		"avito_reasons":    nil,
		"avito_status":     models.VacancyPubStatusPublished,
	}
	logger := i.getLogger(spaceID, vacancyID)
	err = i.vacancyStore.Update(spaceID, vacancyID, updMap)
	if err != nil {
		errMsg := errors.Errorf("не удалось обновить данные опубликованной вакансии (%v)", data.ID)
		logger.WithError(err).Error(errMsg)
		return errMsg
	}
	return nil
}

func (i *impl) GetVacancyInfo(ctx context.Context, spaceID, vacancyID string) (*vacancyapimodels.ExtVacancyInfo, error) {
	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errors.New("вакансия не найдена")
	}
	result := vacancyapimodels.ExtVacancyInfo{
		Url:    rec.AvitoUri,
		Status: models.VacancyPubStatusNone,
	}
	if rec.AvitoID == 0 && rec.AvitoPublishID == "" {
		return &result, nil
	}
	result.Status = rec.AvitoStatus
	result.Reason = rec.AvitoReasons
	return &result, nil
}

func (i *impl) HandleNegotiations(ctx context.Context, data dbmodels.Vacancy) error {
	accessToken, err := i.getToken(ctx, data.SpaceID)
	if err != nil {
		return err
	}
	logger := i.getLogger(data.SpaceID, data.ID)
	lastDateKey := fmt.Sprintf(LastApplicationDateTpl, data.ID)

	updatedAt := data.CreatedAt.Format("2006-01-02")
	lastDate, ok, err := i.extStore.Get(data.SpaceID, lastDateKey)
	if err != nil {
		logger.WithError(err).Warn("не удалось получить дату последнего обновления")
	} else if ok {
		updatedAt = string(lastDate)
	}

	idsResp, err := i.client.GetApplicationIDs(ctx, accessToken, updatedAt, "", data.AvitoID)
	if err != nil {
		return errors.Wrap(err, "ошибка получения идентификаторов откликов")
	}
	if idsResp == nil || len(idsResp.Applies) == 0 {
		return nil
	}
	ids := []string{}
	for _, applie := range idsResp.Applies {
		negotiationID := applie.ID
		found, err := i.applicantStore.IsExistNegotiationID(data.SpaceID, negotiationID, models.ApplicantSourceAvito)
		if err != nil {
			logger.WithError(err).Error("не удалось проверить наличие отклика")
			continue
		}
		if found {
			continue
		}

		ids = append(ids, negotiationID)
		updatedDate, ok := applie.GetUpdatedAt()
		if ok {
			updatedAt = updatedDate.Format("2006-01-02")
		}
	}
	if len(ids) == 0 {
		// из полученных все добавлены
		return nil
	}
	applicationResp, err := i.client.GetApplicationByIDs(ctx, accessToken, ids, data.AvitoID)
	if err != nil {
		return errors.Wrap(err, "ошибка получения списка откликов")
	}
	for _, apply := range applicationResp.Applies {
		if apply.Applicant.ResumeID == 0 {
			logger.
				WithField("negotiation_id", apply.ID).
				Info("отклик без резюме")
			continue
		}
		resume, err := i.client.GetResume(ctx, accessToken, data.AvitoID, apply.Applicant.ResumeID)
		if err != nil {
			logger.WithError(err).Error("ошибка получения резюме по отклику")
			continue
		}
		i.storeApplicant(resume, apply, data)
	}
	i.extStore.Set(data.SpaceID, lastDateKey, []byte(updatedAt))
	return nil
}

func (i *impl) storeApplicant(resume *avitoapimodels.Resume, apply avitoapimodels.Applies, data dbmodels.Vacancy) {
	logger := i.getLogger(data.SpaceID, data.ID).
		WithField("negotiation_id", apply.ID)

	applicantData := dbmodels.Applicant{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: data.SpaceID,
		},
		VacancyID:       data.ID,
		NegotiationID:   apply.ID,
		ExtApplicantID:  apply.Applicant.ID,
		ResumeID:        strconv.Itoa(resume.ID),
		Source:          models.ApplicantSourceAvito,
		NegotiationDate: time.Now(),
		Status:          models.ApplicantStatusNegotiation,
		FirstName:       apply.Applicant.Data.FullName.FirstName,
		LastName:        apply.Applicant.Data.FullName.LastName,
		MiddleName:      apply.Applicant.Data.FullName.Patronymic,
		ResumeTitle:     resume.Title,
		Salary:          resume.Salary,
		Address:         resume.Params.Address,
		Gender:          resume.Params.GetGender(),
		Relocation:      resume.Params.GetRelocationType(),
		Email:           "", //нет данных
		TotalExperience: 0,  //опыт работ в месяцах - нет данных
		Params: dbmodels.ApplicantParams{
			Education:               resume.Params.GetEducationType(),
			HaveAdditionalEducation: false,
			Employments:             []models.Employment{}, //нет данных
			Schedules:               []models.Schedule{},   //нет данных
			Languages:               []dbmodels.Language{},
			TripReadiness:           resume.Params.GetTripReadinessType(),
			DriverLicenseTypes:      resume.Params.GetDriverLicence(),
			SearchStatus:            "", //нет данных
		},
	}
	for _, stage := range data.SelectionStages {
		if stage.Name == dbmodels.NegotiationStage {
			applicantData.SelectionStageID = stage.ID
			break
		}
	}
	birthDate, err := apply.GetBirthDate()
	if err != nil {
		logger.WithError(err).Error("ошибка получения даты рождения кандидата")
	}
	applicantData.BirthDate = birthDate
	if len(apply.Contacts.Phones) > 0 {
		applicantData.Phone = strconv.Itoa(apply.Contacts.Phones[0].Value)
	}

	applicantData.Citizenship = apply.Applicant.Data.Citizenship
	for _, language := range resume.Params.LanguageList {
		lng := dbmodels.Language{
			Name:          language.Language,
			LanguageLevel: language.GetLanguageLevelType(),
		}
		applicantData.Params.Languages = append(applicantData.Params.Languages, lng)
	}
	applicantID, err := i.applicantStore.Create(applicantData)
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения кандидата по отклику")
	}
	changes := applicanthistoryhandler.GetCreateChanges("Кандидат добавлен с работного сайта на вакансию", applicantData)
	i.applicantHistory.Save(applicantData.SpaceID, applicantID, applicantData.VacancyID, "", dbmodels.HistoryTypeNegotiation, changes)
}

func (i *impl) getValue(spaceID string, code models.SpaceSettingCode) (string, error) {
	return i.spaceSettingsStore.GetValueByCode(spaceID, code)
}

func (i *impl) storeToken(spaceID string, token avitoapimodels.ResponseToken, inDb bool) error {
	tokenData := avitoapimodels.TokenData{
		ResponseToken: token,
		ExpiresAt:     time.Now(),
	}
	i.tokenMap.Store(spaceID, tokenData)
	if inDb {
		data, err := json.Marshal(token)
		err = i.extStore.Set(spaceID, TokenCode, data)
		if err != nil {
			return errors.Wrap(err, "ошибка сохранения токена Avito в бд")
		}
	}
	return nil
}

func (i *impl) getToken(ctx context.Context, spaceID string) (string, error) {
	if !i.CheckConnected(spaceID) {
		return "", errors.New("Avito не подключен")
	}
	value, ok := i.tokenMap.Load(spaceID)
	if !ok {
		return "", errors.New("Avito не подключен")
	}
	tokenData := value.(avitoapimodels.TokenData)
	if time.Now().After(tokenData.ExpiresAt) {
		clientID, err := i.getValue(spaceID, models.AvitoClientIDSetting)
		if err != nil {
			return "", errors.Wrap(err, "ошибка получения настройки ClientID для Avito")
		}

		clientSecret, err := i.getValue(spaceID, models.AvitoClientSecretSetting)
		if err != nil {
			return "", errors.Wrap(err, "ошибка получения настройки ClientSecret для Avito")
		}
		req := avitoapimodels.RefreshToken{
			RefreshToken: tokenData.RefreshToken,
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}
		tokenResp, err := i.client.RefreshToken(ctx, req)
		if err != nil {
			return "", errors.New("ошибка получения токена для Avito")
		}
		err = i.storeToken(spaceID, *tokenResp, true)
		if err != nil {
			return "", errors.New("ошибка сохранения токена для Avito")
		}
	}
	return tokenData.AccessToken, nil
}

func (i *impl) fillVacancyData(rec *dbmodels.Vacancy) (*avitoapimodels.VacancyPubRequest, error) {
	if rec.City == nil {
		return nil, errors.New("не указан город публикации")
	}

	if rec.Department == nil {
		return nil, errors.New("для публикации на Avito, необходимо указать подразделение")
	}
	if rec.Experience == "" {
		return nil, errors.New("для публикации на Avito, необходимо указать опыт работы")
	}
	if rec.Schedule == "" {
		return nil, errors.New("для публикации на Avito, необходимо указать режим работы")
	}
	if rec.Employment == "" {
		return nil, errors.New("для публикации на Avito, необходимо указать занятость")
	}
	if len(rec.VacancyName) > 50 {
		return nil, errors.New("для публикации на Avito, название вакансии не должно превышать 50 символов")
	}
	if len(rec.Requirements) < 200 {
		return nil, errors.New("для публикации на Avito, необходимо указать описание не более 5000 символов")
	}
	request := avitoapimodels.VacancyPubRequest{
		ApplyProcessing: avitoapimodels.ApplyProcessing{
			ApplyType: avitoapimodels.ApplyTypeWithResume,
		},
		BillingType:  "package",
		BusinessArea: rec.Department.BusinessAreaID,
		Description:  rec.Requirements,
		Employment:   rec.Employment,
		Experience:   rec.Experience,
		Location: avitoapimodels.Location{
			Address: avitoapimodels.LocationAddress{
				Locality: rec.City.City,
			},
		},
		Schedule: rec.Schedule,
		Title:    rec.VacancyName,
	}
	if rec.Salary.From != 0 || rec.Salary.To != 0 {
		request.SalaryRange = &avitoapimodels.SalaryRange{
			From: rec.From,
			To:   rec.To,
		}
	} else if rec.Salary.InHand != 0 {
		request.SalaryRange = &avitoapimodels.SalaryRange{
			From: rec.Salary.InHand,
			To:   rec.Salary.InHand,
		}
	}
	return &request, nil
}

func allowChange(rec *dbmodels.Vacancy, isEdit bool) error {
	if rec == nil {
		return errors.New("вакансия не найдена")
	}

	if rec.AvitoID == 0 && rec.AvitoPublishID == "" {
		return errors.New("вакансия еще не размещалась")
	}

	if rec.AvitoID == 0 {
		return errors.New("вакансия размещена, но еще не опубликованна")
	}

	if isEdit && rec.AvitoPublishID == "" {
		return errors.New("вакансия недоступна для редактирования")
	}
	return nil
}

func (i *impl) GetCheckList(ctx context.Context, spaceID string, status models.VacancyPubStatus) ([]dbmodels.Vacancy, error) {
	return i.vacancyStore.ListAvitoByStatus(spaceID, status)
}

func (i *impl) CheckIsModerationDone(ctx context.Context, spaceID string, list []dbmodels.Vacancy) error {
	accessToken, err := i.getToken(ctx, spaceID)
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(list))
	vacancyMap := map[string]dbmodels.Vacancy{}
	for _, rec := range list {
		ids = append(ids, rec.AvitoPublishID)
		vacancyMap[rec.AvitoPublishID] = rec
	}
	req := avitoapimodels.StatusRequest{
		IDs: nil,
	}
	logger := i.getLogger(spaceID, "")
	statusResp, err := i.client.VacancyStatus(ctx, accessToken, req)
	if err != nil {
		return err
	}
	if statusResp == nil {
		return nil
	}
	for _, status := range *statusResp {
		if helpers.IsContextDone(ctx) {
			return nil
		}
		rec, ok := vacancyMap[status.ID]
		if !ok {
			continue
		}
		newStatus := status.Vacancy.GetPubStatus()
		if newStatus == rec.AvitoStatus {
			continue
		}
		updMap := map[string]interface{}{
			"avito_id":      status.Vacancy.ID,
			"avito_uri":     status.Vacancy.Url,
			"avito_reasons": fmt.Sprintf("%+v", status.Vacancy.Reasons),
			"avito_status":  newStatus,
		}
		err = i.vacancyStore.Update(spaceID, rec.ID, updMap)
		if err != nil {
			logger.
				WithError(err).
				Error("ошибка обновления статуса публикации")
		}
	}
	return nil
}

func (i *impl) CheckIsActivePublications(ctx context.Context, spaceID string, list []dbmodels.Vacancy) error {
	accessToken, err := i.getToken(ctx, spaceID)
	if err != nil {
		return err
	}
	for _, rec := range list {
		if helpers.IsContextDone(ctx) {
			return nil
		}
		logger := i.getLogger(spaceID, rec.ID)
		vacancyInfo, err := i.client.GetVacancy(ctx, accessToken, rec.AvitoID)
		if err != nil {
			logger.
				WithError(err).
				WithField("avito_vacancy_id", rec.AvitoID).
				Error("не удалось проверить статус публикации вакансии")
			continue
		}
		if !vacancyInfo.IsActive {
			updMap := map[string]interface{}{
				"avito_status": models.VacancyPubStatusClosed,
			}
			err = i.vacancyStore.Update(spaceID, rec.ID, updMap)
			if err != nil {
				logger.
					WithError(err).
					Error("ошибка обновления статуса публикации")
			}
		}
	}
	return nil
}
