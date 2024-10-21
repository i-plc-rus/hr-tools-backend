package dbmodels

import "hr-tools-backend/models"

type ApprovalStage struct {
	BaseSpaceModel
	VacancyRequestID string `gorm:"type:varchar(36)"`
	Stage            int
	SpaceUserID      string `gorm:"type:varchar(36)"`
	SpaceUser        *SpaceUser
	ApprovalStatus   models.ApprovalStatus
}
