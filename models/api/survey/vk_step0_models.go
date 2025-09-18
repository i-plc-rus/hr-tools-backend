package surveyapimodels

import (
	"encoding/json"
	"fmt"
	"hr-tools-backend/config"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"slices"
	"strconv"

	"github.com/pkg/errors"
)

func GetQuestionsStep0(jobTitle string) VkStep0SurveyView {
	questions := VkStep0SurveyView{
		Questions: []VkStep0Question{
			{
				QuestionID:   "1",
				QuestionText: config.Conf.Survey.VkStep0.Question1,
				QuestionType: "single_choice",
				Answers:      answerStep0Q1,
			},
			{
				QuestionID:   "2",
				QuestionText: config.Conf.Survey.VkStep0.Question2,
				QuestionType: "free_text",
			},
			{
				QuestionID:   "3",
				QuestionText: config.Conf.Survey.VkStep0.Question3,
				Answers:      models.EmploymentSlice(),
				QuestionType: "single_choice",
			},
			{
				QuestionID:   "4",
				QuestionText: config.Conf.Survey.VkStep0.Question4,
				QuestionType: "single_choice",
				Answers:      models.ScheduleSlice(),
			},
			{
				QuestionID:   "5",
				QuestionText: config.Conf.Survey.VkStep0.Question5,
				QuestionType: "single_choice",
				Answers:      models.ExperienceSlice(),
			},
		},
	}
	questions.Questions[0].QuestionText = fmt.Sprintf(questions.Questions[0].QuestionText, jobTitle)
	return questions
}

var answerStep0Q1 = []string{"да", "нет"}

type VkStep0SurveyView struct {
	Questions []VkStep0Question `json:"questions"`
}

func (v VkStep0SurveyView) Content() (string, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации структуры вопросов шага 0")
	}
	return string(body), nil
}

type VkStep0Question struct {
	QuestionID   string   `json:"question_id"`   // Идентификатор вопроса
	QuestionText string   `json:"question_text"` // Текст вопроса
	QuestionType string   `json:"question_type"` // Тип вопроса
	Answers      []string `json:"answers"`       // Варианты ответов
}

type VkStep0SurveyAnswers struct {
	Answers []VkStep0Answer `json:"answers"`
}

func (v VkStep0SurveyAnswers) Validate() error {
	validAnswerCount := 0
	for _, answer := range v.Answers {
		switch answer.QuestionID {
		case "1":
			if !slices.Contains(answerStep0Q1, answer.Answer) {
				return errors.New("Для вопроса #1 необходимо выбрать ответ из списка")
			}
			validAnswerCount++
		case "2":
			value, err := strconv.Atoi(answer.Answer)
			if err != nil || value <= 0 {
				return errors.New("Для вопроса #2 необходимо указать целое, положительное число")
			}
			validAnswerCount++
		case "3":
			if !slices.Contains(models.EmploymentSlice(), answer.Answer) {
				return errors.New("Для вопроса #3 необходимо выбрать ответ из списка")
			}
			validAnswerCount++
		case "4":
			if !slices.Contains(models.ScheduleSlice(), answer.Answer) {
				return errors.New("Для вопроса #4 необходимо выбрать ответ из списка")
			}
			validAnswerCount++
		case "5":
			if !slices.Contains(models.ExperienceSlice(), answer.Answer) {
				return errors.New("Для вопроса #5 необходимо выбрать ответ из списка")
			}
			validAnswerCount++
		}
	}
	if validAnswerCount < 5 {
		return errors.New("Необходимо ответить на все вопросы")
	}
	return nil
}

type VkStep0Answer struct {
	QuestionID string `json:"question_id"` // Идентификатор вопроса
	Answer     string `json:"answer"`      // Варианты ответов
}

type VkStep0SurveyResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type VkStep0 struct {
	Url       string            // Ссылка на анкету c типовыми вопросами для кандидата
	Questions []VkStep0Question `json:"questions"`
	Answers   []VkStep0Answer   `json:"answers"`
}

func VkStep0Convert(rec dbmodels.ApplicantVkStep, jobTitle string) VkStep0 {
	result := VkStep0{
		Url:       config.Conf.UIParams.SurveyStep0Path + rec.ID,
		Questions: GetQuestionsStep0(jobTitle).Questions,
		Answers:   []VkStep0Answer{},
	}
	for _, answer := range rec.Step0.Answers {
		result.Answers = append(result.Answers, VkStep0Answer{
			QuestionID: answer.ID,
			Answer:     answer.Answer,
		})
	}
	return result
}
