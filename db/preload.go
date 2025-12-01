package db

import (
	"hr-tools-backend/config"
	adminpaneluserstore "hr-tools-backend/lib/admin-panel/store"
	licenseplanstore "hr-tools-backend/lib/licence/plan-store"
	licensestore "hr-tools-backend/lib/licence/store"
	pushsettingsstore "hr-tools-backend/lib/space/push/settings-store"
	spacestore "hr-tools-backend/lib/space/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"

	log "github.com/sirupsen/logrus"
)

func InitPreload() {
	addSuperAdmin()
	fillCities()
	fillSpaceSettings()
	addPushSettings()
	fillLanguages()
	addBaseLicensePlan()
	addDefLicense()
}

func addSuperAdmin() {
	if config.Conf.Admin.Email == "" {
		log.Warn("суперадмин не добавлен, отсутвует настройка ADMIN_EMAIL")
		return
	}
	adminStore := adminpaneluserstore.NewInstance(DB)
	existedRec, err := adminStore.FindByEmail(config.Conf.Admin.Email)
	if err != nil {
		log.WithError(err).Error("ошибка добавления суперадмина")
		return
	}
	if existedRec != nil {
		return
	}
	rec := dbmodels.AdminPanelUser{
		IsActive:    true,
		Role:        models.UserRoleSuperAdmin,
		Password:    authutils.GetMD5Hash(config.Conf.Admin.Password),
		FirstName:   config.Conf.Admin.FirstName,
		LastName:    config.Conf.Admin.LastName,
		Email:       config.Conf.Admin.Email,
		PhoneNumber: config.Conf.Admin.PhoneNumber,
	}
	_, err = adminStore.Create(rec)
	if err != nil {
		log.WithError(err).Error("ошибка добавления суперадмина")
	}
}

func addPushSettings() {
	store := pushsettingsstore.NewInstance(DB)

	userList, err := store.GetUsersWithoutSettings()
	if err != nil {
		log.WithError(err).Error("ошибка добавления настроек пушей")
		return
	}

	value := false
	rec := dbmodels.SpacePushSetting{
		SystemValue: &value,
		EmailValue:  &value,
		TgValue:     &value,
	}
	for _, user := range userList {
		rec.SpaceID = user.SpaceID
		rec.SpaceUserID = user.ID

		for key := range models.PushCodeMap {
			rec.Code = key
			err := store.Create(rec)
			if err != nil {
				log.WithError(err).Error("ошибка добавления настроек пушей")
				return
			}
		}
	}
}

func addBaseLicensePlan() {
	log.Info("предзаполнение базового плана для лицензий")
	if config.Conf.Sales.DefaultPlan == "" {
		log.Warn("Базовый план не добавлен, отсутвует настройка SALES_DEF_PLAN")
		return
	}
	store := licenseplanstore.NewInstance(DB)
	store.Create(dbmodels.LicensePlan{
		Name:                config.Conf.Sales.DefaultPlan,
		Cost:                10000,
		ExtensionPeriodDays: 30,
	})
}

func addDefLicense() {
	log.Info("добавление лицензий для организаций")
	if config.Conf.Sales.DefaultPlan == "" {
		log.Warn("ошибка установки лицензий, отсутвует настройка SALES_DEF_PLAN")
		return
	}

	store := licensestore.NewInstance(DB)
	spaceStore := spacestore.NewInstance(DB)
	spaceIds, err := spaceStore.GetActiveIds()
	if err != nil {
		log.WithError(err).Error("ошибка добавления лицензий для организаций")
		return
	}

	now := time.Now()
	endAt := now.Add(time.Hour * 24 * 7)
	plan := config.Conf.Sales.DefaultPlan
	for _, id := range spaceIds {
		ok, err := store.IsExist(id)
		if err != nil {
			log.
				WithError(err).
				WithField("space_id", id).
				Error("ошибка проверки наличия лицензии у организации")
			return
		}
		if ok {
			continue
		}
		rec := dbmodels.License{
			BaseSpaceModel: dbmodels.BaseSpaceModel{
				SpaceID: id,
			},
			Status:    models.LicenseStatusActive,
			StartsAt:  &now,
			EndsAt:    &endAt,
			Plan:      plan,
			AutoRenew: false,
		}

		_, err = store.Create(rec)
		if err != nil {
			log.
				WithError(err).
				WithField("space_id", id).
				Error("ошибка добавления лицензии для организации")
			return
		}
	}
}
