package ollamasearchhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hr-tools-backend/config"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/lib/utils/lock"
	aimodels "hr-tools-backend/models/ai"
	ollamamodels "hr-tools-backend/models/api/ollama"
	surveyapimodels "hr-tools-backend/models/api/survey"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	QueryOllama(prompt string) (string, error)
	ExtractAnswer(response string) string
}

type impl struct {
	ctx         context.Context
	ollamaURL   string
	ollamaModel string
	ops         ollamamodels.Options
}

func GetHandler(ctx context.Context) *impl {
	log.Infof("Инициализация ИИ: %v, модель: %v", config.Conf.AI.Ollama.OllamaURL, config.Conf.AI.Ollama.OllamaModel)
	return &impl{
		ctx:         ctx,
		ollamaURL:   config.Conf.AI.Ollama.OllamaURL,
		ollamaModel: config.Conf.AI.Ollama.OllamaModel,
		ops:         ollamamodels.GetDeepSeekConfig(),
	}
}

func (i impl) getLogger() *log.Entry {
	return log.
		WithField("ai", "ollama").
		WithField("model", i.ollamaModel)
}

func (i impl) VkStep1(spaceID, vacancyID string, aiData surveyapimodels.AiData) (resp surveyapimodels.VkStep1, err error) {
	err = i.checkConfig()
	if err != nil {
		return resp, err
	}
	attemps := config.Conf.Survey.VkStep1.RetryAttempts + 1
	delaySec := config.Conf.Survey.VkStep1.RetryDelaySec
	genQuestionsFn := func() (aimodels.Vk1QuestionResult, error) {
		return i.genVk1Questions(aiData)
	}
	questionResult, err := helpers.WithRetry(attemps, delaySec, genQuestionsFn)
	if err != nil {
		return resp, errors.Wrap(err, "ошибка генерации вопросов")
	}

	genIntroOutroFn := func() (aimodels.Vk1IntroResult, error) {
		return i.genVk1IntroOutro(aiData)
	}
	introResult, err := helpers.WithRetry(attemps, delaySec, genIntroOutroFn)
	if err != nil {
		return resp, errors.Wrap(err, "ошибка генерации intro/outro")
	}

	resp = surveyapimodels.VkStep1{
		Questions:   questionResult.Questions,
		ScriptIntro: introResult.Intro,
		ScriptOutro: introResult.Outro,
		Comments:    questionResult.Comments,
	}
	return resp, nil
}

func (i impl) VkStep1Regen(spaceID, vacancyID string, aiData surveyapimodels.AiData) (newQuestions []surveyapimodels.VkStep1Question, comments map[string]string, err error) {
	regenQuestionsFn := func() (aimodels.Vk1QuestionResult, error) {
		prompt := fmt.Sprintf(step1QRegenTemplate, aiData.VacancyInfo, aiData.ApplicantInfo, aiData.GeneratedQuestions)

		now := time.Now()
		// запрос к локальной модели
		response, err := i.QueryOllama(prompt)
		if err != nil {
			return aimodels.Vk1QuestionResult{}, errors.Wrap(err, "ошибка получения пула с новыми вопросами")
		}
		i.getLogger().
			WithField("prompt", prompt).
			WithField("answer", response).
			WithField("answer_duration_sec", time.Now().Sub(now).Seconds()).
			Info("Ответ AI на запрос VkStep1Regen")
		return ParseVk1QuestionsAIResponse(response)
	}
	attemps := config.Conf.Survey.VkStep1.RegenRetryAttempts + 1
	delaySec := config.Conf.Survey.VkStep1.RetryDelaySec
	questionResult, err := helpers.WithRetry(attemps, delaySec, regenQuestionsFn)
	if err != nil {
		return nil, nil, errors.Wrap(err, "ошибка перегенерации вопросов")
	}

	return questionResult.Questions, questionResult.Comments, nil
}

func (i impl) VkStep9Score(aiData surveyapimodels.SemanticData) (scoreResult surveyapimodels.VkStep9ScoreResult, err error) {
	prompt := fmt.Sprintf(step9semanticScoreTemplate, aiData.Question, aiData.Comment, aiData.Answer, "%")

	now := time.Now()
	// запрос к локальной модели
	response, err := i.QueryOllama(prompt)
	if err != nil {
		return surveyapimodels.VkStep9ScoreResult{}, errors.Wrap(err, "ошибка оценки ответа кандидата")
	}
	i.getLogger().
		WithField("prompt", prompt).
		WithField("answer", response).
		WithField("answer_duration_sec", time.Now().Sub(now).Seconds()).
		Info("Ответ AI на запрос VkStep9Score")

	return ParseVkStep9ScoreAIResponse(response)
}

func (i impl) VkStep11Report(spaceID, vacancyID string, aiData surveyapimodels.ReportRequestData) (reportResult surveyapimodels.ReportResult, err error) {
	prompt := fmt.Sprintf(step11ReportTemplate, aiData.VacancyInfo, aiData.Requirements, aiData.ApplicantInfo, aiData.Questions, aiData.ApplicantAnswers,
		aiData.Evalutions, aiData.TotalScore, aiData.Threshold)

	now := time.Now()
	// запрос к локальной модели
	response, err := i.QueryOllama(prompt)
	if err != nil {
		return surveyapimodels.ReportResult{}, errors.Wrap(err, "ошибка формирования отчета")
	}
	i.getLogger().
		WithField("prompt", prompt).
		WithField("answer", response).
		WithField("answer_duration_sec", time.Now().Sub(now).Seconds()).
		Info("Ответ AI на запрос VkStep11Report")

	answer := extractAnswer(response)
	return surveyapimodels.ReportResult{OverallComment: answer}, nil
}

func (i impl) checkConfig() error {
	if i.ollamaURL == "" {
		return errors.New("не указан url для ollama")
	}
	if i.ollamaModel == "" {
		return errors.New("не указана модель для ollama")
	}
	return nil
}

func (i impl) genVk1Questions(aiData surveyapimodels.AiData) (result aimodels.Vk1QuestionResult, err error) {
	prompt := fmt.Sprintf(step1QTemplate, aiData.VacancyInfo, aiData.Requirements, aiData.ApplicantInfo, aiData.Questions, aiData.ApplicantAnswers)

	now := time.Now()
	// запрос к локальной модели
	response, err := i.QueryOllama(prompt)
	if err != nil {
		return aimodels.Vk1QuestionResult{}, errors.Wrap(err, "ошибка получения пула вопросов")
	}
	i.getLogger().
		WithField("prompt", prompt).
		WithField("answer", response).
		WithField("answer_duration_sec", time.Now().Sub(now).Seconds()).
		Info("Ответ AI на запрос genVk1Questions")

	return ParseVk1QuestionsAIResponse(response)
}

func (i impl) genVk1IntroOutro(aiData surveyapimodels.AiData) (result aimodels.Vk1IntroResult, err error) {
	prompt := fmt.Sprintf(step1IntroOutroTemplate, aiData.VacancyInfo, aiData.ApplicantInfo)

	now := time.Now()
	// запрос к локальной модели
	response, err := i.QueryOllama(prompt)
	if err != nil {
		return aimodels.Vk1IntroResult{}, errors.Wrap(err, "ошибка получения текстов сценария intro/outro")
	}

	i.getLogger().
		WithField("prompt", prompt).
		WithField("answer", response).
		WithField("answer_duration_sec", time.Now().Sub(now).Seconds()).
		Info("Ответ AI на запрос genVk1IntroOutro")

	return ParseVk1IntroOutroAIResponse(response)
}

func (i impl) QueryOllama(prompt string) (string, error) {

	if !lock.Resource.Acquire(i.ctx, "QueryOllama") {
		return "", errors.New("ошибка доступа к ресурсам - контекст завершен")
	}
	defer lock.Resource.Release("QueryOllama")
	request := ollamamodels.OllamaRequest{
		Model:   i.ollamaModel,
		Prompt:  prompt,
		Stream:  false,
		Options: i.ops,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(i.ctx, "POST", i.ollamaURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка Ollama API: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ollamaResponse ollamamodels.OllamaResponse
	err = json.Unmarshal(body, &ollamaResponse)
	if err != nil {
		return "", err
	}

	return ollamaResponse.Response, nil
}

func ParseVk1QuestionsAIResponse(response string) (result aimodels.Vk1QuestionResult, err error) {
	answer := extractAnswer(response)
	answer = replaceAnswerFormatTag(answer)

	type questionFormat struct {
		ID      string `json:"id"`
		Text    string `json:"text"`
		Comment string `json:"comment"`
	}

	type answerFormat struct {
		Questions []questionFormat `json:"questions"`
	}

	answerData := answerFormat{}
	err = json.Unmarshal([]byte(answer), &answerData)
	if err != nil {
		return aimodels.Vk1QuestionResult{}, err
	}
	questions := []surveyapimodels.VkStep1Question{}
	comments := map[string]string{}
	for k, question := range answerData.Questions {
		questions = append(questions, surveyapimodels.VkStep1Question{
			ID:    question.ID,
			Text:  question.Text,
			Order: k,
		})
		comments[question.ID] = question.Comment
	}

	return aimodels.Vk1QuestionResult{
		Questions: questions,
		Comments:  comments,
	}, nil
}

func ParseVkStep9ScoreAIResponse(response string) (scoreResult surveyapimodels.VkStep9ScoreResult, err error) {
	answer := extractAnswer(response)
	answer = replaceAnswerFormatTag(answer)
	err = json.Unmarshal([]byte(answer), &scoreResult)
	if err != nil {
		return surveyapimodels.VkStep9ScoreResult{}, err
	}
	return scoreResult, nil
}

func ParseVk1IntroOutroAIResponse(response string) (result aimodels.Vk1IntroResult, err error) {
	answer := extractAnswer(response)
	answer = replaceAnswerFormatTag(answer)

	type questionFormat struct {
		ID      string `json:"id"`
		Text    string `json:"text"`
		Comment string `json:"comment"`
	}

	type answerFormat struct {
		ScriptIntro string `json:"script_intro"`
		ScriptOutro string `json:"script_outro"`
	}

	answerData := answerFormat{}
	err = json.Unmarshal([]byte(answer), &answerData)
	if err != nil {
		return aimodels.Vk1IntroResult{}, err
	}
	return aimodels.Vk1IntroResult{
		Intro: answerData.ScriptIntro,
		Outro: answerData.ScriptOutro,
	}, nil
}

func (i impl) ExtractAnswer(response string) string {
	return extractAnswer(response)
}

func extractAnswer(response string) string {
	responseSlice := strings.Split(response, "</think>")
	if len(responseSlice) == 1 {
		return response
	}
	answer := responseSlice[1]
	answer = strings.TrimLeft(answer, "\n")
	return answer
}

func replaceAnswerFormatTag(answer string) string {
	answer = strings.Replace(answer, "```json", "", 1)
	return strings.Replace(answer, "```", "", 1)
}
