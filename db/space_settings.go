package db

import (
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	spacestore "hr-tools-backend/lib/space/store"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"

	log "github.com/sirupsen/logrus"
)

func fillSpaceSettings() {
	log.Info("предзаполнение дефолтных настроек")
	spaceStore := spacestore.NewInstance(DB)
	settingsStore := spacesettingsstore.NewInstance(DB)
	ids, err := spaceStore.GetActiveIds()
	if err != nil {
		log.WithError(err).Error("ошибка получения активных спейсов")
		return
	}
	for _, spaceID := range ids {
		spaceSettings, err := settingsStore.List(spaceID)
		if err != nil {
			log.WithError(err).
				WithField("space_id", spaceID).
				Error("ошибка получения настроек спейса")
			continue
		}
		for code, spaceSettingData := range dbmodels.DefaultSettinsMap {
			err = checkAndCreateSetting(spaceID, spaceSettings, code, spaceSettingData, settingsStore)
			if err != nil {
				log.WithError(err).
					WithField("space_id", spaceID).
					WithField("setting_code", code).
					Error("ошибка добавления настройки")
				continue
			}
		}
		removeDuplicate(spaceID, spaceSettings, settingsStore)
	}
	log.Info("предзаполнение дефолтных настроек завершено")
}

func checkAndCreateSetting(spaceID string, spaceSettings []dbmodels.SpaceSetting, code models.SpaceSettingCode, spaceSettingData dbmodels.SpaceSetting, settingsStore spacesettingsstore.Provider) error {
	for _, setting := range spaceSettings {
		if setting.Code == code {
			return nil
		}
	}
	//не найдена
	spaceSettingData.SpaceID = spaceID
	return settingsStore.Create(spaceSettingData)
}


//TODO можно удалить просле первого деплоя
func removeDuplicate(spaceID string, spaceSettings []dbmodels.SpaceSetting, settingsStore spacesettingsstore.Provider) {
	codeMap := map[models.SpaceSettingCode]bool{}
	for _, setting := range spaceSettings {
		if codeMap[setting.Code] {
			err := settingsStore.Delete(spaceID, string(setting.Code))
			if err != nil {
				log.WithError(err).
					WithField("space_id", spaceID).
					WithField("setting_code", setting.Code).
					Error("ошибка удаления дубля настройки")
				continue
			}
		} else {
			codeMap[setting.Code] = true
		}
	}
}
