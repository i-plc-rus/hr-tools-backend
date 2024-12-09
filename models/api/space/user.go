package spaceapimodels

import "errors"

type CreateUser struct {
	Password string `json:"password"`
	SpaceUserCommonData
}

type UpdateUser struct {
	Password string `json:"password"`
	SpaceUserCommonData
}

type SpaceUser struct {
	ID string `json:"id"`
	SpaceUserCommonData
	IsEmailVerified bool   `json:"is_email_verified"` // Email подтвержден
	NewEmail        string `json:"new_email"`         // Новый email, который станет основным после подтверждения
}

type SpaceUserCommonData struct {
	SpaceID     string `json:"space_id"`
	Email       string `json:"email"` // Email пользователя
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	IsAdmin     bool   `json:"is_admin"`
	Role        string `json:"role"`
	TextSign    string `json:"text_sign"` // Текст подписи
}

func (r SpaceUserCommonData) Validate() error {
	//TODO: add data validators
	if r.Email == "" {
		return errors.New("не указан емайл")
	}
	return nil
}
