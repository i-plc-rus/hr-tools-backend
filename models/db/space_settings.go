package dbmodels

import (
	"hr-tools-backend/models"
	spaceapimodels "hr-tools-backend/models/api/space"
)

type SpaceSetting struct {
	BaseModel
	SpaceID string                  `gorm:"type:varchar(36);index:idx_setting_code"`
	Name    string                  `gorm:"type:varchar(255)"`
	Code    models.SpaceSettingCode `gorm:"type:varchar(255);index:idx_setting_code"`
	Value   string                  `gorm:"type:varchar(500)"`
}

func (r SpaceSetting) ToModelView() spaceapimodels.SpaceSettingView {
	return spaceapimodels.SpaceSettingView{
		ID:      r.ID,
		SpaceID: r.SpaceID,
		Name:    r.Name,
		Code:    r.Code,
		Value:   r.Value,
	}
}


var DefaultHhClientIDSetting = SpaceSetting{
	SpaceID: "",
	Name:    "client_id для HeadHunter",
	Code:    models.HhClientIDSetting,
	Value:   "",
}

var DefaultHhClientSecretSetting = SpaceSetting{
	SpaceID: "",
	Name:    "client_secret для HeadHunter",
	Code:    models.HhClientSecretSetting,
	Value:   "",
}

var DefaultAvitoClientIDSetting = SpaceSetting{
	SpaceID: "",
	Name:    "client_id для Avito",
	Code:    models.AvitoClientIDSetting,
	Value:   "",
}

var DefaultAvitoClientSecretSetting = SpaceSetting{
	SpaceID: "",
	Name:    "client_secret для Avito",
	Code:    models.AvitoClientSecretSetting,
	Value:   "",
}

var DefaultSpaceSenderEmail = SpaceSetting{
	SpaceID: "",
	Name:    "почта, с которой отправляются письма кандидатам",
	Code:    models.SpaceSenderEmail,
	Value:   "",
}

var DefaultSpaceSupportEmail = SpaceSetting{
	SpaceID: "",
	Name:    "почта, тех поддержки",
	Code:    models.SpaceSupportEmail,
	Value:   "",
}

var DefaultSettinsMap = map[models.SpaceSettingCode]SpaceSetting{
	models.HhClientIDSetting:        DefaultHhClientIDSetting,
	models.HhClientSecretSetting:    DefaultHhClientSecretSetting,
	models.AvitoClientIDSetting:     DefaultAvitoClientIDSetting,
	models.AvitoClientSecretSetting: DefaultAvitoClientSecretSetting,
	models.SpaceSenderEmail:         DefaultSpaceSenderEmail,
}