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
