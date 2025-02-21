package whatsappclient

import (
	"context"
	"net/http"

	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/message"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	SendTextMessage(ctx context.Context, recipient, msg string) error
	SendInteractiveMessage(ctx context.Context, recipient string, params message.InteractiveReplyButtonsRequest) error
}

func GetClient(baseUrl, accessToken, apiVersion, businessAccountID string) (Provider, error) {
	httpClient := &http.Client{}
	clientOptions := []whttp.CoreClientOption[message.Message]{
		whttp.WithCoreClientHTTPClient[message.Message](httpClient),
	}
	return impl{
		coreClient: whttp.NewSender[message.Message](clientOptions...),
		configReader: config.ReaderFunc(func(ctx context.Context) (*config.Config, error) {
			return &config.Config{
				BaseURL:           baseUrl,           // API URL
				AccessToken:       accessToken,       // access token
				APIVersion:        apiVersion,        // PI version
				BusinessAccountID: businessAccountID, // business account ID
			}, nil
		}),
	}, nil
}

type impl struct {
	configReader config.ReaderFunc
	coreClient   *whttp.CoreClient[message.Message]
}

func (i impl) SendTextMessage(ctx context.Context, recipient, msg string) error {
	logger := log.WithField("recipient", recipient)

	client, err := message.NewBaseClient(i.coreClient, i.configReader)
	if err != nil {
		return errors.Wrap(err, "ошибка создания клиента")
	}

	// Define the recipient's WhatsApp phone number (including country code)
	// recipient := "1234567890"

	// Create a new text message request
	textMessage := message.NewRequest(recipient, &message.Text{
		Body: msg,
	}, "")

	// Send
	response, err := client.SendText(ctx, textMessage)
	if err != nil {
		return errors.Wrap(err, "ошибка отправки сообщения")
	}

	logger.Infof("Сообщение успешно отправтено %v. Response: %+v", recipient, response)
	return nil
}

func (i impl) SendInteractiveMessage(ctx context.Context, recipient string, params message.InteractiveReplyButtonsRequest) error {
	logger := log.WithField("recipient", recipient)

	client, err := message.NewBaseClient(i.coreClient, i.configReader)
	if err != nil {
		return errors.Wrap(err, "ошибка создания клиента")
	}

	buttons := make([]*message.InteractiveButton, 0, len(params.Buttons))
	for _, button := range params.Buttons {
		buttons = append(buttons, &message.InteractiveButton{
			Type:  message.InteractiveActionButtonReply,
			Reply: button,
		})
	}
	interactiveMessage := message.NewInteractiveMessageContent(
		message.TypeInteractiveButton,
		message.WithInteractiveAction(&message.InteractiveAction{
			Buttons: buttons,
		}),

		message.WithInteractiveBody(params.Body),
		message.WithInteractiveHeader(params.Header),
	)

	ir := message.NewRequest(recipient, interactiveMessage, "")
	response, err := client.SendInteractiveMessage(ctx, ir)
	if err != nil {
		return errors.Wrap(err, "ошибка отправки сообщения")
	}

	logger.Infof("Сообщение успешно отправтено %v. Response: %+v", recipient, response)
	return nil
}
