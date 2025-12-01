package licenseapimodels

import (
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"

	"github.com/pkg/errors"
)

type License struct {
	ID              string               `json:"id"`
	Status          models.LicenseStatus `json:"status"`
	StartsAt        *time.Time           `json:"starts_at"`
	EndsAt          *time.Time           `json:"ends_at"`
	Plan            LicensePlan          `json:"plan"`
	AutoRenew       bool                 `json:"auto_renew"`
	LicensePayments []LicensePayment     `json:"license_payments"`
	DaysLeft        int                  `json:"days_left"`
}

type LicensePayment struct {
	ID       string                      `json:"id"`
	Amount   float64                     `json:"amount"`
	Currency string                      `json:"currency"`
	Status   models.LicensePaymentStatus `json:"status"`
	Provider string                      `json:"provider"`
	PaidAt   *time.Time                  `json:"paid_at"`
	Meta     string                      `json:"meta"`
}

type LicensePaymentFilter struct {
	Status *models.LicensePaymentStatus `json:"status"`
}

type LicensePlan struct {
	ID                  string  `json:"id"`
	Plan                string  `json:"plan"`
	Cost                float64 `json:"cost"`
	ExtensionPeriodDays int     `json:"extension_period_days"`
}

func LicenseConvert(rec *dbmodels.LicenseExt) License {
	result := License{
		ID:       rec.ID,
		Status:   rec.Status,
		StartsAt: rec.StartsAt,
		EndsAt:   rec.EndsAt,
		Plan: LicensePlan{
			ID:                  rec.PlanID,
			Plan:                rec.PlanName,
			Cost:                rec.PlanCost,
			ExtensionPeriodDays: rec.PlanPeriodDays,
		},
		AutoRenew:       rec.AutoRenew,
		LicensePayments: []LicensePayment{},
		DaysLeft:        0,
	}
	if result.EndsAt != nil && !result.EndsAt.IsZero() {
		endsAt := *result.EndsAt
		daysLeft := int(endsAt.Sub(time.Now()).Hours() / 24)
		if daysLeft >= 0 {
			result.DaysLeft = daysLeft
		}
	}
	for _, payment := range rec.LicensePayments {
		result.LicensePayments = append(result.LicensePayments, LicensePaymentConvert(payment))
	}
	return result
}

func LicensePaymentConvert(rec dbmodels.LicensePayment) LicensePayment {
	return LicensePayment{
		ID:       rec.ID,
		Amount:   rec.Amount,
		Currency: rec.Currency,
		Status:   rec.Status,
		Provider: rec.Provider,
		PaidAt:   rec.PaidAt,
		Meta:     rec.Meta,
	}
}

type LicenseRenew struct {
	OfferAccepter bool `json:"offer_accepter"`
}

func (l LicenseRenew) Validate() error {
	if !l.OfferAccepter {
		return errors.New("необходимо подтвердить согласие с офертой")
	}
	return nil
}

type LicenseRenewResponse struct {
	ID string `json:"id"`
}

type LicenseRenewInfo struct {
	StartsAt   *time.Time `json:"starts_at"`
	EndsAt     *time.Time `json:"ends_at"`
	Plan       string     `json:"plan"`
	Cost       float64    `json:"cost"`
	PeriodDays int        `json:"period_days"`
}

type LicenseRenewConfirm struct {
	ID       string     `json:"id"`
	Provider string     `json:"provider"`
	PaidAt   *time.Time `json:"paid_at"`
	Meta     string     `json:"meta"`
}
