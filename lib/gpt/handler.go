package gpthandler

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	yagptclient "hr-tools-backend/lib/gpt/yagpt-client"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	"hr-tools-backend/models"
	gptmodels "hr-tools-backend/models/api/gpt"
)

type Provider interface {
	GenerateVacancyDescription(spaceID, text string) (resp gptmodels.GenVacancyDescResponse, err error)
}

type impl struct {
	spaceSettingsStore spacesettingsstore.Provider
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
	}
}

func (i impl) GenerateVacancyDescription(spaceID, text string) (resp gptmodels.GenVacancyDescResponse, err error) {
	promt, err := i.spaceSettingsStore.GetValueByCode(spaceID, models.YandexGPTPromtSetting)
	if err != nil {
		log.
			WithField("space_id", spaceID).
			WithError(err).
			Error("ошибка получения инструкции для YandexGPT из настройки space")
		return resp, err
	}
	if promt == "" {
		log.
			WithField("space_id", spaceID).
			Warn("инструкция для YandexGPT из настройки space не должна быть пустой")
		return resp, errors.New("инструкция для YandexGPT из настройки space не должна быть пустой")
	}
	//promt := "Ты - рекрутер компании Рога и Копыта. В компании придерживаемся свободного стиля, используем эмодзи в текстах вакансии"
	resp.Description, err = yagptclient.
		NewClient(config.Conf.YandexGPT.IAMToken, config.Conf.YandexGPT.CatalogID).
		GenerateByPromtAndText(promt, fmt.Sprintf("Сгенерируй описание для вакансии имея эти вводные данные: %s", text))
	if err != nil {
		log.
			WithField("space_id", spaceID).
			WithError(err).
			Error("ошибка генерации описания через YandexGPT")
		return resp, err
	}
	return resp, nil
}
