package pushhandler

import (
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	"hr-tools-backend/lib/smtp"
	pushdatastore "hr-tools-backend/lib/space/push/data-store"
	pushsettingsstore "hr-tools-backend/lib/space/push/settings-store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	initchecker "hr-tools-backend/lib/utils/init-checker"
	vacancyreqstore "hr-tools-backend/lib/vacancy-req/store"
	connectionhub "hr-tools-backend/lib/ws/hub/connection-hub"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	wsmodels "hr-tools-backend/models/ws"
	"time"

	log "github.com/sirupsen/logrus"
)

type Provider interface {
	SendNotification(userIDToSend string, data models.NotificationData)
}

var Instance Provider

func NewHandler() {
	instance := impl{
		spaceUserStore:    spaceusersstore.NewInstance(db.DB),
		pushSettingsStore: pushsettingsstore.NewInstance(db.DB),
		pushDataStore:     pushdatastore.NewInstance(db.DB),
		vrStore:           vacancyreqstore.NewInstance(db.DB),
		systemEmail:       config.Conf.Smtp.EmailSendVerification,
	}
	initchecker.CheckInit(
		"spaceUserStore", instance.spaceUserStore,
		"pushSettingsStore", instance.pushSettingsStore,
		"pushDataStore", instance.pushDataStore,
		"vrStore", instance.vrStore,
	)

	Instance = instance
}

type impl struct {
	spaceUserStore    spaceusersstore.Provider
	pushSettingsStore pushsettingsstore.Provider
	pushDataStore     pushdatastore.Provider
	vrStore           vacancyreqstore.Provider
	systemEmail       string
}

func (i *impl) getLogger(userID, code string) *log.Entry {
	logger := log.
		WithField("user_id", userID).
		WithField("event_code", code)
	return logger
}

func (i impl) SendNotification(userID string, data models.NotificationData) {
	logger := i.getLogger(userID, string(data.Code))
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
	pushSetting, err := i.pushSettingsStore.GetByCode(userID, data.Code)
	if err != nil {
		logger.WithError(err).Error("ошибка получения настройки по событию")
		return
	}
	if pushSetting.SystemValue != nil && *pushSetting.SystemValue {
		go i.sendToWs(userID, data)
	}
	if pushSetting.EmailValue != nil && *pushSetting.EmailValue && user.Email != "" && user.IsEmailVerified &&
		smtp.Instance.IsConfigured() {
		go i.sendToEmail(user.Email, data)
	}
	if pushSetting.TgValue != nil && *pushSetting.TgValue {
		//send Tg push
	}
}

func (i impl) sendToWs(userID string, data models.NotificationData) {
	logger := i.getLogger(userID, string(data.Code))
	if connectionhub.Instance.IsConnected(userID) {
		msg := wsmodels.ServerMessage{
			ToUserID: userID,
			Time:     time.Now().Format("02.01.2006 15:04:05"),
			Code:     string(data.Code),
			Msg:      data.Msg,
			Title:    data.Title,
		}
		connectionhub.Instance.SendMessage(msg)
	} else {
		rec := dbmodels.PushData{
			UserID: userID,
			Code:   data.Code,
			Msg:    data.Msg,
			Title:  data.Title,
		}
		err := i.pushDataStore.Create(rec)
		if err != nil {
			logger.WithError(err).Error("ошибка сохранения данных по событию")
			return
		}
	}
}

func (i impl) sendToEmail(email string, data models.NotificationData) {
	smtp.Instance.SendEMail(i.systemEmail, email, data.Msg, data.Title)
}
