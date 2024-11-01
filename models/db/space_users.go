package dbmodels

import (
	"fmt"
	spaceapimodels "hr-tools-backend/models/api/space"
	"time"
)

type SpaceUser struct {
	BaseModel
	Password    string `gorm:"type:varchar(128)"`
	FirstName   string `gorm:"type:varchar(150)"`
	LastName    string `gorm:"type:varchar(150)"`
	IsAdmin     bool
	Email       string `gorm:"type:varchar(255)"`
	IsActive    bool
	PhoneNumber string `gorm:"type:varchar(15)"`
	SpaceID     string
	Role        string
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
			IsAdmin:     r.IsAdmin,
			SpaceID:     r.SpaceID,
		},
	}
}

func (r SpaceUser) GetFullName() string {
	return fmt.Sprintf("%s %s", r.FirstName, r.LastName)
}
