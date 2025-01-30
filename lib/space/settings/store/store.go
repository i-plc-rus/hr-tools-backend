package spacesettingsstore

import (
	"errors"
	"gorm.io/gorm"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(rec dbmodels.SpaceSetting) error
	Update(spaceID, code, value string) error
	List(spaceID string) (settingsList []dbmodels.SpaceSetting, err error)
	GetValueByCode(spaceID string, code models.SpaceSettingCode) (value string, err error)
	Delete(spaceID, code string) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.SpaceSetting) error {
	return i.db.
		Save(&rec).
		Error
}

func (i impl) GetValueByCode(spaceID string, code models.SpaceSettingCode) (value string, err error) {
	err = i.db.Model(dbmodels.SpaceSetting{}).
		Select("value").
		Where("space_id = ? AND code = ?", spaceID, code).
		First(&value).
		Error
	if err != nil {
		return "", err
	}
	return value, nil
}

func (i impl) List(spaceID string) (settingsList []dbmodels.SpaceSetting, err error) {
	tx := i.db.Model(dbmodels.SpaceSetting{})
	err = tx.
		Where("space_id = ?", spaceID).
		Find(&settingsList).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return settingsList, nil
}

func (i impl) Update(spaceID, code, value string) error {
	updMap := map[string]interface{}{
		"value": value,
	}
	return i.db.
		Model(&dbmodels.SpaceSetting{}).
		Where("space_id = ? AND code = ?", spaceID, code).
		Updates(updMap).
		Error
}

func (i impl) Delete(spaceID, code string) error {
	rec := dbmodels.SpaceSetting{}
	err := i.db.
		Where("space_id = ? AND code = ?", spaceID, code).
		Delete(&rec).
		Error

	if err != nil {
		return err
	}
	return nil
}
