package licensestore

import (
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Provider interface {
	Create(rec dbmodels.License) (id string, err error)
	GetByID(spaceID, id string) (*dbmodels.License, error)
	GetBySpaceExt(spaceID string) (*dbmodels.LicenseExt, error)
	GetBySpace(spaceID string) (rec *dbmodels.License, err error)
	Update(spaceID, id string, updMap map[string]interface{}) error
	Delete(spaceID, id string) error
	ListToExpired(status models.LicenseStatus, endsAt time.Time) ([]dbmodels.License, error)
	IsExist(spaceID string) (bool, error)
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

	isExist, err := i.IsExist(rec.SpaceID)
	if err != nil {
		return "", err
	}
	if isExist {
		return "", errors.New("у организации уже существует лицензия")
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

func (i impl) GetBySpaceExt(spaceID string) (*dbmodels.LicenseExt, error) {
	rec := dbmodels.LicenseExt{}
	err := i.db.
		Select("licenses.*, p.id as plan_id, p.name as plan_name, p.cost as plan_cost, p.extension_period_days as plan_period_days").
		Model(&dbmodels.License{}).
		Joins("left join license_plans as p on plan = p.Name").
		Where("licenses.space_id = ?", spaceID).
		Preload(clause.Associations).
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

func (i impl) ListToExpired(status models.LicenseStatus, endsAt time.Time) ([]dbmodels.License, error) {
	list := []dbmodels.License{}
	tx := i.db.
		Model(dbmodels.License{}).
		Where("status = ?", status).
		Where("ends_at <= ?", endsAt)
	err := tx.Find(&list).Error

	if err != nil {
		return nil, err
	}
	return list, nil
}

func (i impl) IsExist(spaceID string) (bool, error) {
	var rowCount int64
	tx := i.db.Model(dbmodels.License{})
	tx.Where("space_id = ?", spaceID)
	err := tx.Count(&rowCount).Error
	if err != nil {
		return false, err
	}
	return rowCount != 0, nil
}
