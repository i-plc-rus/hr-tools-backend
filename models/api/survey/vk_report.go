package surveyapimodels

import (
	"encoding/json"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
)

type ReportRequestData struct {
	VacancyInfo      string
	Requirements     string
	ApplicantInfo    string
	Questions        string
	ApplicantAnswers string
	Evalutions       string
	TotalScore       int
	Threshold        int
}

type QuestionFormat struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Comment string `json:"comment"`
}

type AnswerFormat struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type AnswerEvalutionFormat struct {
	ID         string `json:"id"`
	Similarity int    `json:"similarity"`
	Comment    string `json:"comment"`
}

type ReportResult struct {
	OverallComment string `json:"overall_comment"`
}

func GetInterviewQuestionsContent(rec dbmodels.ApplicantVkStep) (string, error) {
	questions := []QuestionFormat{}
	for _, q := range rec.Step1.Questions {
		question := QuestionFormat{
			ID:      q.ID,
			Text:    q.Text,
			Comment: rec.Step1.Comments[q.ID],
		}
		questions = append(questions, question)
	}
	body, err := json.Marshal(questions)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации структуры вопросов для промпта")
	}
	return string(body), nil
}

func GetInterviewAnswersContent(rec dbmodels.ApplicantVkStep) (string, error) {
	answers := []AnswerFormat{}
	for _, q := range rec.VideoInterviewEvaluations {
		answer := AnswerFormat{
			ID:   q.QuestionID,
			Text: q.TranscriptText,
		}
		answers = append(answers, answer)
	}
	body, err := json.Marshal(answers)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации структуры с ответами кандидата для промпта")
	}
	return string(body), nil
}

func GetInterviewEvalutionsContent(rec dbmodels.ApplicantVkStep) (string, error) {
	answers := []AnswerEvalutionFormat{}
	for _, q := range rec.VideoInterviewEvaluations {
		answer := AnswerEvalutionFormat{
			ID:         q.QuestionID,
			Similarity: q.Similarity,
			Comment:    q.CommentForSimilarity,
		}
		answers = append(answers, answer)
	}
	body, err := json.Marshal(answers)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации структуры с ответами кандидата для промпта")
	}
	return string(body), nil
}
