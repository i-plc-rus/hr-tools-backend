package negotiationapimodels

import (
	"time"

	"github.com/pkg/errors"
)

type MessengerAvailableRequest struct {
	ApplicantID string `json:"applicant_id"`
}

func (m MessengerAvailableRequest) Validate() error {
	if m.ApplicantID == "" {
		return errors.New("не указан идентификатор кандидата")
	}
	return nil
}

type MessengerAvailableResponse struct {
	IsAvailable bool   `json:"is_available"`
	Service     string `json:"service"` //Avito/HeadHunter
}

type NewMessageRequest struct {
	ApplicantID string `json:"applicant_id"`
	Text        string `json:"text"`
}

func (m NewMessageRequest) Validate() error {
	if m.ApplicantID == "" {
		return errors.New("не указан идентификатор кандидата")
	}
	if m.Text == "" {
		return errors.New("не указано сообщение для отправки")
	}
	return nil
}

type MessageListRequest struct {
	ApplicantID string `json:"applicant_id"`
}

func (m MessageListRequest) Validate() error {
	if m.ApplicantID == "" {
		return errors.New("не указан идентификатор кандидата")
	}
	return nil
}

type MessageItem struct {
	ID              string    `json:"id"`
	MessageDateTime time.Time `json:"message_date_time"` // Дата/время сообщения
	SelfMessage     bool      `json:"self_message"`      // Сообщение от true - работодателя / false - кандидата
	Text            string    `json:"text"`              // Текст сообщения
	AuthorFullName  string    `json:"author_full_name"`  // ФИО автора
}
