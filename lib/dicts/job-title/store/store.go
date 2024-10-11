package store

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
	"strings"
)

type Provider interface {
	Create(rec dbmodels.JobTitle) (id string, err error)
	GetByID(spaceID, id string) (rec *dbmodels.JobTitle, err error)
	FindByName(spaceID, name, departmentID string) (list []dbmodels.JobTitle, err error)
	Update(spaceID, id string, updMap map[string]interface{}) error
	Delete(spaceID, id string) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.JobTitle) (id string, err error) {
	err = rec.Validate()
	if err != nil {
		return "", err
	}

	err = i.isUnique(rec.DepartmentID, "", rec.Name)
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

func (i impl) GetByID(spaceID, id string) (*dbmodels.JobTitle, error) {
	rec := dbmodels.JobTitle{}
	err := i.db.
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
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

func (i impl) FindByName(spaceID, name, departmentID string) (list []dbmodels.JobTitle, err error) {
	list = []dbmodels.JobTitle{}
	tx := i.db.
		Where("space_id = ?", spaceID).
		Where("department_id = ?", departmentID)
	if name != "" {
		tx = tx.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")
	}
	err = tx.Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) Update(spaceID, id string, updMap map[string]interface{}) error {
	if len(updMap) == 0 {
		return nil
	}
	name, ok := updMap["name"]
	if ok {
		rec, err := i.GetByID(spaceID, id)
		if err != nil {
			return err
		}
		if rec == nil {
			return errors.New("запись не найдена")
		}
		err = i.isUnique(rec.DepartmentID, id, name.(string))
		if err != nil {
			return err
		}
	}
	err := i.db.
		Model(&dbmodels.JobTitle{}).
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
		Updates(updMap).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) Delete(spaceID, id string) error {
	rec := dbmodels.JobTitle{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			BaseModel: dbmodels.BaseModel{
				ID: id,
			},
			SpaceID: spaceID,
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

func (i impl) isUnique(departmentID string, selfID, name string) error {
	var rowCount int64
	tx := i.db.Model(dbmodels.JobTitle{})
	tx.Where("department_id = ?", departmentID)
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
