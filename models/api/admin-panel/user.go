package adminpanelapimodels

import (
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

type UserID struct {
	ID string `json:"id"`
}

func (u UserID) Validate() error {
	if u.ID == "" {
		return errors.New("ID пользователя не указан")
	}
	return nil
}

type UserView struct {
	User
	ID        string     `json:"id"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

type User struct {
	Email       string          `json:"email"`
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	PhoneNumber string          `json:"phone_number"`
	Password    string          `json:"password,omitempty"`
	Role        models.UserRole `json:"role"`
}

func (u User) Validate() error {
	if u.Email == "" {
		return errors.New("email не указан")
	}
	return nil
}

func UserConvert(rec dbmodels.AdminPanelUser) UserView {
	return UserView{
		User: User{
			Email:       rec.Email,
			FirstName:   rec.FirstName,
			LastName:    rec.LastName,
			PhoneNumber: rec.PhoneNumber,
			Role:        rec.Role,
		},
		ID:        rec.ID,
		LastLogin: &rec.LastLogin,
	}
}

type UserUpdate struct {
	ID          string           `json:"id"`
	Email       *string          `json:"email"`
	FirstName   *string          `json:"first_name"`
	LastName    *string          `json:"last_name"`
	PhoneNumber *string          `json:"phone_number"`
	Password    *string          `json:"password"`
	Role        *models.UserRole `json:"role"`
	IsActive    *bool            `json:"is_active"`
}

func (u UserUpdate) Validate() error {
	if u.ID == "" {
		return errors.New("ID пользователя не указан")
	}
	return nil
}
