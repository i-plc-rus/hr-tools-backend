package dbmodels

import (
	"database/sql/driver"
	"encoding/json"
)

type ApplicantHistory struct {
	BaseSpaceModel
	ApplicantID string `gorm:"type:varchar(36);index"`
	VacancyID   string
	Vacancy     *Vacancy `gorm:"foreignKey:VacancyID"`
	UserID      *string
	UserName    string
	ActionType  ActionType       `gorm:"type:varchar(255)"`
	Changes     ApplicantChanges `gorm:"type:jsonb"`
}

func (j ApplicantChanges) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *ApplicantChanges) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

type ApplicantChanges struct {
	Description string            `json:"description"` // Комментрий
	Data        []ApplicantChange `json:"data"`        // Список изменений
}

type ApplicantChange struct {
	Field    string      `json:"field"`     // Измененное поле
	OldValue interface{} `json:"old_value"` // Старое значение
	NewValue interface{} `json:"new_value"` // Новое значение
}

type ActionType string

const (
	HistoryTypeComment     ActionType = "comment"      // Добавлен комментраий к кандидату
	HistoryTypeAdded       ActionType = "added"        // Кандидат добавлен
	HistoryTypeUpdate      ActionType = "update"       // Кандидат обновлен
	HistoryTypeNegotiation ActionType = "negotiation"  // Получен отклик от кандидата
	HistoryTypeStageChange ActionType = "stage_change" // Кандидат переведеден на другой этап
	HistoryTypeDuplicate   ActionType = "duplicate"    // Дубликат по кандидату
	HistoryTypeArchive     ActionType = "archive"      // Перемещен в архив
	HistoryTypeReject      ActionType = "reject"       // Кандидат отклонен
	HistoryTypeEmail       ActionType = "reject"       // email
)
