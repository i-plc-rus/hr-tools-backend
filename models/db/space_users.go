package dbmodels

import (
	"fmt"
	"hr-tools-backend/models"
	spaceapimodels "hr-tools-backend/models/api/space"
	"time"
)

type SpaceUser struct {
	BaseModel
	Password            string          `gorm:"type:varchar(128)"`
	FirstName           string          `gorm:"type:varchar(150)"`
	LastName            string          `gorm:"type:varchar(150)"`
	Email               string          `gorm:"type:varchar(255)"`
	IsActive            bool            `json:"is_active"`
	PhoneNumber         string          `gorm:"type:varchar(15)"`
	InternalPhoneNumber string          `gorm:"type:varchar(15)"`
	SpaceID             string          `json:"space_id"`
	Role                models.UserRole `gorm:"type:varchar(50)"`
	LastLogin           time.Time       `json:"last_login"`
	TextSign            string          `gorm:"type:varchar(1000)"` // текст подписи
	NewEmail            string          `gorm:"type:varchar(255)"`
	IsEmailVerified     bool
	ResetCode           string `gorm:"type:varchar(36);index"`
	ResetTime           time.Time
	UsePersonalSign     bool    // использовать личный текст подписи
	JobTitleID          *string `gorm:"type:varchar(36)"`
	JobTitle            *JobTitle
	PushEnabled         bool
	DeletedAt           *time.Time `gorm:"index"`
}

func (r SpaceUser) ToModel() spaceapimodels.SpaceUser {

	result := spaceapimodels.SpaceUser{
		ID: r.ID,
		SpaceUserCommonData: spaceapimodels.SpaceUserCommonData{
			Email:       r.Email,
			FirstName:   r.FirstName,
			LastName:    r.LastName,
			PhoneNumber: r.PhoneNumber,
			IsAdmin:     r.Role.IsSpaceAdmin(),
			SpaceID:     r.SpaceID,
			Role:        r.Role.ToHuman(),
			TextSign:    r.TextSign,
		},
		IsEmailVerified: r.IsEmailVerified,
		NewEmail:        r.NewEmail,
	}
	if r.JobTitle != nil {
		result.JobTitleID = *r.JobTitleID
		result.JobTitleName = r.JobTitle.Name
	}
	return result
}

func (r SpaceUser) ToProfile() spaceapimodels.SpaceUserProfileView {
	result := spaceapimodels.SpaceUserProfileView{
		ID:              r.ID,
		Role:            r.Role.ToHuman(),
		IsEmailVerified: r.IsEmailVerified,
		NewEmail:        r.NewEmail,
		SpaceUserProfileData: spaceapimodels.SpaceUserProfileData{
			Email:               r.Email,
			FirstName:           r.FirstName,
			LastName:            r.LastName,
			PhoneNumber:         r.PhoneNumber,
			InternalPhoneNumber: r.InternalPhoneNumber,
			UsePersonalSign:     r.UsePersonalSign,
			TextSign:            r.TextSign,
		},
	}
	if r.JobTitle != nil {
		result.JobTitleName = r.JobTitle.Name
	}
	return result
}

func (r SpaceUser) GetFullName() string {
	return fmt.Sprintf("%s %s", r.FirstName, r.LastName)
}
