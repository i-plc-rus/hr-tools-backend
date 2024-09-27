package dbmodels

import (
	"time"
)

type BaseModel struct {
	ID        string    `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
