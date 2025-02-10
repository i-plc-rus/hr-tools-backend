package dbmodels

import (
	"hr-tools-backend/models"
)

type RejectReason struct {
	BaseSpaceModel
	Initiator models.RejectInitiator `gorm:"type:varchar(255)"`
	Name      string                 `gorm:"type:varchar(255)"`
}

