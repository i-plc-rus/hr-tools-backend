package spaceusersstore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(rec dbmodels.SpaceUser) error
	ExistByEmail(email string) (bool, error)
	FindByEmail(email string) (rec *dbmodels.SpaceUser, err error)
	GetByID(userID string) (rec *dbmodels.SpaceUser, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) GetByID(userID string) (rec *dbmodels.SpaceUser, err error) {
	err = i.db.
		First(&rec).
		Where("id = ?", userID).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rec, nil
}

func (i impl) FindByEmail(email string) (rec *dbmodels.SpaceUser, err error) {
	err = i.db.
		First(&rec).
		Where("email = ?", email).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rec, nil
}

func (i impl) Create(rec dbmodels.SpaceUser) error {
	return i.db.
		Save(&rec).
		Error
}

func (i impl) ExistByEmail(email string) (bool, error) {
	err := i.db.
		First(&dbmodels.SpaceUser{}).
		Where("email = ?", email).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
