package licenseworker

import (
	"context"
	"hr-tools-backend/db"
	licensestore "hr-tools-backend/lib/licence/store"
	baseworker "hr-tools-backend/lib/utils/base-worker"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/models"
	"time"
)

func StartWorker(ctx context.Context) {
	i := &impl{
		BaseImpl:     *baseworker.NewInstance("LicenseWorker", 15*time.Second, 60*time.Minute),
		licenseStore: licensestore.NewInstance(db.DB),
	}
	go i.Run(ctx, i.handle)
}

type impl struct {
	baseworker.BaseImpl
	licenseStore licensestore.Provider
}

func (i impl) handle(ctx context.Context) {
	// Получаем список лицензий для перевода в EXPIRES_SOON
	expiresSoonDate := time.Now().Add(time.Hour * 24 * 14)
	i.updateStatuses(ctx, expiresSoonDate, models.LicenseStatusActive, models.LicenseStatusExpiresSoon)

	// Получаем список лицензий для перевода в EXPIRED
	expiredDate := time.Now()
	i.updateStatuses(ctx, expiredDate, models.LicenseStatusExpiresSoon, models.LicenseStatusExpired)
}

func (i impl) updateStatuses(ctx context.Context, expireTime time.Time, currentStatus, newStatus models.LicenseStatus) {
	logger := i.GetLogger()
	list, err := i.licenseStore.ListToExpired(currentStatus, expireTime)
	if err != nil {
		logger.WithError(err).Errorf("Ошибка получения списка лицензий для перевода в %v", newStatus)
		return
	}
	for _, licence := range list {
		if helpers.IsContextDone(ctx) {
			break
		}
		updMap := map[string]interface{}{
			"Status": newStatus,
		}
		err = i.licenseStore.Update(licence.SpaceID, licence.ID, updMap)
		if err != nil {
			logger.
				WithError(err).
				WithField("space_id", licence.SpaceID).
				Errorf("Ошибка перевода статуса лицензии в %v", newStatus)
			continue
		}
	}
}
