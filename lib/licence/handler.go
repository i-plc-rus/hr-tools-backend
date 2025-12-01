package licencehandler

import (
	"hr-tools-backend/db"
	licensepaymentstore "hr-tools-backend/lib/licence/payment-store"
	licenseplanstore "hr-tools-backend/lib/licence/plan-store"
	licensestore "hr-tools-backend/lib/licence/store"
	"hr-tools-backend/models"
	licenseapimodels "hr-tools-backend/models/api/license"
	dbmodels "hr-tools-backend/models/db"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Provider interface {
	GetSpaceLicense(spaceID string) (result licenseapimodels.License, err error)
	GetRenewSpaceLicense(spaceID string) (result licenseapimodels.LicenseRenewInfo, err error)
	RenewSpaceLicense(spaceID string) (result licenseapimodels.LicenseRenewResponse, err error)
	ConfirmPayment(reqest licenseapimodels.LicenseRenewConfirm, userID string) (hMsg string, err error)
}

var Instance Provider

func NewHandler() {
	Instance = &impl{
		licenseStore:        licensestore.NewInstance(db.DB),
		licensePaymentStore: licensepaymentstore.NewInstance(db.DB),
		licensePlanStore:    licenseplanstore.NewInstance(db.DB),
	}
}

type impl struct {
	licenseStore        licensestore.Provider
	licensePaymentStore licensepaymentstore.Provider
	licensePlanStore    licenseplanstore.Provider
}

func (i *impl) getLogger(spaceID, userID string) *log.Entry {
	logger := log.WithField("space_id", spaceID)
	if userID != "" {
		logger = logger.WithField("user_id", userID)
	}
	return logger
}

func (i *impl) GetSpaceLicense(spaceID string) (result licenseapimodels.License, err error) {
	rec, err := i.licenseStore.GetBySpaceExt(spaceID)
	if err != nil {
		return licenseapimodels.License{}, err
	}
	if rec == nil {
		return licenseapimodels.License{}, errors.New("лицензия не найдена")
	}
	return licenseapimodels.LicenseConvert(rec), nil
}

func (i *impl) GetRenewSpaceLicense(spaceID string) (result licenseapimodels.LicenseRenewInfo, err error) {
	license, err := i.licenseStore.GetBySpaceExt(spaceID)
	if err != nil {
		return licenseapimodels.LicenseRenewInfo{}, err
	}
	if license == nil {
		return licenseapimodels.LicenseRenewInfo{}, errors.New("лицензия не найдена")
	}

	if license.EndsAt == nil {
		return licenseapimodels.LicenseRenewInfo{}, errors.New("не указан период окончания лицензии")
	}
	endAt := license.EndsAt.Add(time.Hour * 24 * time.Duration(license.PlanPeriodDays))
	return licenseapimodels.LicenseRenewInfo{
		StartsAt:   license.EndsAt,
		EndsAt:     &endAt,
		Plan:       license.PlanName,
		Cost:       license.PlanCost,
		PeriodDays: license.PlanPeriodDays,
	}, nil
}

func (i *impl) RenewSpaceLicense(spaceID string) (result licenseapimodels.LicenseRenewResponse, err error) {
	license, err := i.licenseStore.GetBySpaceExt(spaceID)
	if err != nil {
		return licenseapimodels.LicenseRenewResponse{}, err
	}
	if license == nil {
		return licenseapimodels.LicenseRenewResponse{}, errors.New("лицензия не найдена")
	}
	// ищем существующие черновики
	payPendingStatus := models.LicensePaymentStatusPending
	paymentList, err := i.licensePaymentStore.List(spaceID, licenseapimodels.LicensePaymentFilter{
		Status: &payPendingStatus,
	})
	if err != nil {
		return licenseapimodels.LicenseRenewResponse{}, err
	}
	for _, payment := range paymentList {
		if payment.LicenseID == license.ID && payment.Amount == license.PlanCost {
			return licenseapimodels.LicenseRenewResponse{
				ID: payment.ID,
			}, nil
		}
	}

	payment := dbmodels.LicensePayment{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		LicenseID: license.ID,
		Amount:    license.PlanCost,
		Currency:  "RUR",
		Status:    models.LicensePaymentStatusPending,
		Provider:  "",
		PaidAt:    nil,
		Meta:      "",
	}
	id, err := i.licensePaymentStore.Create(payment)
	if err != nil {
		return licenseapimodels.LicenseRenewResponse{}, errors.Wrap(err, "ошибка создания черновика продления")
	}
	return licenseapimodels.LicenseRenewResponse{
		ID: id,
	}, nil
}

func (i *impl) ConfirmPayment(reqest licenseapimodels.LicenseRenewConfirm, userID string) (hMsg string, err error) {
	payRec, err := i.licensePaymentStore.GetByID(reqest.ID)
	if err != nil {
		return "", err
	}
	if payRec.Status == models.LicensePaymentStatusPaid {
		return "Платеж уже подтвержден", nil
	}
	licRec, err := i.licenseStore.GetByID(payRec.SpaceID, payRec.LicenseID)
	if err != nil {
		return "", err
	}
	planRec, err := i.licensePlanStore.GetByName(licRec.Plan)
	if err != nil {
		return "", err
	}

	logger := i.getLogger(payRec.SpaceID, userID).
		WithField("payment_id", payRec.ID).
		WithField("license_id", payRec.LicenseID)

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		licensePaymentStore := licensepaymentstore.NewInstance(tx)
		licenseStore := licensestore.NewInstance(db.DB)
		payUpdMap := map[string]interface{}{
			"Status":   models.LicensePaymentStatusPaid,
			"PaidAt":   reqest.PaidAt,
			"Provider": reqest.Provider,
			"Meta":     reqest.Meta,
		}
		err = licensePaymentStore.Update(payRec.ID, payRec.Status, payUpdMap)
		if err != nil {
			return err
		}
		endAt := time.Now()
		if licRec.EndsAt != nil {
			endAt = *licRec.EndsAt
		}
		endAt = endAt.Add(time.Hour * 24 * time.Duration(planRec.ExtensionPeriodDays))
		licUpdMap := map[string]interface{}{
			"Status": models.LicenseStatusActive,
			"EndsAt": endAt,
		}
		return licenseStore.Update(licRec.SpaceID, licRec.ID, licUpdMap)
	})
	if err != nil {
		return "", err
	}
	logger.Info("Платеж подтвержден администратором")
	return "", nil
}
