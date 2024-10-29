package gptmodels

import (
	"github.com/pkg/errors"
	"strings"
)

type GenVacancyDescRequest struct {
	Text string `json:"text"` // Текст, на основе которого необходимо сгенерировать описание
}

func (r GenVacancyDescRequest) Validate() error {
	if len(strings.TrimSpace(r.Text)) == 0 {
		return errors.New("текст не должен быть пустым")
	}
	return nil
}

type GenVacancyDescResponse struct {
	Description string `json:"description"` // сгенерированное описание вакансии
}
