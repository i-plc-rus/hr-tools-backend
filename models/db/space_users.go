package dbmodels

import (
	"fmt"
	"hr-tools-backend/models"
	spaceapimodels "hr-tools-backend/models/api/space"
	"time"
)

type SpaceUser struct {
	BaseModel
	Password    string `gorm:"type:varchar(128)"`
	FirstName   string `gorm:"type:varchar(150)"`
	LastName    string `gorm:"type:varchar(150)"`
	Email       string `gorm:"type:varchar(255)"`
	IsActive    bool
	PhoneNumber string `gorm:"type:varchar(15)"`
	SpaceID     string
	Role        models.UserRole `gorm:"type:varchar(50)"`
	LastLogin   time.Time
}

func (r SpaceUser) ToModel() spaceapimodels.SpaceUser {
	return spaceapimodels.SpaceUser{
		ID: r.ID,
		SpaceUserCommonData: spaceapimodels.SpaceUserCommonData{
			Email:       r.Email,
			FirstName:   r.FirstName,
			LastName:    r.LastName,
			PhoneNumber: r.PhoneNumber,
			IsAdmin:     r.Role.IsSpaceAdmin(),
			SpaceID:     r.SpaceID,
			Role:        r.Role.ToHuman(),
		},
	}
}

func (r SpaceUser) GetFullName() string {
	return fmt.Sprintf("%s %s", r.FirstName, r.LastName)
}
