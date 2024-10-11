package dbmodels

import (
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	"time"
)

type AdminPanelUser struct {
	BaseModel
	IsActive    bool
	Role        models.UserRole `gorm:"type:varchar(255)"`
	Password    string          `gorm:"type:varchar(128)"`
	FirstName   string          `gorm:"type:varchar(150)"`
	LastName    string          `gorm:"type:varchar(150)"`
	Email       string          `gorm:"type:varchar(255)"`
	PhoneNumber string          `gorm:"type:varchar(15)"`
	LastLogin   time.Time
}

func (u AdminPanelUser) Validate() error {
	if u.Email == "" {
		return errors.New("email не указан")
	}
	return nil
}

func (u AdminPanelUser) IsSuperAdmin() bool {
	return u.Role == models.UserRoleSuperAdmin
}
