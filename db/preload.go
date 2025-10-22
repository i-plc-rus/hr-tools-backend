package db

import (
	"hr-tools-backend/config"
	adminpaneluserstore "hr-tools-backend/lib/admin-panel/store"
	pushsettingsstore "hr-tools-backend/lib/space/push/settings-store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"

	log "github.com/sirupsen/logrus"
)

func InitPreload() {
	addSuperAdmin()
	fillCities()
	fillSpaceSettings()
	addPushSettings()
	fillLanguages()
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
