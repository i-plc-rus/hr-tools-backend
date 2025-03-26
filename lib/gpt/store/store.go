package ailogstore

import (
	dbmodels "hr-tools-backend/models/db"

	"gorm.io/gorm"
)

type Provider interface {
	Save(rec dbmodels.AiLog) (string, error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Save(rec dbmodels.AiLog) (string, error) {
	err := i.db.
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}
