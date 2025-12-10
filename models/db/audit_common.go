package dbmodels

import (
	"database/sql/driver"
	"encoding/json"
)

type EntityChanges struct {
	Description string         `json:"description"` // Комментрий
	Data        []FieldChanges `json:"data"`        // Список изменений
}

type FieldChanges struct {
	Field    string `json:"field"`     // Измененное поле
	OldValue any    `json:"old_value"` // Старое значение
	NewValue any    `json:"new_value"` // Новое значение
}

func (j EntityChanges) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *EntityChanges) Scan(value any) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}
