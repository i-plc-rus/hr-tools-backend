package spaceapimodels

import (
	"github.com/pkg/errors"
	"strings"
)

type SpaceSettingView struct {
	ID      string `json:"id"`       // идентификтор Настройки
	SpaceID string `json:"space_id"` // идентификатор пространства, которому принадлежит настройка
	Name    string `json:"name"`     // Название настройки
	Code    string `json:"code"`     // Код настройки
	Value   string `json:"value"`    // Значение настройки
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
