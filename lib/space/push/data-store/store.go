package pushdatastore

import (
	dbmodels "hr-tools-backend/models/db"

	"gorm.io/gorm"
)

type Provider interface {
	Create(rec dbmodels.PushData) error
	List(userID string) ([]dbmodels.PushData, error)
	Delete(ids []string) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.PushData) error {
	return i.db.
		Save(&rec).
		Error
}

func (i impl) List(userID string) (list []dbmodels.PushData, err error) {
	tx := i.db.Model(dbmodels.PushData{})
	err = tx.
		Where("user_id = ?", userID).
		Order("created_at").
		Find(&list).
		Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (i impl) Delete(ids []string) error {
	return i.db.Delete(&dbmodels.PushData{}, ids).Error
}
