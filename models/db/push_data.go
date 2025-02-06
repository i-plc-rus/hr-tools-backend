package dbmodels

import "hr-tools-backend/models"

type PushData struct {
	BaseModel
	UserID string                      `gorm:"type:varchar(36);index:idx_user"`
	Code   models.SpacePushSettingCode `gorm:"type:varchar(255);index:idx_setting_code"`
	Msg    string
	Title  string
}
