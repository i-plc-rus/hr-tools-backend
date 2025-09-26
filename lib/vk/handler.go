package vk

import (
	"encoding/json"
	"fmt"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	ollamasearchhandler "hr-tools-backend/lib/ai/ollama-search"
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
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
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
	Step0SucessMsg     = "Ваша анкета была успешно заполнена, с вами свяжутся, чтобы сообщить о результатах"
	Step0FailMsg       = "Ваша анкета была успешно заполнена, с вами свяжутся, чтобы сообщить о результатах."
)

func NewHandler() {
	i := impl{
		vacancyStore:           vacancystore.NewInstance(db.DB),
		applicantStore:         applicantstore.NewInstance(db.DB),
		vkStore:                applicantvkstore.NewInstance(db.DB),
		negotiationChatHandler: negotiationchathandler.Instance,
		companyStore:           companystore.NewInstance(db.DB),
		messageTemplate:        messagetemplate.Instance,
	}
	if config.Conf.AI.VkStep1AI == "Ollama" {
		i.vkAiProvider = ollamasearchhandler.GetHandler()
	} else {
		i.vkAiProvider = gpthandler.GetHandler(false)
	}
	Instance = i
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
	_, vacancy, err := i.getVacancyAndApplicant(rec.SpaceID, rec.ApplicantID)
	if err != nil {
		return nil, err
	}
	jobTitle := ""
	if vacancy.JobTitle != nil {
		jobTitle = vacancy.JobTitle.Name
	}
	result := surveyapimodels.GetQuestionsStep0(jobTitle)
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
	_, vacancyRec, err := i.getVacancyAndApplicant(rec.SpaceID, rec.ApplicantID)
	if err != nil {
		return result, err
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
	points := i.step0CalcPoints(vacancyRec, request)
	if points > 60 {
		isSucess = true
	}
	//Если кандидат подходит, то переходить к шагу 1
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
	aiData, err := i.getStep1Data(applicant, *vacancy, nil)
	if err != nil {
		return false, err
	}

	// запуск ИИ
	resp, err := i.vkAiProvider.VkStep1(vacancy.SpaceID, vacancy.ID, aiData)
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

	aiData, err := i.getStep1Data(applicant, *vacancy, applicant.ApplicantVkStep.Step1.Questions)
	if err != nil {
		return false, err
	}

	rec := *applicant.ApplicantVkStep
	if aiData.GeneratedQuestions == "" {
		i.step1Fail(applicant, rec)
		return false, errors.Wrap(err, "ошибка вызова ИИ при перегенерации черновика скрипта, вопросы не найдены")
	}
	// запуск ИИ
	newQuestions, comments, err := i.vkAiProvider.VkStep1Regen(vacancy.SpaceID, vacancy.ID, aiData)
	if err != nil {
		i.step1Fail(applicant, rec)
		return false, errors.Wrap(err, "ошибка вызова ИИ при перегенерации черновика скрипта")
	}

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

func (i impl) getStep1Data(applicant dbmodels.Applicant, vacancy dbmodels.Vacancy, stepQuestions []dbmodels.VkStep1Question) (aiData surveyapimodels.AiData, err error) {
	vacancyInfo, requirements, err := surveyapimodels.GetVacancyAiDataContent(vacancy)
	if err != nil {
		return surveyapimodels.AiData{}, err
	}

	// получение данных кандидата для промта
	applicantInfo, err := surveyapimodels.GetApplicantDataContent(applicant)
	if err != nil {
		return surveyapimodels.AiData{}, err
	}
	// получение вопросов для промта
	jobTitle := ""
	if vacancy.JobTitle != nil {
		jobTitle = vacancy.JobTitle.Name
	}
	questions, err := surveyapimodels.GetQuestionsStep0(jobTitle).Content()
	if err != nil {
		return surveyapimodels.AiData{}, err
	}

	// получение ответов кандидата для промта
	applicantAnswers, err := applicant.ApplicantVkStep.Step0.AnswerContent()
	if err != nil {
		return surveyapimodels.AiData{}, err
	}
	aiData = surveyapimodels.AiData{
		VacancyInfo:        vacancyInfo,
		Requirements:       requirements,
		ApplicantInfo:      applicantInfo,
		Questions:          questions,
		ApplicantAnswers:   applicantAnswers,
		GeneratedQuestions: "",
	}
	if len(stepQuestions) != 0 {
		body, err := json.Marshal(stepQuestions)
		if err != nil {
			return surveyapimodels.AiData{}, errors.Wrap(err, "ошибка десериализации структуры вопросов на нерегенерацию шага 1")
		}
		aiData.GeneratedQuestions = string(body)
	}
	return aiData, nil
}

func (i impl) getVacancyAndApplicant(spaceID, applicantID string) (applicant *dbmodels.ApplicantExt, vacancy *dbmodels.Vacancy, err error) {
	applicant, err = i.applicantStore.GetByID(spaceID, applicantID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "ошибка получения данных кандидата")
	}
	if applicant == nil {
		return nil, nil, errors.New("кандидат не найден")
	}
	vacancy, err = i.vacancyStore.GetByID(applicant.SpaceID, applicant.VacancyID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "ошибка получения вакансии")
	}
	if vacancy == nil {
		return nil, nil, errors.New("вакансия не найдена")
	}
	return applicant, vacancy, nil
}

func (i impl) step0CalcPoints(vacancy *dbmodels.Vacancy, request surveyapimodels.VkStep0SurveyAnswers) (points float64) {
	for _, answer := range request.Answers {
		switch answer.QuestionID {
		case "1":
			if answer.Answer != "да" {
				return 0
			}
		case "2":
			expectedSalary, _ := strconv.Atoi(answer.Answer) //валидировали через (v VkStep0SurveyAnswers) Validate()
			v := 0
			if vacancy.Salary.InHand > 0 {
				v = vacancy.Salary.InHand
			} else if vacancy.Salary.ByResult > 0 {
				v = vacancy.Salary.ByResult
			} else if vacancy.Salary.To > 0 {
				v = vacancy.Salary.To
			} else if vacancy.Salary.From > 0 {
				v = vacancy.Salary.From
			} else {
				points += 35
				continue
			}
			if expectedSalary <= v {
				points += 35
				continue
			}
			b := 0.1
			if v < expectedSalary && float64(expectedSalary) <= float64(v)*(float64(1)+b) {
				points += 24.5
				continue
			}
			if float64(expectedSalary) <= float64(v)*(float64(1)+2*b) {
				points += 14
				continue
			}
		case "3":
			if vacancy.Employment.ToString() == answer.Answer {
				points += 20
			}
		case "4":
			if vacancy.Schedule.ToString() == answer.Answer {
				points += 20
			}
		case "5":
			fmt.Println(vacancy.Experience.ToPoint())
			fmt.Println(models.ExperienceFromDescr(answer.Answer).ToPoint())
			if vacancy.Experience.ToPoint() <= models.ExperienceFromDescr(answer.Answer).ToPoint() {
				points += 25
			}
		}
	}
	return points
}
