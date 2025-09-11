package dbmodels

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

type StepStatus int

const (
	VkStep0NotSent   = 0  //"Шаг0. Вопросы не отправлены"
	VkStep0Sent      = 10 //"Шаг0. Вопросы отправлены"
	VkStep0Answered  = 20 //"Шаг0. Получены ответы"
	VkStep0Refuse    = 30 //"Шаг0. Кандидат не прошел"
	VkStep0Done      = 40 //"Шаг0. Кандидат прошел"
	VkStep1Draft     = 50 //"Шаг1. Получен черновика скрипта"
	VkStep1DraftFail = 60 //"Шаг1. ошибка понучения черновика скрипта"
)

type ApplicantVkStep struct {
	BaseSpaceModel
	ApplicantID string `gorm:"type:varchar(36);index"`
	Status      StepStatus
	Step0       VkStep0 `gorm:"type:jsonb"`
	Step1       VkStep1 `gorm:"type:jsonb"`
}

func (j VkStep0) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *VkStep0) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

func (j *VkStep0) AnswerContent() (string, error) {
	body, err := json.Marshal(j)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации структуры ответов шага 0")
	}
	return string(body), nil
}

type VkStep0 struct {
	Answers []VkStep0Answer `json:"answers"`
}

type VkStep0Answer struct {
	ID     string `json:"id"`     // Идентификатор вопроса
	Answer string `json:"answer"` // Ответ кардидата
}

func (j VkStep1) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *VkStep1) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

type VkStep1 struct {
	Questions   []VkStep1Question `json:"questions"`
	ScriptIntro string            `json:"script_intro"`
	ScriptOutro string            `json:"script_outro"`
	Comments    map[string]string `json:"comments"`
}

type VkStep1Question struct {
	ID      string   `json:"id"`      // Идентификатор вопроса
	Text    string   `json:"text"`    // Текст вопроса
	Type    string   `json:"type"`    // Тип вопроса
	Options []string `json:"options"` // Варианты ответов
}
