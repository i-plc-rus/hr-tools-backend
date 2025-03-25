package spaceapimodels

import (
	"strings"

	"github.com/pkg/errors"
)

type SalesRequest struct {
	Text string `json:"text"` // Текст заявки
}

func (r SalesRequest) Validate() error {
	if strings.TrimSpace(r.Text) == "" {
		return errors.New("отсутсвует текст заявки")
	}
	return nil
}
