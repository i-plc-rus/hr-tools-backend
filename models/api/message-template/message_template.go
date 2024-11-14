package msgtemplateapimodels

import (
	"github.com/pkg/errors"
	"strings"
)

type MsgTemplateView struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type SendMessage struct {
	ApplicantID   string `json:"applicant_id"`    // ID кандидата/отклика кому отправить сообщение
	MsgTemplateID string `json:"msg_template_id"` // ID шаблона сообщения, которое нужно отправить
}

func (r SendMessage) Validate() error {
	if len(strings.TrimSpace(r.ApplicantID)) == 0 {
		return errors.New("не указан кандидат")
	}
	if len(strings.TrimSpace(r.MsgTemplateID)) == 0 {
		return errors.New("не указан шаблон сообщения")
	}
	return nil
}
