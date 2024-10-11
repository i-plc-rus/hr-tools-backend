package dbmodels

import (
	"github.com/pkg/errors"
	"time"
)

type BaseModel struct {
	ID        string    `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BaseSpaceModel struct {
	BaseModel
	SpaceID string `gorm:"type:varchar(36);index"`
}

func (c BaseSpaceModel) Validate() error {
	if c.SpaceID == "" {
		return errors.New("отсутсвует ссылка на организацию в space")
	}
	return nil
}
