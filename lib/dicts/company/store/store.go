package store

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
	"strings"
)

type Provider interface {
	Create(rec dbmodels.Company) (id string, err error)
	GetByID(spaceID, id string) (rec *dbmodels.Company, err error)
	FindByName(spaceID, name string) (list []dbmodels.Company, err error)
	Update(spaceID, id string, updMap map[string]interface{}) error
	Delete(spaceID, id string) error
	FindOrCreate(spaceID, name string) (id string, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.Company) (id string, err error) {
	err = rec.Validate()
	if err != nil {
		return "", err
	}

	err = i.isUnique(rec.SpaceID, "", rec.Name)
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

func (i impl) GetByID(spaceID, id string) (*dbmodels.Company, error) {
	rec := dbmodels.Company{}
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

func (i impl) FindByName(spaceID, name string) (list []dbmodels.Company, err error) {
	list = []dbmodels.Company{}
	tx := i.db.Where("space_id = ?", spaceID)
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
		err := i.isUnique(spaceID, id, name.(string))
		if err != nil {
			return err
		}
	}
	err := i.db.
		Model(&dbmodels.Company{}).
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
	rec := dbmodels.Company{
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

func (i impl) FindOrCreate(spaceID, name string) (id string, err error) {
	rec := dbmodels.Company{}
	tx := i.db.Where("space_id = ?", spaceID)

	tx = tx.Where("LOWER(name) = ?", strings.ToLower(name))
	err = tx.First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rec.SpaceID = spaceID
			rec.Name = name
			return i.Create(rec)
		}
		return "", err
	}

	return rec.ID, nil
}

func (i impl) isUnique(spaceID string, selfID, name string) error {
	var rowCount int64
	tx := i.db.Model(dbmodels.Company{})
	tx.Where("space_id = ?", spaceID)
	tx.Where("name = ?", name)
	if selfID != "" {
		tx.Where("id <> ?", selfID)
	}
	err := tx.Count(&rowCount).Error
	if err != nil {
		return errors.Wrap(err, "ошибка проверки уникальности компании")
	}
	if rowCount != 0 {
		return errors.New("компания уже существует")
	}
	return nil
}
