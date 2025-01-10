package pushhandler

import (
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	pushsettingsstore "hr-tools-backend/lib/space/push/settings-store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	"hr-tools-backend/models"
)

type Provider interface {
	SendNotification(userID string, code models.SpacePushSettingCode)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		spaceUserStore:    spaceusersstore.NewInstance(db.DB),
		pushSettingsStore: pushsettingsstore.NewInstance(db.DB),
	}
}

type impl struct {
	spaceUserStore    spaceusersstore.Provider
	pushSettingsStore pushsettingsstore.Provider
}

func (i *impl) getLogger(userID, code string) *log.Entry {
	logger := log.
		WithField("user_id", userID).
		WithField("event_code", code)
	return logger
}

func (i impl) SendNotification(userID string, code models.SpacePushSettingCode) {
	logger := i.getLogger(userID, string(code))
	user, err := i.spaceUserStore.GetByID(userID)
	if err != nil {
		logger.WithError(err).Error("ошибка получения пользователя")
		return
	}
	if user == nil {
		logger.Error("пользователь не найден")
		return
	}
	if !user.PushEnabled {
		return
	}
	pushSetting, err := i.pushSettingsStore.GetByCode(userID, code)
	if err != nil {
		logger.WithError(err).Error("ошибка получения настройки по событию")
		return
	}
	if pushSetting.SystemValue != nil && *pushSetting.SystemValue {
		//TODO send system push
	}
	if pushSetting.EmailValue != nil && *pushSetting.EmailValue {
		//TODO send email push
	}
	if pushSetting.TgValue != nil && *pushSetting.TgValue {
		//TODO send Tg push
	}
	// заголовок сообщения email
}
