package dbmodels

import (
	"hr-tools-backend/models"
	spaceapimodels "hr-tools-backend/models/api/space"
)

type SpacePushSetting struct {
	BaseSpaceModel
	SpaceUserID string                      `gorm:"type:varchar(36);uniqueIndex:idx_user"`
	Code        models.SpacePushSettingCode `gorm:"type:varchar(255);index:idx_setting_code"`
	SystemValue *bool
	EmailValue  *bool
	TgValue     *bool
}

func (r SpacePushSetting) ToModelView() spaceapimodels.PushSettingView {
	return spaceapimodels.PushSettingView{
		Name: models.PushCodeMap[r.Code].Name,
		PushSettingData: spaceapimodels.PushSettingData{
			Code: r.Code,
			Value: spaceapimodels.PushSettingValue{
				System: r.SystemValue,
				Email:  r.EmailValue,
				Tg:     r.TgValue,
			},
		},
	}
}
