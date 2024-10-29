package yagptclient

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	yandexgptclient "github.com/sheeiavellie/go-yandexgpt"
)

type Provider interface {
	GenerateVacancyDescription(text string) (description string, err error)
}

type impl struct {
	client    *yandexgptclient.YandexGPTClient
	catalogID string
	promt     string
}

func NewClient(token, catalog, promt string) Provider {
	return impl{
		client:    yandexgptclient.NewYandexGPTClientWithIAMToken(token),
		catalogID: catalog,
		promt:     promt,
	}
}

func (i impl) GenerateVacancyDescription(text string) (description string, err error) {
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
				Text: i.promt,
			},
			{
				Role: yandexgptclient.YandexGPTMessageRoleUser,
				Text: fmt.Sprintf("Сгенерируй описание для вакансии имея эти вводные данные: %s", text),
			},
		},
	}

	response, err := i.client.CreateRequest(context.Background(), request)
	if err != nil {
		return "", errors.Wrap(err, "Ошибка при отправке запроса в API YandexGPT")
	}
	return response.Result.Alternatives[0].Message.Text, nil
}
