package spaceapimodels

import "hr-tools-backend/models"

type PushSettings struct {
	IsActive bool            `json:"is_active"` // Push-уведомления включены
	Settings []PushSettingView `json:"settings"` // Список событий
}

type PushSettingData struct {
	Code  models.SpacePushSettingCode `json:"code"`  // Код события
	Value PushSettingValue            `json:"value"` // Значение настроек пуша
}

type PushSettingView struct {
	PushSettingData
	Name string `json:"name"` // Название события
}

type PushSettingValue struct {
	System *bool `json:"system,omitempty"` // Системные уведомления о событии (вкл/выкл)
	Email  *bool `json:"email,omitempty"`  // email уведомления о событии (вкл/выкл)
	Tg     *bool `json:"tg,omitempty"`     // telegram уведомления о событии (вкл/выкл)
}
