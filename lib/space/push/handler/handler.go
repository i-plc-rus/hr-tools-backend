package pushhandler

import (
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	"hr-tools-backend/lib/smtp"
	pushdatastore "hr-tools-backend/lib/space/push/data-store"
	pushsettingsstore "hr-tools-backend/lib/space/push/settings-store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	connectionhub "hr-tools-backend/lib/ws/hub/connection-hub"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	wsmodels "hr-tools-backend/models/ws"
	"time"

	log "github.com/sirupsen/logrus"
)

type Provider interface {
	SendNotification(userID string, code models.SpacePushSettingCode)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		spaceUserStore:    spaceusersstore.NewInstance(db.DB),
		pushSettingsStore: pushsettingsstore.NewInstance(db.DB),
		pushDataStore:     pushdatastore.NewInstance(db.DB),
		systemEmail:       config.Conf.Smtp.EmailSendVerification,
	}
}

type impl struct {
	spaceUserStore    spaceusersstore.Provider
	pushSettingsStore pushsettingsstore.Provider
	pushDataStore     pushdatastore.Provider
	systemEmail       string
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
		go i.sendToWs(userID, code)
	}
	if pushSetting.EmailValue != nil && *pushSetting.EmailValue && user.Email != "" && user.IsEmailVerified &&
		smtp.Instance.IsConfigured() {
		go i.sendToEmail(user.Email, code)
	}
	if pushSetting.TgValue != nil && *pushSetting.TgValue {
		//send Tg push
	}
}

func (i impl) sendToWs(userID string, code models.SpacePushSettingCode) {
	logger := i.getLogger(userID, string(code))
	if connectionhub.Instance.IsConnected(userID) {
		msg := wsmodels.ServerMessage{
			ToUserID: userID,
			Time:     time.Now().Format("02.01.2006 15:04:05"),
			Code:     string(code),
			Msg:      models.PushCodeMap[code].Msg,
		}
		connectionhub.Instance.SendMessage(msg)
	} else {
		rec := dbmodels.PushData{
			UserID: userID,
			Code:   code,
			Msg:    models.PushCodeMap[code].Msg,
		}
		err := i.pushDataStore.Create(rec)
		if err != nil {
			logger.WithError(err).Error("ошибка сохранения данных по событию")
			return
		}
	}
}

func (i impl) sendToEmail(email string, code models.SpacePushSettingCode) {
	smtp.Instance.SendEMail(i.systemEmail, email, models.PushCodeMap[code].Msg, "Заголовок сообщения") //TODO fix // заголовок сообщения email
}
