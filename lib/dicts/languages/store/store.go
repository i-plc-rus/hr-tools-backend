package languagestore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
	"strings"
)

type Provider interface {
	List(name string) ([]dbmodels.LanguageData, error)
	Add(rec dbmodels.LanguageData, skipDuplicate bool) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) List(name string) ([]dbmodels.LanguageData, error) {
	var result []dbmodels.LanguageData
	tx := i.db.Model(dbmodels.LanguageData{})
	if name != "" {
		tx.Where("LOWER(name) like ?", "%"+strings.ToLower(name)+"%")
	}
	err := tx.Find(&result).Error
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения списка языков")
	}
	return result, nil
}

func (i impl) Add(rec dbmodels.LanguageData, skipDuplicate bool) error {
	item := dbmodels.LanguageData{
		BaseModel: dbmodels.BaseModel{ID: rec.ID},
		Code:      rec.Code,
		Name:      rec.Name,
	}
	unique, err := i.isUnique(item.Code, item)
	if err != nil {
		return err
	}
	if !unique {
		if skipDuplicate {
			return nil
		}
		return errors.New("город уже существует")
	}
	tx := i.db.Save(&rec)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "ошибка добавления города")
	}
	return nil
}

func (i impl) isUnique(selfID string, item dbmodels.LanguageData) (bool, error) {
	var rowCount int64
	tx := i.db.Model(dbmodels.City{})
	tx.Where("LOWER(code) = ?", strings.ToLower(item.Code))
	if selfID != "" {
		tx.Where("id <> ?", selfID)
	}
	err := tx.Count(&rowCount).Error
	if err != nil {
		return false, errors.Wrap(err, "ошибка проверки уникальности языка")
	}
	return rowCount == 0, nil
}
