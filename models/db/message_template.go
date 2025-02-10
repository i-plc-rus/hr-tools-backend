package dbmodels

import (
	"hr-tools-backend/models"
	msgtemplateapimodels "hr-tools-backend/models/api/message-template"
)

type MessageTemplate struct {
	BaseSpaceModel
	Name         string              `gorm:"type:varchar(255)"`
	Title        string              `gorm:"type:varchar(255)"`
	Message      string              `gorm:"type:varchar(255)"`
	TemplateType models.TemplateType `gorm:"type:varchar(255)"`
}

func (r MessageTemplate) ToModel() msgtemplateapimodels.MsgTemplateView {
	return msgtemplateapimodels.MsgTemplateView{
		ID: r.ID,
		MsgTemplateData: msgtemplateapimodels.MsgTemplateData{
			Name:         r.Name,
			Title:        r.Title,
			Message:      r.Message,
			TemplateType: r.TemplateType,
		},
	}
}
