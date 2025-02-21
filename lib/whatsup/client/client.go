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

var Instance Provider

func Connect(ctx context.Context, baseUrl, accessToken, apiVersion, businessAccountID string) error {
	httpClient := &http.Client{}
	clientOptions := []whttp.CoreClientOption[message.Message]{
		whttp.WithCoreClientHTTPClient[message.Message](httpClient),
		whttp.WithCoreClientRequestInterceptor[message.Message](
			func(ctx context.Context, req *http.Request) error {
				// fmt.Println("Request Intercepted")
				return nil
			},
		),
		whttp.WithCoreClientResponseInterceptor[message.Message](
			func(ctx context.Context, resp *http.Response) error {
				// fmt.Println("Response Intercepted")
				return nil
			},
		),
	}
	Instance = impl{
		coreClient: whttp.NewSender[message.Message](clientOptions...),
		configReader: config.ReaderFunc(func(ctx context.Context) (*config.Config, error) {
			return &config.Config{
				BaseURL:           baseUrl,           // Replace with your API URL
				AccessToken:       accessToken,       // Replace with your access token
				APIVersion:        apiVersion,        // WhatsApp API version
				BusinessAccountID: businessAccountID, // Replace with your business account ID
			}, nil
		}),
	}
	return nil
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

	// Send the text message
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
