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
	"regexp"
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
	logger := i.getLogger().WithField("func", "genVk1Questions")
	questionResult, err := helpers.WithRetry(attemps, delaySec, logger, genQuestionsFn)
	if err != nil {
		return resp, errors.Wrap(err, "ошибка генерации вопросов")
	}

	genIntroOutroFn := func() (aimodels.Vk1IntroResult, error) {
		return i.genVk1IntroOutro(aiData)
	}
	logger = i.getLogger().WithField("func", "genVk1IntroOutro")
	introResult, err := helpers.WithRetry(attemps, delaySec, logger, genIntroOutroFn)
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
	logger := i.getLogger().WithField("func", "VkStep1Regen")

	questionResult, err := helpers.WithRetry(attemps, delaySec, logger, regenQuestionsFn)
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

	type answerFormat struct {
		Questions []questionFormat `json:"questions"`
	}

	answerData := answerFormat{}
	isValidParse := false
	if answer != "" {
		// пробуем стандартную сериализацию
		err = json.Unmarshal([]byte(answer), &answerData)
		if err == nil {
			answerData.Questions, isValidParse = validateAndCleanQuestions(answerData.Questions)
		}
	}
	if !isValidParse {
		// пробуем построчное извлечение
		questions, err := extractQuestionsFromBrokenJSON(response)
		if err == nil {
			questions, ok := validateAndCleanQuestions(questions)
			if ok {
				answerData.Questions = questions
			} else {
				// выбираем лучший вариант (больше вопросов)
				if len(questions) > len(answerData.Questions) {
					answerData.Questions = questions
				}
			}
		}
	}
	if len(answerData.Questions) == 0 {
		return aimodels.Vk1QuestionResult{}, errors.New("ошибка извлечения вопросов")
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
	err = json.Unmarshal([]byte(answer), &scoreResult)
	if err != nil {
		return surveyapimodels.VkStep9ScoreResult{}, err
	}
	return scoreResult, nil
}

func ParseVk1IntroOutroAIResponse(response string) (result aimodels.Vk1IntroResult, err error) {
	answer := extractAnswer(response)

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

// извлекает JSON из ответа модели
func extractAnswer(response string) string {
	// Удаляем всё до </think>
	if idx := strings.Index(response, "</think>"); idx != -1 {
		response = response[idx+len("</think>"):]
	}
	response = strings.TrimSpace(response)

	// Пытаемся найти первый блок ```json ... ```
	jsonBlock := extractFirstJSONBlock(response)
	if jsonBlock != "" {
		jsonBlock = cleanTrailingCharacters(jsonBlock)
		jsonBlock = sanitizeJSON(jsonBlock)
		var data any
		err := json.Unmarshal([]byte(jsonBlock), &data)
		if err == nil {
			return jsonBlock
		}
	}

	// Fallback: ищем JSON по скобкам
	jsonStr := extractJSONByBraces(response)
	jsonStr = cleanTrailingCharacters(jsonStr)
	jsonStr = sanitizeJSON(jsonStr)
	var data any
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err == nil {
		return jsonStr
	}
	return ""
}

// находит первый блок ```json ... ```, содержащий JSON
func extractFirstJSONBlock(s string) string {
	re := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(.*?)\\n?```")
	matches := re.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if len(match) > 1 {
			content := strings.TrimSpace(match[1])
			if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
				return content
			}
		}
	}
	return ""
}

// находит JSON по первому '{' или '[' с учётом вложенности
func extractJSONByBraces(s string) string {
	start := strings.IndexAny(s, "{[")
	if start == -1 {
		return ""
	}
	stack := 0
	end := -1
	for i := start; i < len(s); i++ {
		if s[i] == '{' || s[i] == '[' {
			stack++
		} else if s[i] == '}' || s[i] == ']' {
			stack--
			if stack == 0 {
				end = i + 1
				break
			}
		}
	}
	if end == -1 {
		return ""
	}
	return s[start:end]
}

// удаляет лишние символы после последней закрывающей скобки
func cleanTrailingCharacters(jsonCandidate string) string {
	jsonCandidate = strings.TrimSpace(jsonCandidate)
	if jsonCandidate == "" {
		return ""
	}
	if strings.HasPrefix(jsonCandidate, "{") {
		lastClosing := strings.LastIndex(jsonCandidate, "}")
		if lastClosing != -1 {
			return jsonCandidate[:lastClosing+1]
		}
	} else if strings.HasPrefix(jsonCandidate, "[") {
		lastClosing := strings.LastIndex(jsonCandidate, "]")
		if lastClosing != -1 {
			return jsonCandidate[:lastClosing+1]
		}
	}
	return jsonCandidate
}

// исправляет частые ошибки в JSON
func sanitizeJSON(s string) string {
	// Умные кавычки → обычные
	smartQuotes := []string{"“", "”", "„", "«", "»", "’", "‘", "′", "″"}
	for _, q := range smartQuotes {
		s = strings.ReplaceAll(s, q, "\"")
	}

	// Множественные кавычки перед ключами: """text": → "text":
	reMultipleQuotes := regexp.MustCompile(`"+(\w+)"\s*:`)
	s = reMultipleQuotes.ReplaceAllString(s, `"$1":`)

	// Ключи без кавычек после { или ,
	reUnquotedKey1 := regexp.MustCompile(`([{,]\s*)(\w+)\s*:`)
	s = reUnquotedKey1.ReplaceAllString(s, `$1"$2":`)

	// Ключи без кавычек после перевода строки
	reUnquotedKey2 := regexp.MustCompile(`(\n\s*)(\w+)\s*:`)
	s = reUnquotedKey2.ReplaceAllString(s, `$1"$2":`)

	// Строковые значения без кавычек
	reUnquotedStringValue := regexp.MustCompile(`:\s*([a-zA-Zа-яА-ЯёЁ][a-zA-Zа-яА-ЯёЁ0-9_\-]*)\s*([,}\]])`)
	s = reUnquotedStringValue.ReplaceAllString(s, `:"$1"$2`)

	// Удаление запятых перед } или ]
	reTrailingComma := regexp.MustCompile(`,(\s*[}\]])`)
	s = reTrailingComma.ReplaceAllString(s, `$1`)

	// Добавление пропущенных запятых между полями объекта
	reMissingComma := regexp.MustCompile(`("\w+"\s*:\s*(?:"[^"]*"|\{[^}]*\}|\[[^\]]*\]|\d+|\w+))\s*\n?\s*("\w+"\s*:)`)
	s = reMissingComma.ReplaceAllString(s, `$1, $2`)

	return strings.TrimSpace(s)
}

func extractQuestionsFromBrokenJSON(response string) ([]questionFormat, error) {

	// Предварительная очистка от умных кавычек и управляющих символов
	cleaned := sanitizeForExtraction(response)

	// Ищем все блоки, похожие на объекты вопросов (содержат "text" и "comment")
	//    Шаблон: между { и } с возможными вложенными скобками (но для простоты ищем невложенные)
	objectPattern := regexp.MustCompile(`(?s)\{([^{}]*)\}`)
	objects := objectPattern.FindAllStringSubmatch(cleaned, -1)

	var questions []questionFormat
	idCounter := 1

	for _, obj := range objects {
		if len(obj) < 2 {
			continue
		}
		content := obj[1] // содержимое без внешних {}

		// Извлекаем текст вопроса
		text := extractFieldValue(content, "text")
		comment := extractFieldValue(content, "comment")

		if text == "" && comment == "" {
			continue
		}
		// Если нет текста вопроса, пропускаем (вопрос без текста не нужен)
		if text == "" {
			continue
		}

		questions = append(questions, questionFormat{
			ID:      fmt.Sprintf("q%d", idCounter),
			Text:    text,
			Comment: comment,
		})
		idCounter++
	}

	// Если ничего не нашли, пробуем более слабый поиск по строкам без фигурных скобок
	if len(questions) == 0 {
		questions = extractByLineScan(cleaned)
	}

	return questions, nil
}

// fallback для поиска вопросов построчно без фигурных скобок
func extractByLineScan(cleaned string) []questionFormat {
	var questions []questionFormat
	lines := strings.Split(cleaned, "\n")
	var currentText, currentComment strings.Builder
	inText := false
	inComment := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "text") && strings.Contains(line, ":") {
			// Начинаем новый вопрос
			if currentText.Len() > 0 {
				// Сохраняем предыдущий
				questions = append(questions, questionFormat{
					ID:      fmt.Sprintf("q%d", len(questions)+1),
					Text:    cleanString(currentText.String()),
					Comment: cleanString(currentComment.String()),
				})
				currentText.Reset()
				currentComment.Reset()
			}
			// Извлекаем текст из этой строки
			val := extractFieldValue(line, "text")
			if val != "" {
				currentText.WriteString(val)
				inText = true
				inComment = false
			}
		} else if strings.Contains(line, "comment") && strings.Contains(line, ":") {
			val := extractFieldValue(line, "comment")
			if val != "" {
				currentComment.WriteString(val)
				inComment = true
				inText = false
			}
		} else if inText && !strings.Contains(line, "comment") {
			// Продолжение текста вопроса на следующих строках
			currentText.WriteString(" " + line)
		} else if inComment {
			currentComment.WriteString(" " + line)
		}
	}
	// Добавляем последний вопрос
	if currentText.Len() > 0 {
		questions = append(questions, questionFormat{
			ID:      fmt.Sprintf("q%d", len(questions)+1),
			Text:    cleanString(currentText.String()),
			Comment: cleanString(currentComment.String()),
		})
	}
	return questions
}

// ищет в строке поле вида "ключ": "значение" (с учётом различных кавычек и без кавычек)
func extractFieldValue(content, key string) string {
	// Паттерн: возможны пробелы, затем ключ (с кавычками или без), затем :, затем значение до следующей запятой или конца объекта
	// Учитываем, что ключ может быть с умными кавычками или без кавычек, и регистр может быть разный
	// Делаем нечувствительным к регистру
	pattern := regexp.MustCompile(`(?i)(?:"?` + regexp.QuoteMeta(key) + `"?)\s*:\s*(?:"([^"\\]*(?:\\.[^"\\]*)*)"|([^,}\n]+))`)
	matches := pattern.FindStringSubmatch(content)
	if len(matches) == 0 {
		return ""
	}
	// matches[1] — значение в двойных кавычках, matches[2] — значение без кавычек (до запятой/скобки)
	rawVal := ""
	if matches[1] != "" {
		rawVal = matches[1]
	} else if len(matches) > 2 {
		rawVal = matches[2]
	}
	if rawVal == "" {
		return ""
	}
	// Очистка от лишних символов: кавычек, пробелов, управляющих последовательностей
	cleaned := cleanString(rawVal)
	return cleaned
}

// убирает обрамляющие кавычки, лишние пробелы, экранированные символы
func cleanString(s string) string {
	s = strings.TrimSpace(s)
	// Если строка окружена двойными или одинарными кавычками — удаляем их
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		s = s[1 : len(s)-1]
	}
	// Убираем экранирование кавычек внутри (простое)
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\\`, `\`)
	// Удаляем непечатные символы, оставляя пробелы
	var result strings.Builder
	for _, r := range s {
		if r >= 32 || r == '\n' || r == '\t' {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

// заменяет умные кавычки на обычные и удаляет непечатаемые символы
func sanitizeForExtraction(s string) string {
	// Умные кавычки
	smart := map[string]string{
		"“": "\"", "”": "\"", "„": "\"", "«": "\"", "»": "\"",
		"’": "'", "‘": "'",
		"′": "'", "″": "\"",
	}
	for old, new := range smart {
		s = strings.ReplaceAll(s, old, new)
	}
	// Удаляем управляющие символы, кроме пробелов и переносов строк
	var result strings.Builder
	for _, r := range s {
		if r >= 32 || r == '\n' || r == '\r' || r == '\t' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

type questionFormat struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Comment string `json:"comment"`
}

func validateAndCleanQuestions(questions []questionFormat) ([]questionFormat, bool) {
	idCounter := 1
	result := []questionFormat{}
	for _, question := range questions {
		if strings.TrimSpace(question.Text) == "" {
			continue
		}
		result = append(result, questionFormat{
			ID:      fmt.Sprintf("q%d", idCounter),
			Text:    question.Text,
			Comment: question.Comment,
		})
		if idCounter == 15 {
			break
		}
		idCounter++
	}
	return result, len(result) == 15
}