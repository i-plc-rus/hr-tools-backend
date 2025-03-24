package spacesettingshandler

import (
	"hr-tools-backend/db"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	"hr-tools-backend/models"
	spaceapimodels "hr-tools-backend/models/api/space"

	"github.com/pkg/errors"
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
	if !i.isUnique(spaceID, settingCode, settingValue) {
		return errors.New("значение настройки уже используется в другом спейсе")
	}
	err := i.spaceSettingsStore.Update(spaceID, settingCode, settingValue)
	if err != nil {
		return err
	}
	return nil
}

func (i impl) GetList(spaceID string) (settingsList []spaceapimodels.SpaceSettingView, err error) {
	list, err := i.spaceSettingsStore.List(spaceID)
	if err != nil {
		return nil, err
	}
	for _, setting := range list {
		settingsList = append(settingsList, setting.ToModelView())
	}
	return settingsList, nil
}

func (i impl) isUnique(settingSpaceID, settingCode, settingValue string) bool {
	if settingCode != string(models.HhClientIDSetting) &&
		settingCode != string(models.AvitoClientIDSetting) {
		return true
	}
	spaceID, err := i.spaceSettingsStore.GetSpaceIDByCodeAndValue(settingCode, settingValue)
	if err != nil {
		//TODO LOG
		return true
	}
	return spaceID == "" || settingSpaceID == spaceID
}
