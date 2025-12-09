package dbmodels

import (
	"hr-tools-backend/models"
	"time"
)

type ApprovalTask struct {
	BaseSpaceModel
	RequestID      string     `gorm:"type:varchar(36)"`
	AssigneeUserID string     `gorm:"type:varchar(36)"`
	AssigneeUser   *SpaceUser `gorm:"foreignKey:AssigneeUserID"`
	State          models.ApprovalState
	Comment        string
	DecidedAt      *time.Time
}

type ApprovalHistory struct {
	BaseSpaceModel
	RequestID      string     `gorm:"type:varchar(36)"`
	TaskID         string     `gorm:"type:varchar(36)"`
	AssigneeUserID string     `gorm:"type:varchar(36)"`
	AssigneeUser   *SpaceUser `gorm:"foreignKey:AssigneeUserID"`
	State          models.ApprovalState
	Comment        string
	Changes        EntityChanges `gorm:"type:jsonb"`
}
