package spacestore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	CreateSpace(rec dbmodels.Space) (spaceID string, err error)
	UpdateSpace(spaceID string, updMap map[string]interface{}) error
	DeleteSpace(spaceID string) error
	SpaceWithInnExist(inn string) (bool, error)
}

var Instance Provider

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) UpdateSpace(spaceID string, updMap map[string]interface{}) error {
	err := i.db.
		Model(&dbmodels.Space{}).
		Where("id = ?", spaceID).
		Updates(updMap).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) SpaceWithInnExist(inn string) (bool, error) {
	err := i.db.
		First(&dbmodels.Space{}).
		Where("inn = ?", inn).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (i impl) DeleteSpace(spaceID string) error {
	rec := dbmodels.Space{}
	err := i.db.
		Where("id = ?", spaceID).
		Delete(&rec).
		Error

	if err != nil {
		return err
	}
	return nil
}

func (i impl) CreateSpace(rec dbmodels.Space) (spaceID string, err error) {
	err = i.db.
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}
