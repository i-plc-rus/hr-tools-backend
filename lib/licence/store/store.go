package licensestore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"

)

type Provider interface {
	Create(rec dbmodels.License) (id string, err error)
	GetByID(spaceID, id string) (*dbmodels.License, error)
	GetBySpace(spaceID string) (rec *dbmodels.License, err error)
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

func (i impl) Create(rec dbmodels.License) (id string, err error) {
	err = rec.Validate()
	if err != nil {
		return "", err
	}

	err = i.isUnique(rec.SpaceID)
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

func (i impl) GetByID(spaceID, id string) (*dbmodels.License, error) {
	rec := dbmodels.License{}
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


func (i impl) GetBySpace(spaceID string) (*dbmodels.License, error) {
	rec := dbmodels.License{}
	err := i.db.
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

func (i impl) Update(spaceID, id string, updMap map[string]interface{}) error {
	if len(updMap) == 0 {
		return nil
	}
	err := i.db.
		Model(&dbmodels.License{}).
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
	rec := dbmodels.License{
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

func (i impl) isUnique(spaceID string) error {
	var rowCount int64
	tx := i.db.Model(dbmodels.License{})
	tx.Where("space_id = ?", spaceID)
	err := tx.Count(&rowCount).Error
	if err != nil {
		return errors.Wrap(err, "ошибка проверки уникальности штатной должности")
	}
	if rowCount != 0 {
		return errors.New("лицензия уже существует")
	}
	return nil
}
