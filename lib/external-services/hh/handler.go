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
	pushhandler "hr-tools-backend/lib/space/push/handler"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/lib/utils/lock"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	"hr-tools-backend/models"
	hhapimodels "hr-tools-backend/models/api/hh"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"
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
	TokenCode       = "HH_TOKEN"
	VacancyUriTpl   = "https://hh.ru/vacancy/%v"
	NotConnectedMsg = "HeadHunter не подключен"
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
	_, err = i.storeToken(spaceID, *token, time.Now(), true)
	if err != nil {
		logger.Error(err)
	}
}

func (i *impl) CheckConnected(ctx context.Context, spaceID string) bool {
	logger := i.getLogger(spaceID, "")
	_, _, hMsg, err := i.getToken(ctx, spaceID)
	if err != nil {
		logger.WithError(err).Error("ошибка получения токена hh")
		return false
	}
	if hMsg == NotConnectedMsg {
		logger.Info(NotConnectedMsg)
		return false
	}
	if hMsg != "" {
		logger.Errorf("Ошибка получения токена: %v", hMsg)
		return false
	}
	return true
}

func (i *impl) VacancyPublish(ctx context.Context, spaceID, vacancyID string) (hMsg string, err error) {
	_, accessToken, hMsg, err := i.getToken(ctx, spaceID)
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

	id, hMsg, err := i.client.VacancyPublish(ctx, accessToken, *request)
	if err != nil || hMsg != "" {
		updMap := map[string]interface{}{
			"hh_reasons": hMsg,
			"hh_status":  models.VacancyPubStatusNone,
		}
		e := i.vacancyStore.Update(spaceID, vacancyID, updMap)
		if e != nil {
			i.getLogger(spaceID, vacancyID).WithError(e).Error("не удалось сохранить причину размещения вакансии")
		}
		return hMsg, err
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
	_, accessToken, hMsg, err := i.getToken(ctx, spaceID)
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
	self, accessToken, hMsg, err := i.getToken(ctx, spaceID)
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
	return "", i.client.VacancyClose(ctx, accessToken, self.Employer.ID, rec.HhID)
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
	self, accessToken, hMsg, err := i.getToken(ctx, spaceID)
	if err != nil || hMsg != "" {
		return hMsg, err
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
		Reason: rec.HhReasons,
	}
	if rec.HhID == "" {
		return &result, nil
	}
	result.Status = rec.HhStatus
	return &result, nil
}

func (i *impl) HandleNegotiations(ctx context.Context, data dbmodels.Vacancy) error {
	_, accessToken, hMsg, err := i.getToken(ctx, data.SpaceID)
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

		notification := models.GetPushApplicantNegotiation(data.VacancyName, applicantData.GetFIO())
		go i.sendNotification(data, notification)
	}
	return nil
}

func (i *impl) SendMessage(ctx context.Context, data dbmodels.Applicant, msg string) error {
	_, accessToken, hMsg, err := i.getToken(ctx, data.SpaceID)
	if err != nil {
		return err
	}
	if hMsg != "" {
		return errors.New(hMsg)
	}
	return i.client.SendNewMessage(ctx, accessToken, data.VacancyID, data.NegotiationID, msg)
}

func (i *impl) GetMessages(ctx context.Context, user dbmodels.SpaceUser, data dbmodels.Applicant) ([]negotiationapimodels.MessageItem, error) {
	_, accessToken, hMsg, err := i.getToken(ctx, data.SpaceID)
	if err != nil {
		return nil, err
	}
	if hMsg != "" {
		return nil, errors.New(hMsg)
	}
	resp, err := i.client.GetMessages(ctx, accessToken, data.VacancyID, data.NegotiationID)
	if resp.Found == 0 {
		return []negotiationapimodels.MessageItem{}, nil
	}
	result := make([]negotiationapimodels.MessageItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		msg := negotiationapimodels.MessageItem{
			ID:          item.ID,
			SelfMessage: item.Author.ParticipantType == hhapimodels.ParticipantEmployer,
			Text:        item.Text,
		}
		if msg.SelfMessage {
			msg.AuthorFullName = user.GetFullName()
		} else {
			msg.AuthorFullName = data.GetFIO()
		}
		createdAt, err := helpers.ParseHhTime(item.CreatedAt)
		if err != nil {
			log.WithError(err).Warn("ошибка преобразования даты сообщения HeadHunter")
		} else {
			msg.MessageDateTime = createdAt
		}
		result = append(result, msg)
	}
	return result, nil
}

func (i *impl) GetLastInMessage(ctx context.Context, data dbmodels.Applicant) (*negotiationapimodels.MessageItem, error) {
	_, accessToken, hMsg, err := i.getToken(ctx, data.SpaceID)
	if err != nil {
		return nil, err
	}
	if hMsg != "" {
		return nil, errors.New(hMsg)
	}
	resp, err := i.client.GetMessages(ctx, accessToken, data.VacancyID, data.NegotiationID)
	if resp.Found == 0 || len(resp.Items) == 0 {
		return nil, nil
	}
	item := resp.Items[len(resp.Items)-1]
	msg := negotiationapimodels.MessageItem{
		ID:          item.ID,
		SelfMessage: item.Author.ParticipantType == hhapimodels.ParticipantEmployer,
		Text:        item.Text,
	}
	if !msg.SelfMessage {
		msg.AuthorFullName = data.GetFIO()
	}
	createdAt, err := helpers.ParseHhTime(item.CreatedAt)
	if err != nil {
		log.WithError(err).Warn("ошибка преобразования даты сообщения HeadHunter")
	} else {
		msg.MessageDateTime = createdAt
	}

	return &msg, nil
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
		if err != nil {
			return hhapimodels.DictItem{}, err
		}
		for _, area := range areas {
			if area.Name == "Россия" {
				i.fillAreas(area.Areas)
				break
			}
		}
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

func (i *impl) storeToken(spaceID string, token hhapimodels.ResponseToken, requestTime time.Time, inDb bool) (hhapimodels.TokenData, error) {
	tokenData := hhapimodels.TokenData{
		ResponseToken: token,
		ExpiresAt:     requestTime.Add(time.Duration(token.ExpiresIN) * time.Second),
	}
	i.tokenMap.Store(spaceID, tokenData)
	if inDb {
		data, err := json.Marshal(token)
		err = i.extStore.Set(spaceID, TokenCode, data)
		if err != nil {
			return tokenData, errors.Wrap(err, "ошибка сохранения токена HH в бд")
		}
	}
	return tokenData, nil
}

func (i *impl) getToken(ctx context.Context, spaceID string) (self *hhapimodels.MeResponse, accessToken, hMsg string, err error) {

	getTokenSafeFunc := func() error {
		tokenData, ok, err := i.getTokenFromStorage(spaceID)
		if err != nil {
			return err
		}
		if !ok {
			hMsg = NotConnectedMsg
			return nil
		}
		if !tokenData.IsExpired() {
			//Проверяем токен
			accessToken = tokenData.AccessToken
			//
			isExpired := false
			self, isExpired, err = i.client.Me(ctx, accessToken)
			if err != nil {
				return err
			}
			if !isExpired {
				// токен есть и он валиден
				return nil
			}
		}
		body, _ := json.Marshal(tokenData)
		logger := i.getLogger(spaceID, "").
			WithField("token_data", string(body))
		logger.Info("время жизни токена HH истекло, запрашиваем новый")

		tokenResp, isDeactivated, err := i.refresh(ctx, tokenData.RefreshToken)
		if err != nil {
			return err
		}
		if isDeactivated {
			logger.Info("Refresh-токен более не действителен, необходима повторная авторизация")
			i.removeToken(spaceID)
			hMsg = "Необходима повторная авторизация на HeadHunter"
			return nil
		}

		tokenData, err = i.storeToken(spaceID, tokenResp, time.Now(), true)
		if err != nil {
			tokenRespBody, _ := json.Marshal(tokenResp)
			logger.
				WithError(err).
				WithField("refresh_token_data", string(tokenRespBody)).
				Error("ошибка сохранения обновленного токена HH")
			return errors.Wrap(err, "ошибка сохранения обновленного токена")
		}
		accessToken = tokenResp.AccessToken

		//Проверяем новый токен
		isExpired := false
		self, isExpired, err = i.client.Me(ctx, accessToken)
		if err != nil {
			return err
		}
		if isExpired {
			hMsg = "действие с HeadHunter временно не доступно, повторите попытку позже"
			return nil
		}
		return nil
	}
	// выполняем с блокировкой по spaceID
	ok, err := lock.WithDelay(ctx, spaceID, 10*time.Second, getTokenSafeFunc)
	if !ok {
		return nil, "", "ошибка получения токена HeadHunter, операция временно невозможна", nil
	}
	return self, accessToken, hMsg, err
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
		AllowMessages:     true,
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
	_, accessToken, hMsg, err := i.getToken(ctx, spaceID)
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
		if newStatus == rec.HhStatus {
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
		if newStatus == models.VacancyPubStatusPublished {
			notification := models.GetPushVacancyPublished(rec.VacancyName, "HeadHunter")
			go i.sendNotification(rec, notification)
		}
	}
	return nil
}

func (i *impl) downloadResumePdf(ctx context.Context, spaceID, applicantID, resumeUrl string) error {
	_, accessToken, hMsg, err := i.getToken(ctx, spaceID)
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
	return i.filesStorage.Upload(ctx, spaceID, applicantID, body, "resume.pdf", dbmodels.ApplicantResume, "application/pdf")
}

func (i *impl) sendNotification(rec dbmodels.Vacancy, data models.NotificationData) {
	//отправляем автору
	pushhandler.Instance.SendNotification(rec.AuthorID, data)
	for _, teamMember := range rec.VacancyTeam {
		//отправляем команде
		if rec.AuthorID == teamMember.ID {
			continue
		}
		pushhandler.Instance.SendNotification(teamMember.ID, data)
	}
}

func (i *impl) refresh(ctx context.Context, refreshToken string) (resp hhapimodels.ResponseToken, isDeactivated bool, err error) {
	req := hhapimodels.RefreshToken{
		RefreshToken: refreshToken,
	}
	tokenResp, isDeactivated, err := i.client.RefreshToken(ctx, req)
	if err != nil {
		return hhapimodels.ResponseToken{}, false, errors.Wrap(err, "ошибка получения токена для HeadHunter")
	}
	if isDeactivated {
		return hhapimodels.ResponseToken{}, true, nil
	}
	return *tokenResp, false, nil
}

func (i *impl) getTokenFromStorage(spaceID string) (hhapimodels.TokenData, bool, error) {
	logger := i.getLogger(spaceID, "")
	value, ok := i.tokenMap.Load(spaceID)
	if ok {
		tokenData := value.(hhapimodels.TokenData)
		return tokenData, true, nil
	}
	rec, err := i.extStore.GetRec(spaceID, TokenCode)
	if err != nil {
		return hhapimodels.TokenData{}, false, errors.Wrap(err, "ошибка загрузки токена из бд")
	}
	if rec == nil {
		return hhapimodels.TokenData{}, false, nil
	}
	data := rec.Value
	token := hhapimodels.ResponseToken{}
	err = json.Unmarshal(data, &token)
	if err != nil {
		return hhapimodels.TokenData{}, false, errors.Wrap(err, "ошибка сериализации токена из бд")
	}
	tokenData, err := i.storeToken(spaceID, token, rec.UpdatedAt, false)
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения токена")
	}
	return tokenData, true, nil
}

func (i *impl) removeToken(spaceID string) error {
	i.tokenMap.Delete(spaceID)
	return i.extStore.DeleteRec(spaceID, TokenCode)
}
