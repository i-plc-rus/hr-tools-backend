package store

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(rec dbmodels.Department) (id string, err error)
	GetByID(spaceID, id string) (rec *dbmodels.Department, err error)
	FindByCompany(spaceID, companyID string) (list []dbmodels.Department, err error)
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

func (i impl) Create(rec dbmodels.Department) (id string, err error) {
	err = rec.Validate()
	if err != nil {
		return "", err
	}

	err = i.isUnique(rec.CompanyID, "", rec.Name)
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

func (i impl) GetByID(spaceID, id string) (*dbmodels.Department, error) {
	rec := dbmodels.Department{}
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

func (i impl) FindByCompany(spaceID, companyID string) (list []dbmodels.Department, err error) {
	list = []dbmodels.Department{}
	tx := i.db.
		Where("space_id = ?", spaceID)
	if companyID != "" {
		tx = tx.Where("company_id = ?", companyID)
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
		err = i.isUnique(rec.CompanyID, id, name.(string))
		if err != nil {
			return err
		}
	}
	err := i.db.
		Model(&dbmodels.Department{}).
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
	rec := dbmodels.Department{
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

func (i impl) isUnique(companyID string, selfID, name string) error {
	var rowCount int64
	tx := i.db.Model(dbmodels.Department{})
	tx.Where("company_id = ?", companyID)
	tx.Where("name = ?", name)
	if selfID != "" {
		tx.Where("id <> ?", selfID)
	}
	err := tx.Count(&rowCount).Error
	if err != nil {
		return errors.Wrap(err, "ошибка проверки уникальности подразделения")
	}
	if rowCount != 0 {
		return errors.New("подразделение уже существует")
	}
	return nil
}
