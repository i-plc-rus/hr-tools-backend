package vk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	ollamasearchhandler "hr-tools-backend/lib/ai/ollama-search"
	"hr-tools-backend/lib/applicant"
	applicantstore "hr-tools-backend/lib/applicant/store"
	companystore "hr-tools-backend/lib/dicts/company/store"
	negotiationchathandler "hr-tools-backend/lib/external-services/negotiation-chat"
	filestorage "hr-tools-backend/lib/file-storage"
	gpthandler "hr-tools-backend/lib/gpt"
	messagetemplate "hr-tools-backend/lib/message-template"
	"hr-tools-backend/lib/smtp"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	"hr-tools-backend/lib/utils/helpers"
	initchecker "hr-tools-backend/lib/utils/init-checker"
	videonormalize "hr-tools-backend/lib/utils/video-normalize"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	questionhistorystore "hr-tools-backend/lib/vk/question-history-store"
	vkvideoanalyzestore "hr-tools-backend/lib/vk/vk-video-analyze-store"
	"hr-tools-backend/models"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"
	surveyapimodels "hr-tools-backend/models/api/survey"
	dbmodels "hr-tools-backend/models/db"
	"io"
	"mime/multipart"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	RunStep0(applicant dbmodels.Applicant) (ok bool, err error)
	GetSurveyStep0(id string) (*surveyapimodels.VkStep0SurveyView, error)                                                              // анкета для фронта
	HandleSurveyStep0(id string, answers surveyapimodels.VkStep0SurveyAnswers) (result surveyapimodels.VkStep0SurveyResult, err error) // ответы от фронта, сохранение в бд, анализ проходит или нет
	Step1GetData(spaceID, applicantID string) (aiData surveyapimodels.AiData, err error)
	RunStep1(applicant dbmodels.Applicant) (ok bool, err error)
	UpdateStep1(spaceID, applicantID string, stepData surveyapimodels.VkStep1Update) (hMsg string, err error)
	RegenStep1(spaceID, applicantID string, stepData surveyapimodels.VkStep1Regen) (hMsg string, err error)
	RunRegenStep1(applicant dbmodels.Applicant) (ok bool, err error)
	GetVideoSurvey(id string) (*surveyapimodels.VkStep1SurveyView, error)
	UploadVideoAnswer(ctx context.Context, id, questionID string, fileHeader *multipart.FileHeader) error
	GetVideoAnswer(ctx context.Context, id, questionID string) (reader io.Reader, err error)
	ScoreAnswer(videoSurveyRec dbmodels.ApplicantVkVideoSurvey) (err error)
	GenerateReport(vkRec dbmodels.ApplicantVkStep) (ok bool, err error)
	UploadStreamVideoAnswer(ctx context.Context, id, questionID string, reader io.Reader, fileName, contentType string) (info minio.UploadInfo, err error)
	VideoRetry(analyzeID, userID string) error
	VideoSkip(analyzeID, userID string) error
}

var Instance Provider

const (
	defaultCompanyName = "HR-Tools"
	Step0SucessMsg     = "Ваша анкета была успешно заполнена, с вами свяжутся, чтобы сообщить о результатах"
	Step0FailMsg       = "Ваша анкета была успешно заполнена, с вами свяжутся, чтобы сообщить о результатах."
	Step0Done          = "Ваша анкета успешно заполнена, спасибо за уделенное время."
)

func NewHandler(ctx context.Context) {
	instance := impl{
		ctx:                    ctx,
		vacancyStore:           vacancystore.NewInstance(db.DB),
		applicantStore:         applicantstore.NewInstance(db.DB),
		vkStore:                applicantvkstore.NewInstance(db.DB),
		negotiationChatHandler: negotiationchathandler.Instance,
		companyStore:           companystore.NewInstance(db.DB),
		messageTemplate:        messagetemplate.Instance,
		questionHistoryStore:   questionhistorystore.NewInstance(db.DB),
		spaceSettingsStore:     spacesettingsstore.NewInstance(db.DB),
		vkVideoAnalyzeStore:    vkvideoanalyzestore.NewInstance(db.DB),
	}
	if config.Conf.AI.VkStep1AI == "Ollama" {
		instance.vkAiProvider = ollamasearchhandler.GetHandler(ctx)
	} else {
		instance.vkAiProvider = gpthandler.GetHandler(false)
	}
	initchecker.CheckInit(
		"vacancyStore", instance.vacancyStore,
		"applicantStore", instance.applicantStore,
		"vkStore", instance.vkStore,
		"negotiationChatHandler", instance.negotiationChatHandler,
		"companyStore", instance.companyStore,
		"messageTemplate", instance.messageTemplate,
		"questionHistoryStore", instance.questionHistoryStore,
		"spaceSettingsStore", instance.spaceSettingsStore,
		"vkVideoAnalyzeStore", instance.vkVideoAnalyzeStore,
	)
	Instance = instance
}

type impl struct {
	ctx                    context.Context
	vacancyStore           vacancystore.Provider
	applicantStore         applicantstore.Provider
	vkStore                applicantvkstore.Provider
	negotiationChatHandler negotiationchathandler.Provider
	companyStore           companystore.Provider
	messageTemplate        messagetemplate.Provider
	vkAiProvider           surveyapimodels.VkAiProvider // при необходимости поменяем пакет имплементации, пока через настройку config.Conf.AI.VkStep1AI
	questionHistoryStore   questionhistorystore.Provider
	spaceSettingsStore     spacesettingsstore.Provider
	vkVideoAnalyzeStore    vkvideoanalyzestore.Provider
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
	link := rec.GetStep0SurveyUrl(config.Conf)
	companyName := i.getCompanyName(applicantRec.SpaceID, applicantRec.Vacancy.CompanyID)

	chatText, err := messagetemplate.GetSurvayStep0SuggestMessage(companyName, link, false)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка получения сообщения со ссылкой на анкету для отправки кандидату через чат")
		chatText = ""
	}
	emailText, err := messagetemplate.GetSurvaySuggestMessage(companyName, link, true)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка получения сообщения со ссылкой на анкету для отправки кандидату через email")
		emailText = ""
	}
	title := messagetemplate.GetSurvaySuggestTitle()
	isSend := i.sendLink(applicantRec, chatText, emailText, title)
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
	if rec.Status >= dbmodels.VkStep0Answered {
		result = surveyapimodels.VkStep0SurveyResult{
			Success: true,
			Message: Step0Done,
		}
		return result, nil
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

func (i impl) Step1GetData(spaceID, applicantID string) (aiData surveyapimodels.AiData, err error) {
	applicant, err := i.applicantStore.GetByID(spaceID, applicantID)
	if err != nil {
		return surveyapimodels.AiData{}, errors.Wrap(err, "ошибка получения кандидата")
	}
	if applicant == nil {
		return surveyapimodels.AiData{}, errors.New("кандидат не найден")
	}
	vacancy, err := i.vacancyStore.GetByID(applicant.SpaceID, applicant.VacancyID)
	if err != nil {
		return surveyapimodels.AiData{}, errors.Wrap(err, "ошибка получения вакансии")
	}
	if vacancy == nil {
		return surveyapimodels.AiData{}, errors.New("вакансия не найдена")
	}
	return i.getStep1Data(applicant.Applicant, *vacancy, nil)
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
		if helpers.IsContextDone(i.ctx) {
			return false, nil
		}
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
		rec.Status != dbmodels.VkStep1DraftFail &&
		rec.Status != dbmodels.VkStep1Approved {
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
	if rec.Status == dbmodels.VkStep1Approved {
		// сохраняем подтвержденные вопросы для будущего использования
		i.storeQuestions(*rec)
		// отправляем приглашение на видео интервью
		if i.sendVideoSurvaySuggest(*rec) {
			rec.Status = dbmodels.VkStepVideoSuggestSent
			rec.VideoInterviewInviteDate = time.Now()
			_, err = i.vkStore.Save(*rec)
			if err != nil {
				return "", errors.Wrap(err, "ошибка обновления статуса анкеты, при отправке кандидату приглашения на видео интервью")
			}
		}
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
		if helpers.IsContextDone(i.ctx) {
			return false, nil
		}
		i.step1Fail(applicant, rec)
		return false, errors.Wrap(err, "ошибка вызова ИИ при перегенерации черновика скрипта")
	}

	questionResult := []dbmodels.VkStep1Question{}
	var newQuestion surveyapimodels.VkStep1Question
	haveNewQuestion := false

	qCount := 0
	for k, question := range rec.Step1.Questions {
		qCount = k
		if !haveNewQuestion && len(newQuestions) > 0 {
			newQuestion = newQuestions[0]
			if len(newQuestions) > 1 {
				newQuestions = newQuestions[1:]
			} else {
				newQuestions = []surveyapimodels.VkStep1Question{}
			}
			haveNewQuestion = true
		}
		// вопросы без изменений
		if !question.NotSuitable || !haveNewQuestion {
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
		questionRec := dbmodels.VkStep1Question{
			ID:                currentQID,
			Text:              newQuestion.Text,
			Order:             k,
			NotSuitable:       false,
			NotSuitableReason: "",
		}
		questionResult = append(questionResult, questionRec)
		rec.Step1.Comments[currentQID] = comments[newQuestion.ID]
		haveNewQuestion = false
	}
	// добавляем вопросы, если их меньше 15
	if len(questionResult) < 15 && haveNewQuestion {
		qCount++
		qID := fmt.Sprintf("q%v", qCount+1)
		questionResult = append(questionResult, dbmodels.VkStep1Question{
			ID:                qID,
			Text:              newQuestion.Text,
			Order:             qCount,
			NotSuitable:       false,
			NotSuitableReason: "",
		})
		rec.Step1.Comments[qID] = comments[newQuestion.ID]
	}
	if len(questionResult) < 15 {
		for _, q := range newQuestions {
			qCount++
			qID := fmt.Sprintf("q%v", qCount+1)
			questionResult = append(questionResult, dbmodels.VkStep1Question{
				ID:                qID,
				Text:              q.Text,
				Order:             qCount,
				NotSuitable:       false,
				NotSuitableReason: "",
			})
			rec.Step1.Comments[qID] = comments[q.ID]
			if len(questionResult) >= 15 {
				break
			}
		}
	}

	rec.Step1.Questions = questionResult
	rec.Status = dbmodels.VkStep1Draft
	_, err = i.vkStore.Save(rec)
	if err != nil {
		return false, errors.Wrap(err, "ошибка сохранения черновика скрипта после пергенерации")
	}
	return false, nil
}

func (i impl) GetVideoSurvey(id string) (*surveyapimodels.VkStep1SurveyView, error) {
	rec, err := i.vkStore.GetByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения анкеты кандидата")
	}
	if rec == nil {
		return nil, errors.New("анкета не найдена")
	}
	result := surveyapimodels.VkStep1SurveyView{
		Questions:   []surveyapimodels.VkStep1SurveyQuestion{},
		ScriptIntro: rec.Step1.ScriptIntro,
		ScriptOutro: rec.Step1.ScriptOutro,
	}
	for _, q := range rec.Step1.Questions {
		result.Questions = append(result.Questions,
			surveyapimodels.VkStep1SurveyQuestion{
				ID:    q.ID,
				Text:  q.Text,
				Order: q.Order,
			})
	}
	return &result, nil
}

func (i impl) UploadVideoAnswer(ctx context.Context, id, questionID string, fileHeader *multipart.FileHeader) error {
	rec, err := i.vkStore.GetByID(id)
	if err != nil {
		return errors.Wrap(err, "ошибка получения анкеты кандидата")
	}
	if rec == nil {
		return errors.New("анкета не найдена")
	}
	if answer, ok := rec.VideoInterview.Answers[questionID]; ok && answer.FileID != "" {
		return errors.New("ответ уже сохранен")
	}

	// Проверяем тип файла
	contentType := helpers.GetFileContentType(fileHeader)
	if !strings.HasPrefix(contentType, "video/") {
		return errors.New("Ожидается файл с видео")
	}
	// Открываем загруженный файл
	buffer, err := fileHeader.Open()
	if err != nil {
		return errors.Wrap(err, "Не удалось открыть файл")
	}
	defer buffer.Close()

	fileInfo := dbmodels.UploadFileInfo{
		SpaceID:        rec.SpaceID,
		ApplicantID:    rec.ApplicantID,
		FileName:       questionID,
		FileType:       dbmodels.ApplicantVideoInterview,
		ContentType:    contentType,
		IsUniqueByName: true,
	}
	fileID, err := filestorage.Instance.UploadObject(ctx, fileInfo, buffer, int(fileHeader.Size))
	if err != nil {
		return err
	}
	if rec.VideoInterview.Answers == nil {
		rec.VideoInterview = dbmodels.VideoInterview{
			Answers: map[string]dbmodels.VkVideoAnswer{},
		}
	}
	rec.VideoInterview.Answers[questionID] = dbmodels.VkVideoAnswer{
		FileID: fileID,
	}
	_, err = i.vkStore.Save(*rec)
	if err != nil {
		i.getLogger(rec.SpaceID, rec.ApplicantID).
			WithError(err).
			WithField("question_id", questionID).
			WithField("file_id", fileID).
			Error("ошибка добваления информации о видео файле в базу")
	}
	return nil
}

func (i impl) UploadStreamVideoAnswer(ctx context.Context, id, questionID string, body io.Reader, fileName, contentType1 string) (info minio.UploadInfo, err error) {
	rec, err := i.vkStore.GetByID(id)
	if err != nil {
		return minio.UploadInfo{}, errors.Wrap(err, "ошибка получения анкеты кандидата")
	}
	if rec == nil {
		return minio.UploadInfo{}, errors.New("анкета не найдена")
	}
	if answer, ok := rec.VideoInterview.Answers[questionID]; ok && answer.FileID != "" {
		return minio.UploadInfo{}, errors.New("ответ уже сохранен")
	}

	// Читаем первые 512 байт для определения типа
	buf := make([]byte, 512)
	n, err := body.Read(buf)
	if err != nil && err != io.EOF {
		return minio.UploadInfo{}, errors.Wrap(err, "Не удалось определить тип файла")
	}

	// Определяем MIME тип
	contentType := helpers.DetectFileContentType(fileName, buf[:n])

	if !strings.HasPrefix(contentType, "video/") {
		return minio.UploadInfo{}, errors.New("Ожидается файл с видео")
	}

	// Создаем новый reader, который включает прочитанные байты
	reader := io.MultiReader(bytes.NewReader(buf[:n]), body)

	fileInfo := dbmodels.UploadFileInfo{
		SpaceID:        rec.SpaceID,
		ApplicantID:    rec.ApplicantID,
		FileName:       questionID,
		FileType:       dbmodels.ApplicantVideoInterview,
		ContentType:    contentType,
		IsUniqueByName: true,
	}
	info, err = filestorage.Instance.UploadObjectFromStream(ctx, fileInfo, reader)
	if err != nil {
		return minio.UploadInfo{}, err
	}
	if rec.VideoInterview.Answers == nil {
		rec.VideoInterview = dbmodels.VideoInterview{
			Answers: map[string]dbmodels.VkVideoAnswer{},
		}
	}

	// Нормализация видео если включено в настройках
	if config.Conf.Survey.VideoNormalizeEnabled {
		normalizedVideo, err := videonormalize.Run(ctx, fileInfo, info.Location)
		if err != nil {
			log.WithError(err).Error(err, "ошибка нормализации видео")
			// продолжаем работу даже если нормализация видео не удалась
		} else {
			info.Location = normalizedVideo // используем нормализованный файл
		}
	}

	rec.VideoInterview.Answers[questionID] = dbmodels.VkVideoAnswer{
		FileID: info.Location,
	}
	_, err = i.vkStore.Save(*rec)
	if err != nil {
		i.getLogger(rec.SpaceID, rec.ApplicantID).
			WithError(err).
			WithField("question_id", questionID).
			WithField("file_id", info.Location).
			Error("ошибка добваления информации о видео файле в базу")
	}
	return info, nil
}

func (i impl) GetVideoAnswer(ctx context.Context, id, questionID string) (reader io.Reader, err error) {
	rec, err := i.vkStore.GetByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения анкеты кандидата")
	}
	if rec == nil {
		return nil, errors.New("анкета не найдена")
	}
	answer, ok := rec.VideoInterview.Answers[questionID]
	if ok {
		return filestorage.Instance.GetFileObject(ctx, rec.SpaceID, answer.FileID)
	}
	// если файла нет
	return nil, nil
}

func (i impl) ScoreAnswer(videoSurveyRec dbmodels.ApplicantVkVideoSurvey) (err error) {
	// получаем анкету с вопросами
	rec, err := i.vkStore.GetByID(videoSurveyRec.ApplicantVkStepID)
	if err != nil {
		i.failScoreAnswer(videoSurveyRec, "ошибка получения анкеты кандидата")
		return errors.Wrap(err, "ошибка получения анкеты кандидата")
	}
	// заполняем данные для промта, вопрос/коммент/ответ
	aiData := surveyapimodels.SemanticData{
		Answer: videoSurveyRec.TranscriptText,
	}
	for _, question := range rec.Step1.Questions {
		if question.ID == videoSurveyRec.QuestionID {
			aiData.Question = question.Text
			break
		}
	}
	if aiData.Question == "" {
		i.failScoreAnswer(videoSurveyRec, "вопрос не найден")
		return errors.New("вопрос не найден")
	}
	aiData.Comment = rec.Step1.Comments[videoSurveyRec.QuestionID]

	// вызываем ИИ
	result, err := i.vkAiProvider.VkStep9Score(aiData)
	if err != nil {
		i.failScoreAnswer(videoSurveyRec, "ошибка оценки")
		return errors.Wrap(err, "ошибка оценки")
	}
	logger := i.getLogger(rec.SpaceID, rec.ApplicantID).
		WithField("question_id", videoSurveyRec.QuestionID).
		WithField("similarity", result.Similarity).
		WithField("comment_for_similarity", result.Comment)

	logger.Info("получен результат оценки ответа на вопрос")
	videoSurveyRec.IsSemanticEvaluated = true
	videoSurveyRec.Similarity = result.Similarity
	videoSurveyRec.CommentForSimilarity = result.Comment
	_, err = i.vkVideoAnalyzeStore.Save(videoSurveyRec)
	if err != nil {
		return errors.Wrap(err, "ошибка сохранения результата оценки ответа на вопрос")
	}
	return nil
}

func (i impl) GenerateReport(vkRec dbmodels.ApplicantVkStep) (ok bool, err error) {
	applicantRec, err := i.applicantStore.GetByID(vkRec.SpaceID, vkRec.ApplicantID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения кандидата")
	}

	vacancy, err := i.vacancyStore.GetByID(applicantRec.SpaceID, applicantRec.VacancyID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения вакансии")
	}
	if vacancy == nil {
		return false, nil
	}

	// получение данных кандидата для промта
	applicantInfo, err := surveyapimodels.GetApplicantDataContent(applicantRec.Applicant)
	if err != nil {
		return false, err
	}

	vacancyInfo, requirements, err := surveyapimodels.GetVacancyAiDataContent(*vacancy)
	if err != nil {
		return false, err
	}

	questions, err := surveyapimodels.GetInterviewQuestionsContent(vkRec)
	if err != nil {
		return false, err
	}

	answers, err := surveyapimodels.GetInterviewAnswersContent(vkRec)
	if err != nil {
		return false, err
	}

	evalutions, err := surveyapimodels.GetInterviewEvalutionsContent(vkRec)
	if err != nil {
		return false, err
	}

	aiReportData := surveyapimodels.ReportRequestData{
		VacancyInfo:      vacancyInfo,
		Requirements:     requirements,
		ApplicantInfo:    applicantInfo,
		Questions:        questions,
		ApplicantAnswers: answers,
		Evalutions:       evalutions,
		TotalScore:       vkRec.TotalScore,
		Threshold:        vkRec.Threshold,
	}

	// запуск ИИ
	resp, err := i.vkAiProvider.VkStep11Report(vacancy.SpaceID, vacancy.ID, aiReportData)
	if err != nil {
		if helpers.IsContextDone(i.ctx) {
			return false, nil
		}
		return false, errors.Wrap(err, "ошибка вызова ИИ при генерации отчета по интервью")
	}
	vkRec.OverallComment = resp.OverallComment
	vkRec.Status = dbmodels.VkStep11Report
	_, err = i.vkStore.Save(vkRec)
	if err != nil {
		return false, errors.Wrap(err, "ошибка сохранения анкеты")
	}
	return true, nil
}

func (i impl) VideoRetry(analyzeID, userID string) error {
	rec, err := i.vkVideoAnalyzeStore.GetByID(analyzeID)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("запись не найдена")
	}
	rec.ManualSkip = false
	rec.ManualRetry = true
	rec.ManualUserID = userID
	_, err = i.vkVideoAnalyzeStore.Save(*rec)
	return err
}

func (i impl) VideoSkip(analyzeID, userID string) error {
	rec, err := i.vkVideoAnalyzeStore.GetByID(analyzeID)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("запись не найдена")
	}
	rec.ManualSkip = true
	rec.ManualRetry = false
	rec.ManualUserID = userID
	_, err = i.vkVideoAnalyzeStore.Save(*rec)
	return err
}

func (i impl) sendLink(applicantRec dbmodels.Applicant, chatText, emailText, emailTitle string) (isSend bool) {
	logger := i.getLogger(applicantRec.SpaceID, applicantRec.ID)
	if chatText != "" {
		isChatAvailable := false
		availability, err := i.negotiationChatHandler.IsVailable(applicantRec.SpaceID, applicantRec.ID)
		if err != nil {
			logger.
				WithError(err).
				Warn("ошибка проверки доступности чата с кандидатом")
		} else {
			isChatAvailable = availability.IsAvailable
		}

		isSend = false
		if isChatAvailable {
			if i.sendToChat(applicantRec.SpaceID, applicantRec.ID, chatText, logger) {
				isSend = true
			}
		}
	}
	if applicantRec.Email != "" && emailText != "" {
		emailFrom, err := i.messageTemplate.GetSenderEmail(applicantRec.SpaceID)
		if err != nil {
			logger.
				WithError(err).
				Warn("ошибка получения почты компании для отправки сообщения с ссылкой на анкету на email кандидату")
		} else if emailFrom != "" {
			if i.sendToEmail(emailFrom, applicantRec.Email, emailText, emailTitle, logger) {
				isSend = true
			}
		}
	}
	return isSend
}

func (i impl) sendToChat(spaceID, applicantID, text string, logger *log.Entry) bool {
	req := negotiationapimodels.NewMessageRequest{
		ApplicantID: applicantID,
		Text:        text,
	}
	err := i.negotiationChatHandler.SendMessage(spaceID, req)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка отправки сообщения в чат с кандидатом")
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

func (i impl) sendToEmail(emailFrom, mailTo, text, title string, logger *log.Entry) bool {
	err := smtp.Instance.SendHtmlEMail(emailFrom, mailTo, text, title, nil)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка отправки сообщения на email кандидату")
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

func (i impl) storeQuestions(approvedRec dbmodels.ApplicantVkStep) {
	logger := i.getLogger(approvedRec.SpaceID, approvedRec.ApplicantID)
	applicant, err := i.applicantStore.GetByID(approvedRec.SpaceID, approvedRec.ApplicantID)
	if err != nil {
		logger.WithError(err).Warn("ошибка сохранения подтвержденных вопросов для интервью, не удалось получить данные кандидата")
		return
	}
	if applicant == nil {
		logger.Warn("ошибка сохранения подтвержденных вопросов для интервью, данные кандидата не найдены")
		return
	}

	for _, q := range approvedRec.Step1.Questions {
		rec := dbmodels.QuestionHistory{
			VacancyID:    applicant.VacancyID,
			JobTitleName: "",
			VacancyName:  "",
			Text:         q.Text,
			Comment:      approvedRec.Step1.Comments[q.ID],
		}
		if applicant.Vacancy != nil {
			rec.VacancyName = applicant.Vacancy.VacancyName
			if applicant.Vacancy.JobTitle != nil {
				rec.JobTitleName = applicant.Vacancy.JobTitle.Name
			}
		}
		err = i.questionHistoryStore.Save(rec)
		if err != nil {
			logger.WithError(err).Warn("ошибка сохранения подтвержденных вопросов для интервью в бд")
			return
		}
	}
}

func (i impl) sendVideoSurvaySuggest(approvedRec dbmodels.ApplicantVkStep) (isSend bool) {
	// отправка приглашения на видео интервью
	logger := i.getLogger(approvedRec.SpaceID, approvedRec.ApplicantID)
	applicant, err := i.applicantStore.GetByID(approvedRec.SpaceID, approvedRec.ApplicantID)
	if err != nil {
		logger.WithError(err).Warn("ошибка отправки приглашения на видео интервью, не удалось получить данные кандидата")
		return
	}
	if applicant == nil {
		logger.Warn("ошибка отправки приглашения на видео интервью, данные кандидата не найдены")
		return
	}

	link := approvedRec.GetVideoSurveyUrl(config.Conf)
	companyName := i.getCompanyName(applicant.SpaceID, applicant.Vacancy.CompanyID)

	supportEmail, err := i.getSupportEmail(approvedRec.SpaceID)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка получения почты тех поддержки для шаблона приглашения на видео интервью")
		supportEmail = ""
	}

	chatText, err := messagetemplate.GetVideoSurvaySuggestMessage(applicant.Applicant, companyName, link, supportEmail, false)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка получения текста приглашения на видео интервью для отправки кандидату через чат")
		chatText = ""
	}
	emailText, err := messagetemplate.GetVideoSurvaySuggestMessage(applicant.Applicant, companyName, link, supportEmail, true)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка получения текста приглашения на видео интервью для отправки кандидату через email")
		emailText = ""
	}
	title := messagetemplate.GetSurvaySuggestTitle()
	return i.sendLink(applicant.Applicant, chatText, emailText, title)
}

func (i impl) getSupportEmail(spaceID string) (string, error) {
	email, err := i.spaceSettingsStore.GetValueByCode(spaceID, models.SpaceSenderEmail)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения почты тех поддержки")
	}
	return email, nil
}

func GetVideoAnswerFileName(applicantID, questionID string) string {
	return fmt.Sprintf("%v_%v", applicantID, questionID)
}

func (i impl) failScoreAnswer(rec dbmodels.ApplicantVkVideoSurvey, errMsg string) {
	rec.Error = errMsg
	_, err := i.vkVideoAnalyzeStore.Save(rec)
	if err != nil {
		log.
			WithError(err).
			WithField("vk_step_id", rec.ApplicantVkStepID).
			WithField("question_id", rec.QuestionID).
			Error("ошибка сохранения результата оценки ответа")
	}
}
