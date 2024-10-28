package models

type SpaceSettingCode string

const (
	YandexGPTPromtSetting SpaceSettingCode = "ya_gpt_promt" // Инструкции для Yandex GPT при генерации описания вакансии
	HhClientIDSetting     SpaceSettingCode = "HHClientID"
	HhClientSecretSetting SpaceSettingCode = "HHClientSecret"
)
