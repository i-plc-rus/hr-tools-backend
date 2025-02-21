package whatsup

import (
	"context"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	whatsappclient "hr-tools-backend/lib/whatsup/client"
	"hr-tools-backend/models"

	"github.com/piusalfred/whatsapp/message"
	webhookmessage "github.com/piusalfred/whatsapp/webhooks/message"
	"github.com/pkg/errors"
)

type Provider interface {
	SendWelcome(ctx context.Context, spaceID, recipient string) error
	SendAgreement(ctx context.Context, spaceID, recipient string) error
	SendMsgWithFreeText(ctx context.Context, spaceID, recipient, msg string) error
	SendMsgWithSelection(ctx context.Context, spaceID, recipient, msg string, buttons []*message.InteractiveReplyButton) error
	HandleWebHook(ctx context.Context, entity *webhookmessage.Entry)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		interactiveHeader:  &message.InteractiveHeader{Text: "Whatsapp Cloud API Client", Type: "text"},
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
	}
}

type impl struct {
	interactiveHeader  *message.InteractiveHeader
	spaceSettingsStore spacesettingsstore.Provider
}

const (
	welcomeMsgBody   = "Здравствуйте! Вы откликнулись на вакансию «Менеджер по продажам в IT». Хотите пройти тест? Нажмите кнопку 'Да' для начала."
	agreementMsgBody = "Пожалуйста, подтвердите согласие на обработку ваших персональных данных для участия в процессе подбора. Нажмите 'Согласен' для подтверждения."
)

func (i impl) getClient(spaceID string) (whatsappclient.Provider, error) {
	accessToken, err := i.spaceSettingsStore.GetValueByCode(spaceID, models.WhatsAppAccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения токена WhatsApp")
	}
	if accessToken == "" {
		return nil, errors.New("WhatsApp токен не указан")
	}
	businessAccountID, err := i.spaceSettingsStore.GetValueByCode(spaceID, models.WhatsAppBusinessAccountID)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения business account id WhatsApp")
	}
	if businessAccountID == "" {
		return nil, errors.New("WhatsApp business account id не указан")
	}
	return whatsappclient.GetClient(config.Conf.WhatsUpp.BaseUrl, accessToken, config.Conf.WhatsUpp.APIVersion, businessAccountID)
}

func (i impl) SendWelcome(ctx context.Context, spaceID, recipient string) error {
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
	client, err := i.getClient(spaceID)
	if err != nil {
		return err
	}
	return client.SendInteractiveMessage(ctx, recipient, params)
}

func (i impl) SendAgreement(ctx context.Context, spaceID, recipient string) error {
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
	client, err := i.getClient(spaceID)
	if err != nil {
		return err
	}
	return client.SendInteractiveMessage(ctx, recipient, params)
}

func (i impl) SendMsgWithFreeText(ctx context.Context, spaceID, recipient, msg string) error {
	client, err := i.getClient(spaceID)
	if err != nil {
		return err
	}
	return client.SendTextMessage(ctx, recipient, msg)
}

func (i impl) SendMsgWithSelection(ctx context.Context, spaceID, recipient, msg string, buttons []*message.InteractiveReplyButton) error {
	params := i.getRequest(msg)
	params.Buttons = buttons
	client, err := i.getClient(spaceID)
	if err != nil {
		return err
	}
	return client.SendInteractiveMessage(ctx, recipient, params)
}

func (i impl) HandleWebHook(ctx context.Context, entity *webhookmessage.Entry) {
	//TODO Получение ответов кандидата, сохранение, обработка
	//entity.Changes[0].Value.Messages[0].Interactive.ButtonReply
}

func (i impl) getRequest(body string) message.InteractiveReplyButtonsRequest {
	return message.InteractiveReplyButtonsRequest{
		Body:   body,
		Header: i.interactiveHeader,
	}
}
