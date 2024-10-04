package authapimodels

import (
	"github.com/pkg/errors"
	"strings"
)

type JWTResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type JWTRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r JWTRefreshRequest) Validate() error {
	if len(strings.TrimSpace(r.RefreshToken)) == 0 {
		return errors.New("refresh token не должен быть пустым")
	}
	return nil
}
