package dbmodels

import (
	"database/sql/driver"
	"encoding/json"
)

type ApplicantSurvey struct {
	BaseSpaceModel
	VacancySurveyID string                   `gorm:"type:varchar(36);index"`
	ApplicantID     string                   `gorm:"type:varchar(36);index"`
	Survey          ApplicantSurveyQuestions `gorm:"type:jsonb"`
	IsFilledOut     bool                     // Анкета заполнена кандидатом и может использоваться для оценки
}

func (j ApplicantSurveyQuestions) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *ApplicantSurveyQuestions) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

type ApplicantSurveyQuestions struct {
	Questions []ApplicantSurveyQuestion `json:"questions"`
}

type ApplicantSurveyQuestion struct {
	QuestionID   string   `json:"question_id"`   // Идентификатор вопроса
	QuestionText string   `json:"question_text"` // Текст вопроса
	QuestionType string   `json:"question_type"` // Тип вопроса
	Answers      []string `json:"answers"`       // Варианты ответов
	Comment      string   `json:"comment"`       // Комментарий
	Weight       int      `json:"weight"`
	Selected     string   `json:"selected,omitempty"` // Ответ кандидата
}
