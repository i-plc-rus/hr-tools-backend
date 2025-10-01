package models

type SpaceSettingCode string

const (
	YandexGPTPromtSetting    SpaceSettingCode = "ya_gpt_promt" // Инструкции для Yandex GPT при генерации описания вакансии
	HhClientIDSetting        SpaceSettingCode = "HHClientID"
	HhClientSecretSetting    SpaceSettingCode = "HHClientSecret"
	AvitoClientIDSetting     SpaceSettingCode = "AvitoClientID"
	AvitoClientSecretSetting SpaceSettingCode = "AvitoClientSecret"
	SpaceSenderEmail         SpaceSettingCode = "SpaceSenderEmail"  // почта, с которой отправляются письма кандидатам
	SpaceSupportEmail        SpaceSettingCode = "SpaceSupportEmail" // почта, тех поддержки
)
