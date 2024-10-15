package citystore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
	"strings"
)

type Provider interface {
	List(address string) ([]dbmodels.City, error)
	Add(rec dbmodels.City, skipDuplicate bool) error
	GetByID(id string) (*dbmodels.City, error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) List(address string) ([]dbmodels.City, error) {
	var result []dbmodels.City
	tx := i.db.Model(dbmodels.City{})
	if address != "" {
		tx.Where("LOWER(address) like ?", "%"+strings.ToLower(address)+"%")
	}
	err := tx.Find(&result).Error
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения списка городов")
	}
	return result, nil
}

func (i impl) Add(rec dbmodels.City, skipDuplicate bool) error {
	item := dbmodels.City{
		BaseModel: dbmodels.BaseModel{ID: rec.ID},
		Address:   rec.Address,
	}
	unique, err := i.isUnique(item.ID, item)
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

func (i impl) GetByID(id string) (*dbmodels.City, error) {
	rec := dbmodels.City{BaseModel: dbmodels.BaseModel{ID: id}}
	err := i.db.First(&rec).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (i impl) isUnique(selfID string, item dbmodels.City) (bool, error) {
	var rowCount int64
	tx := i.db.Model(dbmodels.City{})
	tx.Where("LOWER(address) = ?", strings.ToLower(item.Address))
	if selfID != "" {
		tx.Where("id <> ?", selfID)
	}
	err := tx.Count(&rowCount).Error
	if err != nil {
		return false, errors.Wrap(err, "ошибка проверки уникальности города")
	}
	return rowCount == 0, nil
}
