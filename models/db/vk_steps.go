package dbmodels

import (
	"database/sql/driver"
	"encoding/json"
	"hr-tools-backend/config"
	"time"

	"github.com/pkg/errors"
)

type StepStatus int

const (
	VkStep0NotSent               = 0   //"Шаг0. Вопросы не отправлены"
	VkStep0Sent                  = 10  //"Шаг0. Вопросы отправлены"
	VkStep0Answered              = 20  //"Шаг0. Получены ответы"
	VkStep0Refuse                = 30  //"Шаг0. Кандидат не прошел"
	VkStep0Done                  = 40  //"Шаг0. Кандидат прошел"
	VkStep1Draft                 = 50  //"Шаг1. Получен черновика скрипта"
	VkStep1DraftFail             = 60  //"Шаг1. Ошибка получения черновика скрипта"
	VkStep1Regen                 = 70  //"Шаг1. Перегенерация"
	VkStep1Approved              = 80  //"Шаг1. Список вопросов подтвержден"
	VkStepVideoSuggestSent       = 90  //"Шаг7. Приглашение на видео интервью отправлено кандидату"
	VkStepVideoTranscripted      = 100 //"Шаг9. Транскрибация выполнена"
	VkStepVideoSemanticEvaluated = 110 //"Шаг9. Семантическая оценка расчитана"
)

func (s StepStatus) String() string {
	switch s {
	case VkStep0NotSent:
		return "Шаг0. Ссылка на анкету с типовыми вопросами не отправлена"
	case VkStep0Sent:
		return "Шаг0. Ссылка на анкету с типовыми вопросами отправлена кандидату"
	case VkStep0Answered:
		return "Шаг0. Кандидат ответил на анкету"
	case VkStep0Refuse:
		return "Шаг0. Кандидат не прошел отбор"
	case VkStep0Done:
		return "Шаг0. Кандидат прошел отбор"
	case VkStep1Draft:
		return "Шаг1. Получен черновика скрипта"
	case VkStep1DraftFail:
		return "Шаг1. Не удалось получить черновик скрипта"
	case VkStep1Regen:
		return "Шаг1. Перегенерация списка вопросов"
	case VkStep1Approved:
		return "Шаг1. Список вопросов подтвержден"
	case VkStepVideoSuggestSent:
		return "Шаг7. Приглашение на видео интервью отправлено кандидату"
	case VkStepVideoTranscripted:
		return "Шаг9. Транскрибация видео выполнена"
	case VkStepVideoSemanticEvaluated:
		return "Шаг9. Семантическая оценка ответов расчитана"
	default:
		return "Не известный статус"
	}
}

type ApplicantVkStep struct {
	BaseSpaceModel
	ApplicantID               string `gorm:"type:varchar(36);index"`
	Status                    StepStatus
	Step0                     VkStep0                  `gorm:"type:jsonb"`
	Step1                     VkStep1                  `gorm:"type:jsonb"` // вопросы для видео интервью
	VideoInterview            VideoInterview           `gorm:"type:jsonb"` // ссылки на файлы с видео ответами
	VideoInterviewInviteDate  time.Time                // время отправки ссылки на видео интервью
	VideoInterviewEvaluations []ApplicantVkVideoSurvey `gorm:"foreignKey:ApplicantVkStepID"` // Транскрибация и семантическая оценка ответов видео интервью
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
	ID                string `json:"id"`   // Идентификатор вопроса
	Text              string `json:"text"` // Текст вопроса
	Order             int    `json:"order"`
	NotSuitable       bool   `json:"not_suitable"`        // не подходит
	NotSuitableReason string `json:"not_suitable_reason"` // причина по которой вопрос не подходит
}

func (r ApplicantVkStep) GetStep0SurveyUrl(conf *config.Configuration) string {
	return conf.UIParams.SurveyStep0Path + r.ID
}

func (r ApplicantVkStep) GetVideoSurveyUrl(conf *config.Configuration) string {
	return conf.UIParams.VideoSurveyStepPath + r.ID
}

type VideoInterview struct {
	Answers map[string]VkVideoAnswer `json:"answers"` // map[questionID]VkVideoAnswer
}

type VkVideoAnswer struct {
	FileID string `json:"file_id"` // Идентификатор файла
}

func (j VideoInterview) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *VideoInterview) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

type ApplicantVkVideoSurvey struct {
	BaseModel
	ApplicantVkStepID    string
	QuestionID           string
	Error                string
	TranscriptText       string
	VoiceAmplitudeFileID string
	FramesFileID         string
	EmotionFileID        string
	SentimentFileID      string
	IsSemanticEvaluated  bool
	Similarity           int    // совпадение
	CommentForSimilarity string // краткий комментарий оценки
}
