package whatsup

import (
	"context"
	whatsappclient "hr-tools-backend/lib/whatsup/client"

	"github.com/piusalfred/whatsapp/message"
)

type Provider interface {
	SendWelcome(ctx context.Context, recipient string) error
	SendAgreement(ctx context.Context, recipient string) error
	SendMsgWithFreeText(ctx context.Context, recipient, msg string) error
	SendMsgWithSelection(ctx context.Context, recipient, msg string, buttons []*message.InteractiveReplyButton) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		client:            whatsappclient.Instance,
		interactiveHeader: &message.InteractiveHeader{Text: "Whatsapp Cloud API Client", Type: "text"},
	}
}

type impl struct {
	client            whatsappclient.Provider
	interactiveHeader *message.InteractiveHeader
}

const (
	welcomeMsgBody   = "Здравствуйте! Вы откликнулись на вакансию «Менеджер по продажам в IT». Хотите пройти тест? Нажмите кнопку 'Да' для начала."
	agreementMsgBody = "Пожалуйста, подтвердите согласие на обработку ваших персональных данных для участия в процессе подбора. Нажмите 'Согласен' для подтверждения."
)

func (i impl) SendWelcome(ctx context.Context, recipient string) error {
	params := i.getRequest(welcomeMsgBody)
	params.Buttons = []*message.InteractiveReplyButton{
		{
			ID:    "submit_test",
			Title: "Да",
		},
		{
			ID:    "decline_test",
			Title: "Нет",
		},
	}
	return i.client.SendInteractiveMessage(ctx, recipient, params)
}

func (i impl) SendAgreement(ctx context.Context, recipient string) error {
	params := i.getRequest(agreementMsgBody)
	params.Buttons = []*message.InteractiveReplyButton{
		{
			ID:    "consent_granted",
			Title: "Согласен",
		},
		{
			ID:    "consent_declined",
			Title: "Не согласен",
		},
	}
	return i.client.SendInteractiveMessage(ctx, recipient, params)
}

func (i impl) SendMsgWithFreeText(ctx context.Context, recipient, msg string) error {
	return i.client.SendTextMessage(ctx, recipient, msg)
}

func (i impl) SendMsgWithSelection(ctx context.Context, recipient, msg string, buttons []*message.InteractiveReplyButton) error {
	params := i.getRequest(msg)
	params.Buttons = buttons
	return i.client.SendInteractiveMessage(ctx, recipient, params)
}

func (i impl) getRequest(body string) message.InteractiveReplyButtonsRequest {
	return message.InteractiveReplyButtonsRequest{
		Body:   body,
		Header: i.interactiveHeader,
	}
}
