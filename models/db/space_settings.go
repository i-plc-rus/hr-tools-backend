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
	Value   string                  `gorm:"type:varchar(255)"`
}

func (r SpaceSetting) ToModelView() spaceapimodels.SpaceSettingView {
	return spaceapimodels.SpaceSettingView{
		ID:      r.ID,
		SpaceID: r.SpaceID,
		Name:    r.Name,
		Code:    r.Name,
		Value:   r.Value,
	}
}
