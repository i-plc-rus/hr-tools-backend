package dbmodels

import "time"

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
