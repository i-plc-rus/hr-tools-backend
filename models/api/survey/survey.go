package surveyapimodels

import (
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
)

type Survey struct {
	Questions   []dbmodels.SurveyQuestion `json:"questions"`
}

type SurveyView struct {
	Survey
	IsFilledOut bool `json:"is_filled_out"` // "анкета полностью заполнена"
}

func (s Survey) Validate() error {
	if len(s.Questions) == 0 {
		return errors.New("в анкете отсутсвуют вопросы")
	}
	for _, question := range s.Questions {
		if question.QuestionID == "" {
			return errors.New("в одном из вопросов анкеты отсутсвует идентификатор вопроса")
		}
		if question.Selected != "Не подходит" {
			if question.QuestionText == "" {
				return errors.New("в одном из вопросов анкеты отсутсвует текст вопроса, для перегенерации выберите вариант \"Не подходит\"")
			}
			if question.QuestionType == "" {
				return errors.New("в одном из вопросов анкеты отсутсвует тип вопроса, для перегенерации выберите вариант \"Не подходит\"")
			}
		}
	}

	return nil
}

type ReGenerateSurvey struct {
	Questions []NotSuitableQuestion `json:"questions"`
}

type NotSuitableQuestion struct {
	QuestionID   string `json:"question_id"`
	QuestionText string `json:"question_text"`
	QuestionType string `json:"question_type"`
}
