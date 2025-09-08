package vk

import (
	"encoding/json"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	companystore "hr-tools-backend/lib/dicts/company/store"
	negotiationchathandler "hr-tools-backend/lib/external-services/negotiation-chat"
	gpthandler "hr-tools-backend/lib/gpt"
	messagetemplate "hr-tools-backend/lib/message-template"
	"hr-tools-backend/lib/smtp"
	applicantsurveystore "hr-tools-backend/lib/survey/applicant-survey-store"
	vacancysurveystore "hr-tools-backend/lib/survey/vacancy-survey-store"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"
	surveyapimodels "hr-tools-backend/models/api/survey"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	RunStep0(spaceID, vacancyID, applicantID string) (ok bool, err error)
	GetStep0Survey(id string) (*surveyapimodels.ApplicantVkStep0SurveyView, error) // анкета для фронта
	// HandleStep0Survey // ответы от фронта, сохранение в бд, анализ проходит или нет
	RunStep1(spaceID, vacancyID, applicantID string) (ok bool, err error)
}

var Instance Provider

const (
	defaultCompanyName = "HR-Tools"
)

func NewHandler() {
	Instance = impl{
		vSurveyStore:           vacancysurveystore.NewInstance(db.DB),
		vacancyStore:           vacancystore.NewInstance(db.DB),
		applicantStore:         applicantstore.NewInstance(db.DB),
		vkStore:                applicantvkstore.NewInstance(db.DB),
		negotiationChatHandler: negotiationchathandler.Instance,
		companyStore:           companystore.NewInstance(db.DB),
		messageTemplate:        messagetemplate.Instance,
	}
}

type impl struct {
	vSurveyStore           vacancysurveystore.Provider
	vacancyStore           vacancystore.Provider
	applicantStore         applicantstore.Provider
	aSurveyStore           applicantsurveystore.Provider
	vkStore                applicantvkstore.Provider
	negotiationChatHandler negotiationchathandler.Provider
	companyStore           companystore.Provider
	messageTemplate        messagetemplate.Provider
}

func (i impl) getLogger(spaceID, applicantID string) *logrus.Entry {
	return log.
		WithField("space_id", spaceID).
		WithField("applicant_id", applicantID)
}

func (i impl) RunStep0(spaceID, vacancyID, applicantID string) (ok bool, err error) {
	applicantRec, err := i.applicantStore.GetByID(spaceID, applicantID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения данных кандидата")
	}
	if applicantRec == nil {
		return false, nil
	}
	if applicantRec.ApplicantSurvey == nil {
		return false, nil
	}
	rec, err := i.vkStore.GetByApplicantID(spaceID, applicantID)
	if err != nil {
		return false, err
	}
	if rec != nil {
		if rec.Status != dbmodels.VkStep0NotSent {
			return false, errors.Wrap(err, "вопросы уже отправлены кандидату")
		}
	} else {
		rec = &dbmodels.ApplicantVkStep{
			BaseSpaceModel: dbmodels.BaseSpaceModel{SpaceID: spaceID},
			ApplicantID:    applicantID,
			Status:         dbmodels.VkStep0NotSent,
			Step0: dbmodels.VkStep0{
				Answers: []dbmodels.VkStep0{},
			},
		}
		id, err := i.vkStore.Save(*rec)
		if err != nil {
			return false, errors.Wrap(err, "ошибка сохранения данных по опросу в бд")
		}
		rec.ID = id
	}

	// отправка ссылки на анкету
	logger := i.getLogger(spaceID, applicantID)
	isChatAvailable := false
	availability, err := i.negotiationChatHandler.IsVailable(spaceID, applicantID)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка проверки доступности чата с кандидатом")
	} else {
		isChatAvailable = availability.IsAvailable
	}

	companyName := i.getCompanyName(spaceID, applicantRec.Vacancy.CompanyID)
	link := config.Conf.UIParams.SurveyStep0Path + rec.ID
	isSend := false
	if isChatAvailable {
		if i.sendToChat(spaceID, applicantID, companyName, link, logger) {
			isSend = true
		}
	}
	if applicantRec.Email != "" {
		emailFrom, err := i.messageTemplate.GetSenderEmail(spaceID)
		if err != nil {
			logger.
				WithError(err).
				Warn("ошибка получения почты компании для отправки сообщения с ссылкой на анкету на email кандидату")
		} else if emailFrom != "" {
			if i.sendToEmail(emailFrom, applicantRec.Email, companyName, link, logger) {
				isSend = true
			}
		}
	}
	if isSend {
		rec = &dbmodels.ApplicantVkStep{
			BaseSpaceModel: dbmodels.BaseSpaceModel{SpaceID: spaceID},
			ApplicantID:    applicantID,
			Status:         dbmodels.VkStep0Sent,
			Step0: dbmodels.VkStep0{
				Answers: []dbmodels.VkStep0{},
			},
		}
		_, err = i.vkStore.Save(*rec)
		if err != nil {
			return false, errors.Wrap(err, "ошибка сохранения данных по опросу в бд")
		}
		return true, nil
	}
	return false, nil
}

func (i impl) GetStep0Survey(id string) (*surveyapimodels.ApplicantVkStep0SurveyView, error) {
	rec, err := i.vkStore.GetByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения анкеты кандидата")
	}
	if rec == nil {
		return nil, errors.New("анкета не найдена")
	}
	result := surveyapimodels.ApplicantVkStep0SurveyView{
		Questions: []surveyapimodels.ApplicantVkStep0Question{
			{
				QuestionID:   "1",
				QuestionText: TypicalQuestion1,
			},
			// TODO Добавить типовые вопросы
		},
	}
	return &result, nil
}

// HandleStep0Survey

func (i impl) RunStep1(spaceID, vacancyID, applicantID string) (ok bool, err error) {
	// applicantSurveyRec dbmodels.ApplicantSurvey
	applicantRec, err := i.applicantStore.GetByID(spaceID, applicantID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения данных кандидата")
	}
	if applicantRec == nil {
		return false, nil
	}
	if applicantRec.ApplicantSurvey == nil {
		return false, nil
	}

	vacancy, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения вакансии")
	}
	if vacancy == nil || vacancy.HRSurvey == nil {
		return false, nil
	}

	vacancyInfo, err := surveyapimodels.GetVacancyDataContent(*vacancy)
	if err != nil {
		return false, err
	}

	applicantInfo, err := surveyapimodels.GetApplicantDataContent(applicantRec.Applicant)
	if err != nil {
		return false, err
	}

	hrSurvey, err := surveyapimodels.GetHRDataContent(*vacancy.HRSurvey)
	if err != nil {
		return false, err
	}
	applicantAnswers, err := surveyapimodels.GetApplicantAnswersContent(*applicantRec.ApplicantSurvey)
	resp, err := gpthandler.Instance.VkStep1(vacancy.SpaceID, vacancy.ID, vacancyInfo, applicantInfo, hrSurvey, applicantAnswers)
	if err != nil {
		return false, errors.Wrap(err, "ошибка вызова ИИ при генерации черновика скрипта")
	}
	vkStep1 := dbmodels.VkStep1{}
	err = json.Unmarshal([]byte(resp.Description), &vkStep1)
	if err != nil {
		return false, errors.Wrapf(err, "ошибка декодирования json в структуру оценки кандидата, json: %v", resp.Description)
	}

	rec := dbmodels.ApplicantVkStep{
		BaseSpaceModel: dbmodels.BaseSpaceModel{SpaceID: spaceID},
		ApplicantID:    applicantID,
		VkStep1:        vkStep1,

		// BaseSpaceModel:  dbmodels.BaseSpaceModel{SpaceID: spaceID},
		// VacancySurveyID: vacancyRec.HRSurvey.ID,
		// ApplicantID:     applicantID,
		// Survey:          dbmodels.ApplicantSurveyQuestions{Questions: surveyData.Questions},
		// IsFilledOut:     false,
		// HrThreshold:     vacancyRec.HRSurvey.Survey.GetThreshold(),
	}
	_, err = i.vkStore.Save(rec)
	if err != nil {
		return false, errors.Wrap(err, "ошибка сохранения черновика скрипта")
	}
	return true, nil
}

func (i impl) sendToChat(spaceID, applicantID, companyName, link string, logger *log.Entry) bool {
	text, err := messagetemplate.GetSurvayStep0SuggestMessage(companyName, link, false)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка получения сообщения с ссылкой на анкету для отправки кандидату через чат")
		return false
	}
	req := negotiationapimodels.NewMessageRequest{
		ApplicantID: applicantID,
		Text:        text,
	}
	err = i.negotiationChatHandler.SendMessage(spaceID, req)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка отправки сообщения с ссылкой на анкету в чат с кандидатом")
		return false
	}
	return true
}

func (i impl) getCompanyName(spaceID string, companyID *string) string {
	if companyID != nil {
		company, err := i.companyStore.GetByID(spaceID, *companyID)
		if err == nil && company != nil && company.Name != "" {
			return company.Name
		}
	}
	return defaultCompanyName
}

func (i impl) sendToEmail(emailFrom, mailTo, companyName, link string, logger *log.Entry) bool {
	text, err := messagetemplate.GetSurvaySuggestMessage(companyName, link, true)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка получения сообщения с ссылкой на анкету для отправки кандидату через email")
		return false
	}
	title := messagetemplate.GetSurvaySuggestTitle()
	err = smtp.Instance.SendHtmlEMail(emailFrom, mailTo, text, title, nil)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка отправки сообщения с ссылкой на анкету на email кандидату")
		return false
	}
	return true
}
