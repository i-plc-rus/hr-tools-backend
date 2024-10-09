package spaceusersstore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(rec dbmodels.SpaceUser) error
	Update(userID string, updMap map[string]interface{}) error
	Delete(userID string) error
	GetList(spaceID string, page, limit int) (userList []dbmodels.SpaceUser, err error)
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

func (i impl) GetList(spaceID string, page, limit int) (userList []dbmodels.SpaceUser, err error) {
	tx := i.db.Model(dbmodels.SpaceUser{})
	i.setPage(tx, page, limit)
	err = tx.
		Where("space_id = ?", spaceID).
		Find(&userList).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return userList, nil
}

func (i impl) Delete(userID string) error {
	return i.db.
		Where("id = ?", userID).
		Delete(&dbmodels.SpaceUser{}).
		Error
}

func (i impl) Update(userID string, updMap map[string]interface{}) error {
	err := i.db.
		Model(&dbmodels.Space{}).
		Where("id = ?", userID).
		Updates(updMap).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) GetByID(userID string) (rec *dbmodels.SpaceUser, err error) {
	err = i.db.
		Where("id = ?", userID).
		First(&rec).
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
		Where("email = ?", email).
		First(&rec).
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
		Where("email = ?", email).
		First(&dbmodels.SpaceUser{}).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (i impl) setPage(tx *gorm.DB, pageValue, limitValue int) {
	page, limit := GetPage(pageValue, limitValue)
	offset := (page - 1) * limit
	tx.Limit(limit).Offset(offset)
}

func GetPage(pageValue, limitValue int) (page, limit int) {
	page = 1
	limit = 10
	if pageValue > 0 {
		page = pageValue
	}
	if limitValue > 0 {
		limit = limitValue
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
