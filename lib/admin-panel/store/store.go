package adminpaneluserstore

import (
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(rec dbmodels.AdminPanelUser) (userID string, err error)
	GetByID(userID string) (*dbmodels.AdminPanelUser, error)
	FindByEmail(email string) (*dbmodels.AdminPanelUser, error)
	Update(userID string, updMap map[string]interface{}) error
	Delete(userID string) error
	List() ([]dbmodels.AdminPanelUser, error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.AdminPanelUser) (userID string, err error) {
	if rec.Email == "" {
		return "", errors.New("email не указан")
	}
	r, err := i.FindByEmail(rec.Email)
	if err != nil {
		return "", err
	}
	if r != nil {
		fmt.Println(r.ID)
		return "", errors.New("пользователь уже сущетсвует")
	}
	err = i.db.
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) GetByID(userID string) (*dbmodels.AdminPanelUser, error) {
	rec := dbmodels.AdminPanelUser{}
	err := i.db.
		Where("id = ?", userID).
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

func (i impl) FindByEmail(email string) (*dbmodels.AdminPanelUser, error) {
	rec := dbmodels.AdminPanelUser{}
	err := i.db.
		Where("email = ?", email).
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

func (i impl) Update(userID string, updMap map[string]interface{}) error {
	if len(updMap) == 0 {
		return nil
	}
	email, ok := updMap["Email"]
	if ok {
		existedRec, err := i.FindByEmail(email.(string))
		if err != nil {
			return err
		}
		if existedRec != nil && existedRec.ID != userID {
			return errors.New("пользователь с указанным email уже сущетсвует")
		}
	}
	err := i.db.
		Model(&dbmodels.AdminPanelUser{}).
		Where("id = ?", userID).
		Updates(updMap).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) Delete(userID string) error {
	rec := dbmodels.AdminPanelUser{}
	err := i.db.
		Where("id = ?", userID).
		Delete(&rec).
		Error

	if err != nil {
		return err
	}
	return nil
}

func (i impl) List() ([]dbmodels.AdminPanelUser, error) {
	list := []dbmodels.AdminPanelUser{}
	err := i.db.
		Find(&list).
		Error
	if err != nil {
		return nil, err
	}
	return list, nil
}
