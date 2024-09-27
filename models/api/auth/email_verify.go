package authapimodels

import (
	"github.com/pkg/errors"
	"net/mail"
)

type SendEmail struct {
	Email string `json:"email"` // Почта, на которую надо отправить письмо с подтверждением
}

func (r SendEmail) Validate() error {
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return errors.New("почта имеет не правильный формат")
	}
	return nil
}
