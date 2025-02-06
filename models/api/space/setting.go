package spaceapimodels

import (
	"hr-tools-backend/models"
	"strings"

	"github.com/pkg/errors"
)

type SpaceSettingView struct {
	ID      string                  `json:"id"`       // идентификтор Настройки
	SpaceID string                  `json:"space_id"` // идентификатор пространства, которому принадлежит настройка
	Name    string                  `json:"name"`     // Название настройки
	Code    models.SpaceSettingCode `json:"code"`     // Код настройки
	Value   string                  `json:"value"`    // Значение настройки
}

type UpdateSpaceSettingValue struct {
	Value string `json:"value"` // Новое значение настройки
}

func (r UpdateSpaceSettingValue) Validate() error {
	if strings.TrimSpace(r.Value) == "" {
		return errors.New("не указано новое значение настройки")
	}
	return nil
}
