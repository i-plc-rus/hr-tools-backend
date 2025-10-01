package surveyapimodels

import (
	"hr-tools-backend/config"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
)

type VkAiProvider interface {
	VkStep1(spaceID, vacancyID string, aiData AiData) (resp VkStep1, err error)
	VkStep1Regen(spaceID, vacancyID string, aiData AiData) (newQuestions []VkStep1Question, comments map[string]string, err error)
}

type VkStep1 struct {
	Questions   []VkStep1Question `json:"questions"`
	ScriptIntro string            `json:"script_intro"`
	ScriptOutro string            `json:"script_outro"`
	Comments    map[string]string `json:"comments"`
}

type VkStep1Update struct {
	VkStep1
	Approve bool `json:"approve"`
}

type VkStep1View struct {
	VkStep1
	Url              string `json:"url"`                // Ссылка на анкету для видео интервью
	DateOfInvitation string `json:"date_of_invitation"` // Дата отправки приглашения
}

func (r VkStep1Update) Validate() error {
	if len(r.Questions) < 15 {
		return errors.New("Необходимо указать 15 вопросов")
	}
	if r.ScriptIntro == "" {
		return errors.New("отсутсвует текст сценария для intro")
	}
	if r.ScriptOutro == "" {
		return errors.New("отсутсвует текст сценария для outro")
	}
	return nil
}

type VkStep1Regen struct {
	Questions []VkStep1RegenQuestion `json:"questions"`
}

func (r VkStep1Regen) Validate() error {
	if len(r.Questions) == 0 {
		return errors.New("Необходимо указать хотябы один вопрос для повторной генерации")
	}
	for _, q := range r.Questions {
		if q.NotSuitable {
			return nil
		}
	}
	return errors.New("Для перегенерации необходимо отметить не подходящие вопросы")
}

type VkStep1Question struct {
	ID    string `json:"id"`   // Идентификатор вопроса
	Text  string `json:"text"` // Текст вопроса
	Order int    `json:"order"`
}

type VkStep1RegenQuestion struct {
	VkStep1Question
	NotSuitable       bool   `json:"not_suitable"`        // вопрос не подходит необходима пергенерация
	NotSuitableReason string `json:"not_suitable_reason"` // причина по которой вопрос не подходит
}

type VkStep1QuestionUpdate struct {
	ID    string `json:"id"`    // Идентификатор вопроса
	Text  string `json:"text"`  // Текст вопроса
	Order int    `json:"order"` // Порядковый номер
}

func VkStep1Convert(rec dbmodels.ApplicantVkStep) VkStep1View {
	result := VkStep1View{
		VkStep1: VkStep1{
			Questions:   []VkStep1Question{},
			ScriptIntro: rec.Step1.ScriptIntro,
			ScriptOutro: rec.Step1.ScriptOutro,
			Comments:    rec.Step1.Comments,
		},
		Url: rec.GetVideoSurveyUrl(config.Conf),
	}
	for _, question := range rec.Step1.Questions {
		result.Questions = append(result.Questions, VkStep1Question{
			ID:    question.ID,
			Text:  question.Text,
			Order: question.Order,
		})
	}

	if !rec.VideoInterviewInviteDate.IsZero() {
		result.DateOfInvitation = rec.VideoInterviewInviteDate.Format("02.01.2006 15:04:05")
	}
	return result
}
