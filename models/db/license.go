package dbmodels

import (
	"hr-tools-backend/models"
	"time"

	"github.com/pkg/errors"
)

type License struct {
	BaseSpaceModel
	Status          models.LicenseStatus `gorm:"type:varchar(255)"`
	StartsAt        *time.Time
	EndsAt          *time.Time
	Plan            string
	AutoRenew       bool
	LicensePayments []LicensePayment
}

type LicenseExt struct {
	License
	PlanID          string
	PlanName        string
	PlanCost        float64
	PlanPeriodDays  int
}

func (j License) Validate() error {
	if err := j.BaseSpaceModel.Validate(); err != nil {
		return err
	}
	if j.Status == "" {
		return errors.New("отсутсвует статус")
	}
	if j.Plan == "" {
		return errors.New("не указано тариф")
	}
	return nil
}

type LicensePayment struct {
	BaseSpaceModel
	LicenseID string `gorm:"index"`
	Amount    float64
	Currency  string
	Status    models.LicensePaymentStatus
	Provider  string
	PaidAt    *time.Time
	Meta      string
}

func (j LicensePayment) Validate() error {
	if err := j.BaseSpaceModel.Validate(); err != nil {
		return err
	}
	if j.Status == "" {
		return errors.New("отсутсвует статус")
	}
	if j.Amount <= 0 {
		return errors.New("не указана сумма")
	}
	return nil
}

type LicensePlan struct {
	BaseModel
	Name                string
	Cost                float64
	ExtensionPeriodDays int
}

func (j LicensePlan) Validate() error {
	if j.Name == "" {
		return errors.New("отсутсвует название тарифа")
	}
	if j.Cost <= 0 {
		return errors.New("не указана стоимость продления лицензии")
	}
	if j.ExtensionPeriodDays <= 0 {
		return errors.New("не указан период продления лицензии")
	}
	return nil
}
