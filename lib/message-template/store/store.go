package messagetemplatestore

import (
	"errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(rec dbmodels.MessageTemplate) (string, error)
	Update(id string, updMap map[string]interface{}) error
	GetByID(spaceID, id string) (rec *dbmodels.MessageTemplate, err error)
	List(spaceID string) (list []dbmodels.MessageTemplate, err error)
	Delete(spaceID, id string) error
}

func NewInstance(db *gorm.DB) Provider {
	return &impl{
		db: db,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) GetByID(spaceID, id string) (rec *dbmodels.MessageTemplate, err error) {
	err = i.db.
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
	return rec, nil
}

func (i impl) Create(rec dbmodels.MessageTemplate) (string, error) {
	err := i.db.Save(&rec).Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) Update(id string, updMap map[string]interface{}) error {
	return i.db.
		Model(&dbmodels.MessageTemplate{}).
		Where("id = ?", id).
		Updates(updMap).
		Error
}

func (i impl) List(spaceID string) (list []dbmodels.MessageTemplate, err error) {
	err = i.db.
		Model(&dbmodels.MessageTemplate{}).
		Where("space_id = ?", spaceID).
		Find(&list).
		Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (i impl) Delete(spaceID, id string) error {
	rec := dbmodels.MessageTemplate{
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
