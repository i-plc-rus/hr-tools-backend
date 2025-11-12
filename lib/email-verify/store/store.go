package emailverifystore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

type Provider interface {
	Create(verifyData dbmodels.EmailVerify) error
	GetByCode(code string) (*dbmodels.EmailVerify, error)
	Exist(email string) (bool, error)
	DeleteByCode(code string) error
	UpdateByCode(code string, updMap map[string]interface{}) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) UpdateByCode(code string, updMap map[string]interface{}) error {
	err := i.db.
		Model(&dbmodels.EmailVerify{}).
		Where("code = ?", code).
		Updates(updMap).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) DeleteByCode(code string) error {
	return i.db.
		Where("code = ?", code).
		Delete(&dbmodels.EmailVerify{}).
		Error
}

func (i impl) Exist(email string) (bool, error) {
	err := i.db.
		Where("email = ?", email).
		Where("date_expires > ?", time.Now()).                 // игнорируем просроченные
		Where("date_used < ?", time.Now().AddDate(-50, 0, 0)). // игнорируем использованные
		First(&dbmodels.EmailVerify{}).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (i impl) Create(verifyData dbmodels.EmailVerify) error {
	err := i.db.
		Save(&verifyData).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) GetByCode(code string) (*dbmodels.EmailVerify, error) {
	rec := dbmodels.EmailVerify{}
	err := i.db.
		Where("code = ?", code).
		First(&rec).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}
