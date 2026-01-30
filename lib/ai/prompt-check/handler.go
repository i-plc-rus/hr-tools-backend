package promptcheckhandler

import (
	"bytes"
	"context"
	"hr-tools-backend/db"
	ollamasearchhandler "hr-tools-backend/lib/ai/ollama-search"
	promptcheckstore "hr-tools-backend/lib/ai/prompt-check-store"
	"hr-tools-backend/lib/vk"
	aiapimodels "hr-tools-backend/models/api/ai"
	surveyapimodels "hr-tools-backend/models/api/survey"
	dbmodels "hr-tools-backend/models/db"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	Status() (data aiapimodels.StatusResponse)
	ExecutionInfo(id string) (data aiapimodels.ExecutionResult, err error)
	RunPrompt(prompt string) (id string, err error)
	RunQuestionsPrompt(prompt string) (id string, err error)
	RunQuestionsPromptOnApplicant(promptTpl, spaceID, applicantID string) (id string, err error)
	ValidateQPromptTemplate(promptTpl, spaceID, applicantID string) (result aiapimodels.ValidateQPromptResponse, err error)
}

var Instance Provider

func NewHandler(ctx context.Context) {
	Instance = &impl{
		ctx:              ctx,
		promptCheckStore: promptcheckstore.NewInstance(db.DB),
		ollama:           ollamasearchhandler.GetHandler(ctx),
	}
}

type impl struct {
	ctx              context.Context
	ollama           ollamasearchhandler.Provider
	aiBusy           atomic.Bool
	last             string
	lastAt           time.Time
	promptCheckStore promptcheckstore.Provider
}

func (i *impl) Status() (data aiapimodels.StatusResponse) {
	return aiapimodels.StatusResponse{
		IsFree:             !i.aiBusy.Load(),
		ExecutingRequestID: i.last,
	}
}

func (i *impl) ExecutionInfo(id string) (data aiapimodels.ExecutionResult, err error) {
	rec, err := i.promptCheckStore.GetByID(id)
	if err != nil {
		return aiapimodels.ExecutionResult{}, err
	}
	if rec == nil {
		return aiapimodels.ExecutionResult{}, errors.New("данные не найдены")
	}

	data = aiapimodels.ExecutionResult{
		SysPromt:   rec.SysPromt,
		UserPromt:  rec.UserPromt,
		Answer:     rec.Answer,
		ReqestType: rec.ReqestType,
		Status:     rec.Status,
	}

	//TODO render ParsedData
	if rec.Status == dbmodels.PromptExecutionResponse {
		rData, err := i.renderAnswer(rec.Answer, rec.ReqestType)
		if err != nil {
			data.ParsedData = err.Error()
		} else {
			data.ParsedData = rData
		}
	}

	return data, nil
}

func (i *impl) RunPrompt(prompt string) (id string, err error) {
	id, err = i.startExecution("", prompt, dbmodels.PromptTypeCheck)
	if err != nil {
		return "", err
	}
	go i.execute(id, prompt)
	return id, nil
}

func (i *impl) RunQuestionsPrompt(prompt string) (id string, err error) {
	id, err = i.startExecution("", prompt, dbmodels.PromptTypeQuestions)
	if err != nil {
		return "", err
	}
	go i.execute(id, prompt)
	return id, nil
}

func (i *impl) ValidateQPromptTemplate(promptTpl, spaceID, applicantID string) (result aiapimodels.ValidateQPromptResponse, err error) {

	prompt, tplData, err := i.questionsPromptTplBuild(promptTpl, spaceID, applicantID)
	if err != nil {
		return aiapimodels.ValidateQPromptResponse{}, err
	}

	missedTags := []string{}
	for _, tag := range tplData.GetTagList() {
		if !strings.Contains(promptTpl, tag) {
			missedTags = append(missedTags, tag)
		}
	}

	return aiapimodels.ValidateQPromptResponse{
		ResultPrompt: prompt,
		MissedTags:   missedTags,
	}, nil
}

func (i *impl) RunQuestionsPromptOnApplicant(promptTpl, spaceID, applicantID string) (id string, err error) {
	prompt, _, err := i.questionsPromptTplBuild(promptTpl, spaceID, applicantID)
	if err != nil {
		return "", err
	}

	id, err = i.startExecution("", prompt, dbmodels.PromptTypeQuestions)
	if err != nil {
		return "", err
	}
	go i.execute(id, prompt)
	return id, nil
}

// запрос к модели
func (i *impl) execute(id, prompt string) {
	defer i.unlockExecution()

	now := time.Now()
	response, err := i.ollama.QueryOllama(prompt)
	logger := log.
		WithField("id", id)
	if err != nil {
		logger.
			WithError(err).
			Error(err, "ошибка выполнения запроса к ИИ для проверки промпта")
		i.updateExecutionData(id, err.Error(), dbmodels.PromptExecutionError)
		return
	}

	logger.
		WithField("answer", response).
		WithField("answer_duration_sec", time.Now().Sub(now).Seconds()).
		Info("Ответ AI на запрос проверки промпта")
	answer := i.ollama.ExtractAnswer(response)
	i.updateExecutionData(id, answer, dbmodels.PromptExecutionResponse)
}

// формирование и проверка ответа
func (i *impl) renderAnswer(answer string, rType dbmodels.PromptType) (data any, err error) {
	switch rType {
	case dbmodels.PromptTypeCheck:
		return nil, nil
	case dbmodels.PromptTypeQuestions:
		questions, comments, err := ollamasearchhandler.ParseVk1QuestionsAIResponse(answer)
		if err != nil {
			return nil, err
		}
		return surveyapimodels.VkStep1{
			Questions:   questions,
			Comments:    comments,
			ScriptIntro: "заполняется другим промптом",
			ScriptOutro: "заполняется другим промптом",
		}, nil
	default:
		return nil, nil
	}
}

func (i *impl) saveExecutionData(sysPromt, userPromt string, rType dbmodels.PromptType) (string, error) {
	rec := dbmodels.PromptExecution{
		SysPromt:   sysPromt,
		UserPromt:  userPromt,
		Answer:     "",
		ReqestType: rType,
		Status:     dbmodels.PromptExecutionSent,
	}
	id, err := i.promptCheckStore.Save(rec)
	if err != nil {
		return "", errors.Wrap(err, "ошибка сохранения запроса выполнения промтра ИИ")
	}
	return id, nil
}

func (i *impl) updateExecutionData(id, answer string, status dbmodels.PromptExecutionStatus) (string, error) {
	updMap := map[string]any{
		"answer": answer,
		"status": status,
	}
	err := i.promptCheckStore.Update(id, updMap)
	if err != nil {
		return "", errors.Wrap(err, "ошибка обновления запроса выполнения промтра ИИ")
	}
	return id, nil
}

func (i *impl) startExecution(sysPromt, prompt string, rType dbmodels.PromptType) (id string, err error) {
	if err = i.lockExecution(); err != nil {
		return "", err
	}
	id, err = i.saveExecutionData(sysPromt, prompt, rType)
	if err != nil {
		i.unlockExecution()
		return "", err
	}
	i.last = id
	return id, nil
}

func (i *impl) lockExecution() error {
	if !i.aiBusy.CompareAndSwap(false, true) {
		return errors.New("сервис уже обрабатывает запрос")
	}
	return nil
}

func (i *impl) unlockExecution() {
	i.aiBusy.Store(false)
	i.last = ""
}

func (i *impl) questionsPromptTplBuild(promptTpl string, spaceID, applicantID string) (result string, tplData aiapimodels.ApplicantTemplateData, err error) {
	tpl, err := template.New("ValidateQPromptTemplate").Parse(promptTpl)
	if err != nil {
		return "", tplData, err
	}

	aiData, err := vk.Instance.Step1GetData(spaceID, applicantID)
	if err != nil {
		return "", tplData, err
	}
	tplData = aiapimodels.ApplicantTemplateData{
		Vacancy:                 aiData.VacancyInfo,
		Requirements:            aiData.Requirements,
		Applicant:               aiData.ApplicantInfo,
		TypicalQuestions:        aiData.Questions,
		TypicalQuestionsAnswers: aiData.ApplicantAnswers,
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, tplData)
	if err != nil {
		return "", tplData, err
	}
	return buf.String(), tplData, nil
}
