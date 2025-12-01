package licenseplanstore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
	"strings"
)

type Provider interface {
	Create(rec dbmodels.LicensePlan) (id string, err error)
	GetByID(id string) (rec *dbmodels.LicensePlan, err error)
	GetByName(name string) (rec *dbmodels.LicensePlan, err error)
	FindByName(name string) (list []dbmodels.LicensePlan, err error)
	Update(id string, updMap map[string]interface{}) error
	Delete(id string) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.LicensePlan) (id string, err error) {
	err = rec.Validate()
	if err != nil {
		return "", err
	}

	err = i.isUnique("", rec.Name)
	if err != nil {
		return "", err
	}
	err = i.db.
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) GetByID(id string) (*dbmodels.LicensePlan, error) {
	rec := dbmodels.LicensePlan{}
	err := i.db.
		Where("id = ?", id).
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

func (i impl) GetByName(name string) (*dbmodels.LicensePlan, error) {
	rec := dbmodels.LicensePlan{}
	err := i.db.
		Where("name = ?", name).
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

func (i impl) FindByName(name string) (list []dbmodels.LicensePlan, err error) {
	list = []dbmodels.LicensePlan{}
	tx := i.db.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")

	err = tx.Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) Update(id string, updMap map[string]interface{}) error {
	if len(updMap) == 0 {
		return nil
	}
	name, ok := updMap["name"]
	if ok {
		rec, err := i.GetByID(id)
		if err != nil {
			return err
		}
		if rec == nil {
			return errors.New("запись не найдена")
		}
		err = i.isUnique(id, name.(string))
		if err != nil {
			return err
		}
	}
	err := i.db.
		Model(&dbmodels.LicensePlan{}).
		Where("id = ?", id).
		Updates(updMap).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) Delete(id string) error {
	rec := dbmodels.LicensePlan{
		BaseModel: dbmodels.BaseModel{
			ID: id,
		},
	}
	err := i.db.
		Delete(&rec).
		Error

	if err != nil {
		return err
	}
	return nil
}

func (i impl) isUnique(selfID, name string) error {
	var rowCount int64
	tx := i.db.Model(dbmodels.LicensePlan{})
	tx.Where("name = ?", name)
	if selfID != "" {
		tx.Where("id <> ?", selfID)
	}
	err := tx.Count(&rowCount).Error
	if err != nil {
		return errors.Wrap(err, "ошибка проверки уникальности штатной должности")
	}
	if rowCount != 0 {
		return errors.New("штатная должность уже существует")
	}
	return nil
}
