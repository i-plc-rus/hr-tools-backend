package spacesettingshandler

import (
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	spaceapimodels "hr-tools-backend/models/api/space"
)

type Provider interface {
	UpdateSettingValue(spaceID, settingCode, settingValue string) error
	GetList(spaceID string) (settingsList []spaceapimodels.SpaceSettingView, err error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
	}
}

type impl struct {
	spaceSettingsStore spacesettingsstore.Provider
}

func (i impl) UpdateSettingValue(spaceID, settingCode, settingValue string) error {
	err := i.spaceSettingsStore.Update(spaceID, settingCode, settingValue)
	if err != nil {
		log.WithFields(log.Fields{
			"space_id":      spaceID,
			"setting_code":  settingCode,
			"setting_value": settingValue,
		}).
			WithError(err).
			Error("ошибка обновления настройки пространства")
		return err
	}
	return nil
}

func (i impl) GetList(spaceID string) (settingsList []spaceapimodels.SpaceSettingView, err error) {
	list, err := i.spaceSettingsStore.List(spaceID)
	if err != nil {
		log.
			WithField("space_id", spaceID).
			WithError(err).
			Error("ошибка получения списка настроек пространства")
		return nil, err
	}
	for _, setting := range list {
		settingsList = append(settingsList, setting.ToModelView())
	}
	return settingsList, nil
}
