package vk

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	"hr-tools-backend/lib/applicant"
	applicantstore "hr-tools-backend/lib/applicant/store"
	companystore "hr-tools-backend/lib/dicts/company/store"
	negotiationchathandler "hr-tools-backend/lib/external-services/negotiation-chat"
	gpthandler "hr-tools-backend/lib/gpt"
	messagetemplate "hr-tools-backend/lib/message-template"
	"hr-tools-backend/lib/smtp"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	"hr-tools-backend/models"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"
	surveyapimodels "hr-tools-backend/models/api/survey"
	dbmodels "hr-tools-backend/models/db"
	"sort"
)

type Provider interface {
	RunStep0(applicant dbmodels.Applicant) (ok bool, err error)
	GetSurveyStep0(id string) (*surveyapimodels.VkStep0SurveyView, error)                                                              // анкета для фронта
	HandleSurveyStep0(id string, answers surveyapimodels.VkStep0SurveyAnswers) (result surveyapimodels.VkStep0SurveyResult, err error) // ответы от фронта, сохранение в бд, анализ проходит или нет
	RunStep1(applicant dbmodels.Applicant) (ok bool, err error)
	UpdateStep1(spaceID, applicantID string, stepData surveyapimodels.VkStep1Update) (hMsg string, err error)
	RegenStep1(spaceID, applicantID string, stepData surveyapimodels.VkStep1Regen) (hMsg string, err error)
	RunRegenStep1(applicant dbmodels.Applicant) (ok bool, err error)
}

var Instance Provider

const (
	defaultCompanyName = "HR-Tools"
	Step0SucessMsg     = "Ваша анкета успешно сформирована, с вами свяжутся для информирования о результатах"  //TODO нужен текст ответа
	Step0FailMsg       = "Ваша анкета успешно сформирована, с вами свяжутся для информирования о результатах." //TODO нужен текст ответа
)

func NewHandler() {
	Instance = impl{
		vacancyStore:           vacancystore.NewInstance(db.DB),
		applicantStore:         applicantstore.NewInstance(db.DB),
		vkStore:                applicantvkstore.NewInstance(db.DB),
		negotiationChatHandler: negotiationchathandler.Instance,
		companyStore:           companystore.NewInstance(db.DB),
		messageTemplate:        messagetemplate.Instance,
		vkAiProvider:           gpthandler.GetHandler(false),
	}
}

type impl struct {
	vacancyStore           vacancystore.Provider
	applicantStore         applicantstore.Provider
	vkStore                applicantvkstore.Provider
	negotiationChatHandler negotiationchathandler.Provider
	companyStore           companystore.Provider
	messageTemplate        messagetemplate.Provider
	vkAiProvider           surveyapimodels.VkAiProvider // при необходимости поменяем пакет имплементации, пока через настройку config.Conf.AI.VkStep1AI
}

func (i impl) getLogger(spaceID, applicantID string) *logrus.Entry {
	return log.
		WithField("space_id", spaceID).
		WithField("applicant_id", applicantID)
}

func (i impl) RunStep0(applicantRec dbmodels.Applicant) (ok bool, err error) {
	rec, err := i.vkStore.GetByApplicantID(applicantRec.SpaceID, applicantRec.ID)
	if err != nil {
		return false, err
	}
	if rec != nil {
		if rec.Status != dbmodels.VkStep0NotSent {
			return false, errors.Wrap(err, "вопросы уже отправлены кандидату")
		}
	} else {
		rec = &dbmodels.ApplicantVkStep{
			BaseSpaceModel: dbmodels.BaseSpaceModel{SpaceID: applicantRec.SpaceID},
			ApplicantID:    applicantRec.ID,
			Status:         dbmodels.VkStep0NotSent,
			Step0: dbmodels.VkStep0{
				Answers: []dbmodels.VkStep0Answer{},
			},
		}
		id, err := i.vkStore.Save(*rec)
		if err != nil {
			return false, errors.Wrap(err, "ошибка сохранения данных по опросу в бд")
		}
		rec.ID = id
	}

	// отправка ссылки на анкету
	logger := i.getLogger(applicantRec.SpaceID, applicantRec.ID)
	isChatAvailable := false
	availability, err := i.negotiationChatHandler.IsVailable(applicantRec.SpaceID, applicantRec.ID)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка проверки доступности чата с кандидатом")
	} else {
		isChatAvailable = availability.IsAvailable
	}

	companyName := i.getCompanyName(applicantRec.SpaceID, applicantRec.Vacancy.CompanyID)
	link := config.Conf.UIParams.SurveyStep0Path + rec.ID
	isSend := false
	if isChatAvailable {
		if i.sendToChat(applicantRec.SpaceID, applicantRec.ID, companyName, link, logger) {
			isSend = true
		}
	}
	if applicantRec.Email != "" {
		emailFrom, err := i.messageTemplate.GetSenderEmail(applicantRec.SpaceID)
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
			BaseSpaceModel: dbmodels.BaseSpaceModel{SpaceID: applicantRec.SpaceID},
			ApplicantID:    applicantRec.ID,
			Status:         dbmodels.VkStep0Sent,
			Step0: dbmodels.VkStep0{
				Answers: []dbmodels.VkStep0Answer{},
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

func (i impl) GetSurveyStep0(id string) (*surveyapimodels.VkStep0SurveyView, error) {
	rec, err := i.vkStore.GetByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения анкеты кандидата")
	}
	if rec == nil {
		return nil, errors.New("анкета не найдена")
	}
	result := surveyapimodels.QuestionsStep0
	return &result, nil
}

func (i impl) HandleSurveyStep0(id string, request surveyapimodels.VkStep0SurveyAnswers) (result surveyapimodels.VkStep0SurveyResult, err error) {
	rec, err := i.vkStore.GetByID(id)
	if err != nil {
		return result, errors.Wrap(err, "ошибка получения анкеты кандидата")
	}
	if rec == nil {
		return result, errors.New("анкета не найдена")
	}
	rec.Step0 = dbmodels.VkStep0{
		Answers: []dbmodels.VkStep0Answer{},
	}
	for _, answer := range request.Answers {
		rec.Step0.Answers = append(rec.Step0.Answers,
			dbmodels.VkStep0Answer{
				ID:     answer.QuestionID,
				Answer: answer.Answer,
			})
	}
	rec.Status = dbmodels.VkStep0Answered
	_, err = i.vkStore.Save(*rec)
	if err != nil {
		return result, errors.Wrap(err, "ошибка сохранения анкеты")
	}

	isSucess := false
	//TODO принятие решения о прохождении
	//Если кандидат подходит, то переходить к шагу 1
	// ---- удалить когда будет алгоритм принятия
	if len(rec.Step0.Answers) > 4 || rec.Step0.Answers[2].Answer == "да" {
		isSucess = true
	}
	// ----
	if isSucess {
		rec.Status = dbmodels.VkStep0Done
	} else {
		rec.Status = dbmodels.VkStep0Refuse
	}
	_, err = i.vkStore.Save(*rec)
	if err != nil {
		return result, errors.Wrap(err, "ошибка сохранения анкеты")
	}
	if isSucess {
		result = surveyapimodels.VkStep0SurveyResult{
			Success: true,
			Message: Step0SucessMsg,
		}
		hMsg, err := applicant.Instance.UpdateStatus(rec.SpaceID, rec.ApplicantID, "", models.NegotiationStatusAccepted)
		if err != nil {
			i.getLogger(rec.SpaceID, rec.ApplicantID).
				WithError(err).
				Error("ВК. Шаг 0. Ошибка обновления статуса кандидата после успешного прохождения опроса")
		}
		if hMsg != "" {
			i.getLogger(rec.SpaceID, rec.ApplicantID).
				WithField("h_msg", hMsg).
				Error("ВК. Шаг 0. Ошибка обновления статуса кандидата после успешного прохождения опроса")
		}
		return result, nil
	}
	result = surveyapimodels.VkStep0SurveyResult{
		Success: false,
		Message: Step0FailMsg,
	}
	hMsg, err := applicant.Instance.UpdateStatus(rec.SpaceID, rec.ApplicantID, "", models.NegotiationStatusRejected)
	if err != nil {
		i.getLogger(rec.SpaceID, rec.ApplicantID).
			WithError(err).
			Error("ВК. Шаг 0. Ошибка обновления статуса кандидата после провального прохождения опроса")
	}
	if hMsg != "" {
		i.getLogger(rec.SpaceID, rec.ApplicantID).
			WithField("h_msg", hMsg).
			Error("ВК. Шаг 0. Ошибка обновления статуса кандидата после провального прохождения опроса")
	}
	return result, nil
}

func (i impl) RunStep1(applicant dbmodels.Applicant) (ok bool, err error) {
	vacancy, err := i.vacancyStore.GetByID(applicant.SpaceID, applicant.VacancyID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения вакансии")
	}
	if vacancy == nil {
		return false, nil
	}
	rec, err := i.vkStore.GetByID(applicant.ApplicantVkStep.ID)
	if err != nil {
		return false, err
	}
	vacancyInfo, applicantInfo, questions, applicantAnswers, _, err := i.getStep1Data(applicant, *vacancy, nil)
	if err != nil {
		return false, err
	}
	// запуск ИИ
	resp, err := i.vkAiProvider.VkStep1(vacancy.SpaceID, vacancy.ID, vacancyInfo, applicantInfo, questions, applicantAnswers)
	if err != nil {
		i.step1Fail(applicant, *rec)
		return false, errors.Wrap(err, "ошибка вызова ИИ при генерации черновика скрипта")
	}

	rec.Step1 = dbmodels.VkStep1{
		Questions:   []dbmodels.VkStep1Question{},
		ScriptIntro: resp.ScriptIntro,
		ScriptOutro: resp.ScriptOutro,
		Comments:    map[string]string{},
	}
	for k, q := range resp.Questions {
		qID := fmt.Sprintf("q%v", k+1)
		rec.Step1.Questions = append(rec.Step1.Questions, dbmodels.VkStep1Question{
			ID:                qID,
			Text:              q.Text,
			Order:             k,
			NotSuitable:       false,
			NotSuitableReason: "",
		})
		rec.Step1.Comments[qID] = resp.Comments[q.ID]
	}
	rec.Status = dbmodels.VkStep1Draft
	_, err = i.vkStore.Save(*rec)
	if err != nil {
		return false, errors.Wrap(err, "ошибка сохранения черновика скрипта")
	}
	return true, nil
}

func (i impl) UpdateStep1(spaceID, applicantID string, stepData surveyapimodels.VkStep1Update) (hMsg string, err error) {
	rec, err := i.vkStore.GetByApplicantID(spaceID, applicantID)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения анкеты кандидата")
	}
	if rec == nil {
		return "анкета не найдена", nil
	}
	if rec.Status != dbmodels.VkStep1Draft &&
		rec.Status != dbmodels.VkStep1DraftFail {
		return "невозможно отредактировать анкету", nil
	}
	rec.Step1.ScriptIntro = stepData.ScriptIntro
	rec.Step1.ScriptOutro = stepData.ScriptOutro
	rec.Step1.Questions = []dbmodels.VkStep1Question{}
	sort.Slice(stepData.Questions, func(k, j int) bool {
		return stepData.Questions[k].Order < stepData.Questions[j].Order
	})

	for k, question := range stepData.Questions {
		rec.Step1.Questions = append(rec.Step1.Questions, dbmodels.VkStep1Question{
			ID:                question.ID,
			Text:              question.Text,
			Order:             k,
			NotSuitable:       false,
			NotSuitableReason: "",
		})
	}
	rec.Step1.Comments = stepData.Comments
	if stepData.Approve {
		rec.Status = dbmodels.VkStep1Approved
	}
	_, err = i.vkStore.Save(*rec)
	if err != nil {
		return "", errors.Wrap(err, "ошибка сохранения черновика скрипта")
	}
	return "", nil
}

func (i impl) RegenStep1(spaceID, applicantID string, stepData surveyapimodels.VkStep1Regen) (hMsg string, err error) {
	applicant, err := i.applicantStore.GetByID(spaceID, applicantID)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения данных кандидата")
	}
	if applicant == nil {
		return "кандидат не найден", nil
	}
	rec, err := i.vkStore.GetByID(applicant.ApplicantVkStep.ID)
	if err != nil {
		return "", err
	}
	if rec == nil {
		return "данные по анкете не найдены", nil
	}
	if rec.Status != dbmodels.VkStep1Draft &&
		rec.Status != dbmodels.VkStep1DraftFail {
		return "невозможно отправить анкету на перегенерацию", nil
	}
	rec.Step1.Questions = []dbmodels.VkStep1Question{}
	for k, question := range stepData.Questions {
		rec.Step1.Questions = append(rec.Step1.Questions, dbmodels.VkStep1Question{
			ID:                question.ID,
			Text:              question.Text,
			Order:             k,
			NotSuitable:       question.NotSuitable,
			NotSuitableReason: question.NotSuitableReason,
		})
	}
	rec.Status = dbmodels.VkStep1Regen
	_, err = i.vkStore.Save(*rec)
	if err != nil {
		return "", errors.Wrap(err, "ошибка сохранения черновика скрипта")
	}
	return "", nil
}

func (i impl) RunRegenStep1(applicant dbmodels.Applicant) (ok bool, err error) {

	vacancy, err := i.vacancyStore.GetByID(applicant.SpaceID, applicant.VacancyID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения вакансии")
	}
	if vacancy == nil {
		return false, nil
	}

	vacancyInfo, applicantInfo, questions, applicantAnswers, generated, err := i.getStep1Data(applicant, *vacancy, applicant.ApplicantVkStep.Step1.Questions)
	if err != nil {
		return false, err
	}
	// запуск ИИ
	rec := *applicant.ApplicantVkStep
	newQuestions, comments, err := i.vkAiProvider.VkStep1Regen(vacancy.SpaceID, vacancy.ID, vacancyInfo, applicantInfo, questions, applicantAnswers, generated)
	if err != nil {
		i.step1Fail(applicant, rec)
		return false, errors.Wrap(err, "ошибка вызова ИИ при перегенерации черновика скрипта")
	}

	// maps.Copy(rec.Step1.Comments, comments)

	questionResult := []dbmodels.VkStep1Question{}
	for k, question := range rec.Step1.Questions {
		// вопросы без изменений
		if !question.NotSuitable {
			questionRec := dbmodels.VkStep1Question{
				ID:                question.ID,
				Text:              question.Text,
				Order:             k,
				NotSuitable:       false,
				NotSuitableReason: "",
			}
			questionResult = append(questionResult, questionRec)
			continue
		}
		currentQID := question.ID
		// обновленнные вопросы
		newQuestion := newQuestions[0]
		newQuestions = newQuestions[1:]
		questionRec := dbmodels.VkStep1Question{
			ID:                currentQID,
			Text:              newQuestion.Text,
			Order:             k,
			NotSuitable:       false,
			NotSuitableReason: "",
		}
		questionResult = append(questionResult, questionRec)
		rec.Step1.Comments[currentQID] = comments[newQuestion.ID]
	}

	rec.Step1.Questions = questionResult
	rec.Status = dbmodels.VkStep1Draft
	_, err = i.vkStore.Save(rec)
	if err != nil {
		return false, errors.Wrap(err, "ошибка сохранения черновика скрипта после пергенерации")
	}
	return false, nil
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

func (i impl) step1Fail(applicant dbmodels.Applicant, rec dbmodels.ApplicantVkStep) {
	rec.Status = dbmodels.VkStep1DraftFail
	_, err := i.vkStore.Save(rec)
	if err != nil {
		i.getLogger(applicant.SpaceID, applicant.ID).
			WithError(err).Error("ВК. Шаг 1. Ошибка изменения статуса")
	}
}

func (i impl) getStep1Data(applicant dbmodels.Applicant, vacancy dbmodels.Vacancy, stepQuestions []dbmodels.VkStep1Question) (vacancyInfo, applicantInfo, questions, applicantAnswers, generated string, err error) {

	vacancyInfo, err = surveyapimodels.GetVacancyDataContent(vacancy)
	if err != nil {
		return "", "", "", "", "", err
	}

	// получение данных кандидата для промта
	applicantInfo, err = surveyapimodels.GetApplicantDataContent(applicant)
	if err != nil {
		return "", "", "", "", "", err
	}
	// получение вопросов для промта
	questions, err = surveyapimodels.QuestionsStep0.Content()
	if err != nil {
		return "", "", "", "", "", err
	}

	// получение ответов кандидата для промта
	applicantAnswers, err = applicant.ApplicantVkStep.Step0.AnswerContent()
	if err != nil {
		return "", "", "", "", "", err
	}
	generated = ""
	if len(stepQuestions) != 0 {
		body, err := json.Marshal(stepQuestions)
		if err != nil {
			return "", "", "", "", "", errors.Wrap(err, "ошибка десериализации структуры вопросов на нерегенерацию шага 1")
		}
		generated = string(body)
	}
	return vacancyInfo, applicantInfo, questions, applicantAnswers, generated, nil
}
