package licensepaymentstore

import (
	"hr-tools-backend/models"
	licenseapimodels "hr-tools-backend/models/api/license"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Provider interface {
	Create(rec dbmodels.LicensePayment) (id string, err error)
	GetByID(id string) (rec *dbmodels.LicensePayment, err error)
	Update(id string, status models.LicensePaymentStatus, updMap map[string]interface{}) error
	Delete(id string) error
	List(spaceID string, filter licenseapimodels.LicensePaymentFilter) (list []dbmodels.LicensePayment, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.LicensePayment) (id string, err error) {
	err = rec.Validate()
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

func (i impl) GetByID(id string) (*dbmodels.LicensePayment, error) {
	rec := dbmodels.LicensePayment{}
	err := i.db.
		Where("id = ?", id).
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

func (i impl) Update(id string, status models.LicensePaymentStatus, updMap map[string]interface{}) error {
	if len(updMap) == 0 {
		return nil
	}
	tx := i.db.
		Model(&dbmodels.LicensePayment{}).
		Where("id = ?", id).
		Where("status = ?", status).
		Updates(updMap)
	if tx.RowsAffected == 0 {
		return errors.Errorf("запись со статусом %v не найдена", status)
	}
	err := tx.Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) Delete(id string) error {
	rec := dbmodels.LicensePayment{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			BaseModel: dbmodels.BaseModel{
				ID: id,
			},
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

func (i impl) List(spaceID string, filter licenseapimodels.LicensePaymentFilter) (list []dbmodels.LicensePayment, err error) {
	var result []dbmodels.LicensePayment
	tx := i.db.
		Model(dbmodels.LicensePayment{}).
		Where("space_id = ?", spaceID)
	i.addFilter(tx, filter)
	err = tx.Find(&result).Error
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения истории платежей")
	}
	return result, nil
}

func (i impl) addFilter(tx *gorm.DB, filter licenseapimodels.LicensePaymentFilter) {
	if filter.Status != nil {
		tx.Where("license_payments.status = ?", *filter.Status)
	}

	tx.Order("created_at desc")
}
