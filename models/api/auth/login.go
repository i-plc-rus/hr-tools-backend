package authapimodels

import (
	"github.com/pkg/errors"
	"net/mail"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginRequest) Validate() error {
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return errors.New("почта имеет неправильный формат")
	}
	return nil
}

type PasswordRecovery struct {
	Email string `json:"email"` // емайл для отправки письма с иснтвукцией, он же логин
}

type PasswordResetRequest struct {
	ResetCode   string `json:"reset_code"`
	NewPassword string `json:"new_password"`
}

func (r PasswordRecovery) Validate() error {
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return errors.New("почта имеет неправильный формат")
	}
	return nil
}

func (r PasswordResetRequest) Validate() error {
	if r.ResetCode == "" {
		return errors.New("получен некорректный код для сброса")
	}
	if r.NewPassword == "" {
		return errors.New("не указан новый пароль")
	}
	return nil
}
