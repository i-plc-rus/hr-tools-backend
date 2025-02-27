package dbmodels

import (
	"database/sql/driver"
	"encoding/json"
)

type HRSurvey struct {
	BaseSpaceModel
	VacancyID   string            `gorm:"type:varchar(36);index"`
	Survey      HRSurveyQuestions `gorm:"type:jsonb"`
	IsFilledOut bool              // Анкета заполнена и может использоваться для оценки
}

func (j HRSurveyQuestions) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *HRSurveyQuestions) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

// Настройка анкеты
type HRSurveyQuestions struct {
	Questions []HRSurveyQuestion `json:"questions"`
}

type HRSurveyQuestionGenerated struct {
	QuestionID   string          `json:"question_id"`   // Идентификатор вопроса
	QuestionText string          `json:"question_text"` // Текст вопроса
	QuestionType string          `json:"question_type"` // Тип вопроса
	Answers      []SurveyAnswers `json:"answers"`       // Варианты ответов
	Comment      string          `json:"comment"`       // Комментарий
}

type HRSurveyQuestion struct {
	HRSurveyQuestionGenerated
	Weight   int    `json:"weight,omitempty"`   // Вес вопроса, заполняется автоматически
	Selected string `json:"selected,omitempty"` // Выбранный ответ
}

type SurveyAnswers struct {
	Value string `json:"value"`
}
