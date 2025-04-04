package yagptclient

import (
	"context"
	"github.com/pkg/errors"
	yandexgptclient "github.com/sheeiavellie/go-yandexgpt"
)

type Provider interface {
	GenerateByPromtAndText(promt, text string) (generatedText string, err error)
}

type impl struct {
	client    *yandexgptclient.YandexGPTClient
	catalogID string
}

func NewClient(token, catalog string) Provider {
	return impl{
		client:    yandexgptclient.NewYandexGPTClientWithIAMToken(token),
		catalogID: catalog,
	}
}

func (i impl) GenerateByPromtAndText(promt, text string) (description string, err error) {
	request := yandexgptclient.YandexGPTRequest{
		ModelURI: yandexgptclient.MakeModelURI(i.catalogID, yandexgptclient.YandexGPTModelLite),
		CompletionOptions: yandexgptclient.YandexGPTCompletionOptions{
			Stream:      false,
			Temperature: 0.3,
			MaxTokens:   2000,
		},
		Messages: []yandexgptclient.YandexGPTMessage{
			{
				Role: yandexgptclient.YandexGPTMessageRoleSystem,
				Text: promt,
			},
			{
				Role: yandexgptclient.YandexGPTMessageRoleUser,
				Text: text,
			},
		},
	}

	response, err := i.client.CreateRequest(context.Background(), request)
	if err != nil {
		return "", errors.Wrap(err, "Ошибка при отправке запроса на генерацию в API YandexGPT")
	}
	return response.Result.Alternatives[0].Message.Text, nil
}
