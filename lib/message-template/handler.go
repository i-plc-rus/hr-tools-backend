package messagetemplate

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	messagetemplatestore "hr-tools-backend/lib/message-template/store"
	"hr-tools-backend/lib/smtp"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	"hr-tools-backend/models"
	msgtemplateapimodels "hr-tools-backend/models/api/message-template"
)

func NewHandler() {
	Instance = &impl{
		msgTemplateStore:   messagetemplatestore.NewInstance(db.DB),
		applicantStore:     applicantstore.NewInstance(db.DB),
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
	}
}

type Provider interface {
	SendEmailMessage(spaceID, templateID, applicantID string) error
	GetListTemplates(spaceID string) (list []msgtemplateapimodels.MsgTemplateView, err error)
}

type impl struct {
	msgTemplateStore   messagetemplatestore.Provider
	applicantStore     applicantstore.Provider
	spaceSettingsStore spacesettingsstore.Provider
}

func (i impl) SendEmailMessage(spaceID, templateID, applicantID string) error {
	logger := log.WithFields(log.Fields{
		"space_id":     spaceID,
		"template_id":  templateID,
		"applicant_id": applicantID,
	})

	email, err := i.spaceSettingsStore.GetValueByCode(spaceID, models.SpaceSenderEmail)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения почты для отправки из настроек пространства")
		return err
	}
	if email == "" {
		logger.
			Error("в настройках пространства не указана почта для отправки")
		return errors.New("в настройках пространства не указана почта для отправки")
	}

	msgTemplate, err := i.msgTemplateStore.GetByID(spaceID, templateID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка поиска шаблона сообщения по ID")
		return err
	}
	if msgTemplate == nil {
		logger.
			Error("не найден шаблон сообщения по указанномму ID")
		return errors.New("не найден шаблон сообщения по указанномму ID")
	}

	applicant, err := i.applicantStore.GetByID(spaceID, applicantID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка поиска кандидата по ID")
		return err
	}
	if applicant == nil {
		logger.
			Error("не найден кандидат по указанномму ID")
		return errors.New("не найден кандидат по указанномму ID")
	}

	if applicant.Email == "" {
		logger.
			Error("у кандидата не указана почта")
		return errors.New("у кандидата не указана почта")
	}

	err = smtp.Instance.SendEMail(email, applicant.Email, msgTemplate.Message, msgTemplate.Title)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка отправки почты кандидату")
		return err
	}
	return nil
}

func (i impl) GetListTemplates(spaceID string) (list []msgtemplateapimodels.MsgTemplateView, err error) {
	logger := log.WithField("space_id", spaceID)
	recList, err := i.msgTemplateStore.List(spaceID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения списка шаблонов сообщения")
		return nil, err
	}
	for _, template := range recList {
		list = append(list, template.ToModel())
	}
	return list, nil
}

var Instance Provider
