package pushsettingsstore

import (
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"

	"gorm.io/gorm"
)

type Provider interface {
	Create(rec dbmodels.SpacePushSetting) error
	Update(spaceID, userID string, code models.SpacePushSettingCode, updMap map[string]interface{}) error
	List(spaceID, userID string) (settingsList []dbmodels.SpacePushSetting, err error)
	GetByCode(userID string, code models.SpacePushSettingCode) (*dbmodels.SpacePushSetting, error)
	GetUsersWithoutSettings() (userList []dbmodels.SpaceUser, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.SpacePushSetting) error {
	return i.db.
		Save(&rec).
		Error
}

func (i impl) Update(spaceID, userID string, code models.SpacePushSettingCode, updMap map[string]interface{}) error {
	return i.db.
		Model(&dbmodels.SpacePushSetting{}).
		Where("space_id = ?", spaceID).
		Where("space_user_id = ?", userID).
		Where("code = ?", code).
		Updates(updMap).
		Error
}

func (i impl) List(spaceID, userID string) (settingsList []dbmodels.SpacePushSetting, err error) {
	tx := i.db.Model(dbmodels.SpacePushSetting{})
	err = tx.
		Where("space_id = ?", spaceID).
		Where("space_user_id = ?", userID).
		Find(&settingsList).
		Error
	if err != nil {
		return nil, err
	}
	return settingsList, nil
}

func (i impl) GetByCode(userID string, code models.SpacePushSettingCode) (rec *dbmodels.SpacePushSetting, err error) {
	err = i.db.Model(dbmodels.SpacePushSetting{}).
		Where("space_user_id = ?", userID).
		Where("code = ?", code).
		First(&rec).
		Error
	if err != nil {
		return nil, err
	}
	return rec, nil
}

func (i impl) GetUsersWithoutSettings() (userList []dbmodels.SpaceUser, err error) {
	tx := i.db.Model(dbmodels.SpaceUser{})
	subQuery := i.db.Select("space_user_id").Table("space_push_settings").Group("space_user_id")
	tx.Where("id not in (?)", subQuery)
	err = tx.
		Find(&userList).
		Error
	if err != nil {
		return nil, err
	}
	return userList, nil
}