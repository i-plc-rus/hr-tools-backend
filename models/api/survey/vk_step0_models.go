package surveyapimodels

import (
	"encoding/json"
	"hr-tools-backend/config"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
)

const (
	TypicalQuestion1 string = "Ваши ожидания по заработной плате?"
	TypicalQuestion2 string = "Готовы ли вы к командировкам?"
	TypicalQuestion3 string = "Имеется ли у вас опыт работы?"
	TypicalQuestion4 string = "Как у вас проходит адаптация к новомо коллективу?"
	TypicalQuestion5 string = "Ваше семейное положение?"
)

var (
	Question2Answers = []string{"да", "нет"}
	Question3Answers = []string{"да", "нет"}
	Question5Answers = []string{"свободен", "женат/замужем"}
)

var QuestionsStep0 = VkStep0SurveyView{
	Questions: []VkStep0Question{
		{
			QuestionID:   "1",
			QuestionText: TypicalQuestion1,
		},
		{
			QuestionID:   "2",
			QuestionText: TypicalQuestion2,
			Answers:      Question2Answers,
		},
		{
			QuestionID:   "3",
			QuestionText: TypicalQuestion3,
			Answers:      Question3Answers,
		},
		{
			QuestionID:   "4",
			QuestionText: TypicalQuestion4,
		},
		{
			QuestionID:   "5",
			QuestionText: TypicalQuestion5,
			Answers:      Question5Answers,
		},
		// TODO Добавить типовые вопросы
	},
}

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

func VkStep0Convert(rec dbmodels.ApplicantVkStep) VkStep0 {
	result := VkStep0{
		Url:       config.Conf.UIParams.SurveyStep0Path + rec.ID,
		Questions: QuestionsStep0.Questions,
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
