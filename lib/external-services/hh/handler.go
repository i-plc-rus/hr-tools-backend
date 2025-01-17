package hhhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"hr-tools-backend/db"
	applicanthistoryhandler "hr-tools-backend/lib/applicant-history"
	applicantstore "hr-tools-backend/lib/applicant/store"
	externalservices "hr-tools-backend/lib/external-services"
	"hr-tools-backend/lib/external-services/hh/hhclient"
	extservicestore "hr-tools-backend/lib/external-services/store"
	filestorage "hr-tools-backend/lib/file-storage"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	"hr-tools-backend/lib/utils/helpers"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	"hr-tools-backend/models"
	hhapimodels "hr-tools-backend/models/api/hh"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var Instance externalservices.JobSiteProvider

func NewHandler() {
	Instance = &impl{
		//client:             hhclient.Instance,
		client:             hhclient.Instance,
		extStore:           extservicestore.NewInstance(db.DB),
		spaceUserStore:     spaceusersstore.NewInstance(db.DB),
		vacancyStore:       vacancystore.NewInstance(db.DB),
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
		applicantStore:     applicantstore.NewInstance(db.DB),
		tokenMap:           sync.Map{},
		cityMap:            map[string]string{},
		filesStorage:       filestorage.Instance,
		applicantHistory:   applicanthistoryhandler.Instance,
	}
}

type impl struct {
	client             hhclient.Provider
	extStore           extservicestore.Provider
	spaceUserStore     spaceusersstore.Provider
	vacancyStore       vacancystore.Provider
	spaceSettingsStore spacesettingsstore.Provider
	applicantStore     applicantstore.Provider
	tokenMap           sync.Map
	cityMap            map[string]string
	filesStorage       filestorage.Provider
	applicantHistory   applicanthistoryhandler.Provider
}

const (
	TokenCode     = "HH_TOKEN"
	VacancyUriTpl = "https://hh.ru/vacancy/%v"
)

func (i *impl) getLogger(spaceID, vacancyID string) *log.Entry {
	logger := log.WithField("integration", "HeadHunter")
	if spaceID != "" {
		logger = logger.WithField("space_id", spaceID)
	}
	if vacancyID != "" {
		logger = logger.WithField("vacancy_id", vacancyID)
	}
	return logger
}

func (i *impl) GetConnectUri(spaceID string) (uri string, err error) {
	clientID, err := i.getValue(spaceID, models.HhClientIDSetting)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения настройки ClientID для HH")
	}
	_, err = i.getValue(spaceID, models.HhClientSecretSetting)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения настройки ClientSecret для HH")
	}
	return i.client.GetLoginUri(clientID, spaceID)
}

func (i *impl) RequestToken(spaceID, code string) {
	logger := i.getLogger(spaceID, "")
	clientID, err := i.getValue(spaceID, models.HhClientIDSetting)
	if err != nil {
		logger.WithError(err).Error("ошибка получения настройки ClientID для HH")
		return
	}

	clientSecret, err := i.getValue(spaceID, models.HhClientSecretSetting)
	if err != nil {
		logger.WithError(err).Error("ошибка получения настройки ClientSecret для HH")
		return
	}

	req := hhapimodels.RequestToken{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Code:         code,
		RedirectUri:  "",
	}
	token, err := i.client.RequestToken(context.TODO(), req)
	if err != nil {
		logger.WithError(err).Error("ошибка получения токена HH")
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
	token := hhapimodels.ResponseToken{}
	err = json.Unmarshal(data, &token)
	if err != nil {
		return false
	}
	i.storeToken(spaceID, token, false)
	return true
}

func (i *impl) VacancyPublish(ctx context.Context, spaceID, vacancyID string) (hMsg string, err error) {
	accessToken, hMsg, err := i.getToken(ctx, spaceID)
	if err != nil || hMsg != "" {
		return hMsg, err
	}

	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return "", err
	}
	if rec == nil {
		return "вакансия не найдена", nil
	}

	if models.VacancyStatusOpened != rec.Status {
		return fmt.Sprintf("неподходящей статус вакансии %v, для публикации", rec.Status), nil
	}

	if rec.HhID != "" && rec.HhStatus != models.VacancyPubStatusNone && rec.HhStatus != models.VacancyPubStatusClosed {
		return "вакансия уже размещена", nil
	}

	request, hMsg := i.fillVacancyData(ctx, rec)
	if hMsg != "" {
		return hMsg, nil
	}

	id, err := i.client.VacancyPublish(ctx, accessToken, *request)
	if err != nil {
		return "", err
	}
	vacancyUrl := fmt.Sprintf(VacancyUriTpl, id)
	updMap := map[string]interface{}{
		"hh_id":     id,
		"hh_uri":    vacancyUrl,
		"hh_status": models.VacancyPubStatusModeration,
	}
	err = i.vacancyStore.Update(spaceID, vacancyID, updMap)
	if err != nil {
		return "", errors.Errorf("не удалось сохранить идентификатор опубликованной вакансии (%v)", id)
	}
	return "", nil
}

func (i *impl) VacancyUpdate(ctx context.Context, spaceID, vacancyID string) (hMsg string, err error) {
	accessToken, hMsg, err := i.getToken(ctx, spaceID)
	if err != nil || hMsg != "" {
		return hMsg, err
	}

	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return "", err
	}
	if rec == nil {
		return "вакансия не найдена", nil
	}
	if rec.HhID == "" {
		return "вакансия еще не опубликованна", nil
	}

	request, hMsg := i.fillVacancyData(ctx, rec)
	if hMsg != "" {
		return hMsg, nil
	}

	err = i.client.VacancyUpdate(ctx, accessToken, rec.HhID, *request)
	updMap := map[string]interface{}{
		"hh_status": models.VacancyPubStatusModeration,
	}
	err = i.vacancyStore.Update(spaceID, vacancyID, updMap)
	if err != nil {
		return "", errors.New("не удалось обновить статус публикации")
	}
	return "", nil
}

func (i *impl) VacancyClose(ctx context.Context, spaceID, vacancyID string) (hMsg string, err error) {
	accessToken, hMsg, err := i.getToken(ctx, spaceID)
	if err != nil || hMsg != "" {
		return hMsg, err
	}
	meResp, err := i.client.Me(ctx, accessToken)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения информации о токене HH")
	}
	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return "", err
	}
	if rec == nil {
		return "вакансия не найдена", nil
	}
	if rec.HhID == "" {
		return "вакансия еще не опубликованна", nil
	}
	return "", i.client.VacancyClose(ctx, accessToken, meResp.Employer.ID, rec.HhID)
}

func (i *impl) VacancyAttach(ctx context.Context, spaceID, vacancyID, hhID string) (hMsg string, err error) {
	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return "", err
	}
	if rec == nil {
		return "вакансия не найдена", nil
	}
	if rec.HhID != "" {
		return "ссылка на вакансию уже добавлена", nil
	}
	if models.VacancyStatusOpened != rec.Status {
		return fmt.Sprintf("неподходящей статус вакансии: %v", rec.Status), nil
	}
	accessToken, hMsg, err := i.getToken(ctx, spaceID)
	if err != nil || hMsg != "" {
		return hMsg, err
	}

	self, err := i.client.Me(ctx, accessToken)
	if err != nil {
		return "", err
	}
	info, err := i.client.GetVacancy(ctx, accessToken, hhID)
	if err != nil {
		return "", err
	}
	if info == nil {
		return "вакансия не найдена на сайте HeadHunter", nil
	}
	if info.Employer.ID != self.Employer.ID {
		return "вакансия принадлежит другой компании", nil
	}
	updMap := map[string]interface{}{
		"hh_id":     hhID,
		"hh_uri":    info.AlternateUrl,
		"hh_status": models.VacancyPubStatusPublished,
	}
	err = i.vacancyStore.Update(spaceID, vacancyID, updMap)
	if err != nil {
		return "", errors.Errorf("не удалось обновить данные опубликованной вакансии (%v)", hhID)
	}
	return "", nil
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
		Url:    rec.HhUri,
		Status: models.VacancyPubStatusNone,
	}
	if rec.HhID == "" {
		return &result, nil
	}
	result.Status = rec.HhStatus
	return &result, nil
}

func (i *impl) HandleNegotiations(ctx context.Context, data dbmodels.Vacancy) error {
	accessToken, hMsg, err := i.getToken(ctx, data.SpaceID)
	if err != nil {
		return err
	}
	if hMsg != "" {
		return errors.New(hMsg)
	}
	resp, err := i.client.Negotiations(ctx, accessToken, data.HhID, 0, 20)
	if err != nil {
		return err
	}
	for _, item := range resp.Items {
		logger := i.getLogger(data.SpaceID, data.ID)
		logger = logger.
			WithField("negotiation_id", item.ID).
			WithField("resume_id", item.Resume.ID)
		found, err := i.applicantStore.IsExistNegotiationID(data.SpaceID, item.ID, models.ApplicantSourceHh)
		if err != nil {
			logger.WithError(err).Error("не удалось проверить наличие отклика")
			continue
		}
		if found {
			continue
		}
		if item.Resume.ResumeUrl == "" {
			logger.Info("отклик без резюме")
			continue
		}
		resume, err := i.client.GetResume(ctx, accessToken, item.Resume.ResumeUrl)
		if err != nil {
			logger.WithError(err).Error("не удалось получить резюме по отклику")
			continue
		}
		applicantData := dbmodels.Applicant{
			BaseSpaceModel: dbmodels.BaseSpaceModel{
				SpaceID: data.SpaceID,
			},
			VacancyID:       data.ID,
			NegotiationID:   item.ID,
			ResumeID:        resume.ID,
			ExtApplicantID:  resume.Owner.ID,
			Source:          models.ApplicantSourceHh,
			NegotiationDate: time.Now(),
			Status:          models.ApplicantStatusNegotiation,
			FirstName:       resume.FirstName,
			LastName:        resume.LastName,
			MiddleName:      resume.MiddleName,
			ResumeTitle:     resume.Title,
			Salary:          resume.Salary.Amount,
			Address:         resume.Area.Name,
			Gender:          models.GenderType(resume.Gender.ID),
			Relocation:      resume.Relocation.GetRelocationType(),
			TotalExperience: resume.TotalExperience.Months, //опыт работ в месяцах
			Params: dbmodels.ApplicantParams{
				Education:               models.EducationType(resume.Education.Level.ID),
				HaveAdditionalEducation: len(resume.Education.Additional) > 0,
				Employments:             []models.Employment{},
				Schedules:               []models.Schedule{},
				Languages:               []dbmodels.Language{},
				TripReadiness:           models.TripReadinessType(resume.BusinessTripReadiness.ID),
				DriverLicenseTypes:      []models.DriverLicenseType{},
				SearchStatus:            models.SearchStatusType(resume.JobSearchStatusesEmployer.ID),
			},
		}
		for _, stage := range data.SelectionStages {
			if stage.Name == dbmodels.NegotiationStage {
				applicantData.SelectionStageID = stage.ID
				break
			}
		}

		for _, employment := range resume.Employments {
			applicantData.Params.Employments = append(applicantData.Params.Employments, models.Employment(employment.ID))
		}

		for _, schedule := range resume.Schedules {
			applicantData.Params.Schedules = append(applicantData.Params.Schedules, models.Schedule(schedule.ID))
		}
		for _, driverLicense := range resume.DriverLicenseTypes {
			applicantData.Params.DriverLicenseTypes = append(applicantData.Params.DriverLicenseTypes, models.DriverLicenseType(driverLicense.ID))
		}

		birthDate, err := resume.GetBirthDate()
		if err != nil {
			logger.WithError(err).Error("ошибка получения даты рождения кандидата")
		}
		applicantData.BirthDate = birthDate
		for _, contact := range resume.Contact {
			switch contact.Type.ID {
			case hhapimodels.ContactTypeCell:
				value, ok := contact.Value.(map[string]interface{})
				if !ok {
					logger.Error("ошибка получения мобильного телефона кандидата")
					continue
				}
				fValue, ok := value["formatted"].(string)
				if ok {
					applicantData.Phone = fValue
				}
			case hhapimodels.ContactTypeHome:
				if applicantData.Phone != "" {
					continue
				}
				value, ok := contact.Value.(map[string]interface{})
				if !ok {
					logger.Error("ошибка получения домашнего телефона кандидата")
					continue
				}
				fValue, ok := value["formatted"].(string)
				if ok {
					applicantData.Phone = fValue
				}

			case hhapimodels.ContactTypeEmail:
				value, ok := contact.Value.(string)
				if !ok {
					logger.Error("ошибка получения email кандидата")
					continue
				}
				applicantData.Email = value
			}
		}
		for _, area := range resume.Citizenship {
			applicantData.Citizenship = area.Name
			if applicantData.Citizenship != "" {
				break
			}
		}
		for _, language := range resume.Language {
			lng := dbmodels.Language{
				Name:          language.Name,
				LanguageLevel: models.LanguageLevelType(language.Level.ID),
			}
			applicantData.Params.Languages = append(applicantData.Params.Languages, lng)
		}
		applicantID, err := i.applicantStore.Create(applicantData)
		if err != nil {
			logger.WithError(err).Error("ошибка сохранения кандидата по отклику")
		}

		if resume.Actions.Pdf.Url != "" {
			err = i.downloadResumePdf(ctx, data.SpaceID, applicantID, resume.Actions.Pdf.Url)
			if err != nil {
				logger.
					WithField("resume_url", resume.Actions.Pdf.Url).
					WithError(err).
					Error("ошибка загрузки резюме из HH")
			}
		}
		changes := applicanthistoryhandler.GetCreateChanges("Кандидат добавлен с работного сайта на вакансию", applicantData)
		i.applicantHistory.Save(applicantData.SpaceID, applicantID, applicantData.VacancyID, "", dbmodels.HistoryTypeNegotiation, changes)
	}
	return nil
}

func (i *impl) getValue(spaceID string, code models.SpaceSettingCode) (string, error) {
	return i.spaceSettingsStore.GetValueByCode(spaceID, code)
}

func (i *impl) fillAreas(areas []hhapimodels.Area) {
	for _, area := range areas {
		if len(area.Areas) == 0 {
			i.cityMap[area.Name] = area.ID
			continue
		}
		i.fillAreas(area.Areas)
	}
}

func (i *impl) getArea(ctx context.Context, city *dbmodels.City) (hhapimodels.DictItem, error) {
	if len(i.cityMap) == 0 {
		areas, err := i.client.GetAreas(ctx)
		for _, area := range areas {
			if area.Name == "Россия" {
				i.fillAreas(area.Areas)
				break
			}
		}
		return hhapimodels.DictItem{}, err
	}
	id, ok := i.cityMap[city.City]
	if ok {
		return hhapimodels.DictItem{
			ID: id,
		}, nil
	}
	cityRegion := fmt.Sprintf("%v (%v", city.City, city.Region)
	for key, value := range i.cityMap {
		if strings.HasPrefix(key, cityRegion) {
			return hhapimodels.DictItem{
				ID: value,
			}, nil
		}
	}
	return hhapimodels.DictItem{}, errors.New("город публикации не найден в справочнике")
}

func (i *impl) storeToken(spaceID string, token hhapimodels.ResponseToken, inDb bool) error {
	tokenData := hhapimodels.TokenData{
		ResponseToken: token,
		ExpiresAt:     time.Now(),
	}
	i.tokenMap.Store(spaceID, tokenData)
	if inDb {
		data, err := json.Marshal(token)
		err = i.extStore.Set(spaceID, TokenCode, data)
		if err != nil {
			return errors.Wrap(err, "ошибка сохранения токена HH в бд")
		}
	}
	return nil
}

func (i *impl) getToken(ctx context.Context, spaceID string) (token, hMsg string, err error) {
	if !i.CheckConnected(spaceID) {
		return "", "HeadHunter не подключен", nil
	}
	value, ok := i.tokenMap.Load(spaceID)
	if !ok {
		return "", "HeadHunter не подключен", nil
	}
	tokenData := value.(hhapimodels.TokenData)
	if time.Now().After(tokenData.ExpiresAt) {
		req := hhapimodels.RefreshToken{
			RefreshToken: tokenData.RefreshToken,
		}
		tokenResp, err := i.client.RefreshToken(ctx, req)
		if err != nil {
			return "", "", errors.Wrap(err, "ошибка получения токена для HeadHunter")
		}
		err = i.storeToken(spaceID, *tokenResp, true)
		if err != nil {
			return "", "", errors.Wrap(err, "ошибка сохранения токена для HeadHunter")
		}
	}
	return tokenData.AccessToken, "", nil
}

func (i *impl) fillVacancyData(ctx context.Context, rec *dbmodels.Vacancy) (req *hhapimodels.VacancyPubRequest, hMsg string) {
	if rec.City == nil {
		return nil, "не указан город публикации"
	}
	area, err := i.getArea(ctx, rec.City)
	if err != nil {
		return nil, err.Error()
	}

	if rec.JobTitle == nil {
		return nil, "для публикации на HeadHunter, необходимо указать должность"
	}
	if len(rec.Requirements) < 200 {
		return nil, "для публикации на HeadHunter, необходимо указать описание не менее 200 символов"
	}
	request := hhapimodels.VacancyPubRequest{
		Description:       rec.Requirements,
		Name:              rec.VacancyName,
		Area:              area,
		ProfessionalRoles: []hhapimodels.DictItem{{ID: rec.JobTitle.HhRoleID}},
		BillingType: hhapimodels.DictItem{
			ID: "free",
		},
		Type: hhapimodels.DictItem{
			ID: "open",
		},
	}
	salary := hhapimodels.Salary{Currency: "RUR"}
	if rec.Salary.InHand != 0 {
		salary.From = rec.Salary.InHand
		salary.To = rec.Salary.InHand
		salary.Gross = false
		request.Salary = &salary
	} else if rec.Salary.From != 0 || rec.Salary.To != 0 {
		salary = hhapimodels.Salary{
			From:  rec.Salary.From,
			To:    rec.Salary.To,
			Gross: true,
		}
		request.Salary = &salary
	}
	if rec.Schedule != "" {
		request.Schedule = &hhapimodels.DictItem{ID: string(rec.Schedule)}
	}
	return &request, ""
}

func (i *impl) GetCheckList(ctx context.Context, spaceID string, status models.VacancyPubStatus) ([]dbmodels.Vacancy, error) {
	return i.vacancyStore.ListHhByStatus(spaceID, status)
}

func (i *impl) CheckIsModerationDone(ctx context.Context, spaceID string, list []dbmodels.Vacancy) error {
	return i.checkPublications(ctx, spaceID, list)
}

func (i *impl) CheckIsActivePublications(ctx context.Context, spaceID string, list []dbmodels.Vacancy) error {
	return i.checkPublications(ctx, spaceID, list)
}

func (i *impl) checkPublications(ctx context.Context, spaceID string, list []dbmodels.Vacancy) error {
	logger := i.getLogger(spaceID, "")
	accessToken, hMsg, err := i.getToken(ctx, spaceID)
	if err != nil {
		return err
	}
	if hMsg != "" {
		return errors.New(hMsg)
	}
	for _, rec := range list {
		if helpers.IsContextDone(ctx) {
			return nil
		}
		info, err := i.client.GetVacancy(ctx, accessToken, rec.HhID)
		if err != nil {
			logger.
				WithError(err).
				Error("не удалось проверить статус публикации")
			continue
		}
		if info == nil {
			continue
		}
		newStatus := info.GetPubStatus()
		if newStatus == rec.AvitoStatus {
			continue
		}
		updMap := map[string]interface{}{
			"hh_status": newStatus,
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

func (i *impl) downloadResumePdf(ctx context.Context, spaceID, applicantID, resumeUrl string) error {
	accessToken, hMsg, err := i.getToken(ctx, spaceID)
	if err != nil {
		return err
	}
	if hMsg != "" {
		return errors.New(hMsg)
	}
	body, err := i.client.DownloadResume(ctx, accessToken, resumeUrl)
	if err != nil {
		return err
	}
	return i.filesStorage.Upload(ctx, spaceID, applicantID, body, "resume.pdf", dbmodels.ApplicantResume)
}
