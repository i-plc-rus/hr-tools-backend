package messagetemplate

import (
	"fmt"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	messagetemplatestore "hr-tools-backend/lib/message-template/store"
	"hr-tools-backend/lib/smtp"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	"hr-tools-backend/models"
	applicantapimodels "hr-tools-backend/models/api/applicant"
	msgtemplateapimodels "hr-tools-backend/models/api/message-template"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	SendEmailMessage(spaceID, templateID, applicantID, userID string) error
	GetListTemplates(spaceID string) (list []msgtemplateapimodels.MsgTemplateView, err error)
	MultiSendEmail(spaceID, userID string, data applicantapimodels.MultiEmailRequest) (failMails []string, err error)
}

var Instance Provider

func NewHandler() {
	Instance = &impl{
		msgTemplateStore:   messagetemplatestore.NewInstance(db.DB),
		applicantStore:     applicantstore.NewInstance(db.DB),
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
		spaceUsersStore:    spaceusersstore.NewInstance(db.DB),
	}
}

type impl struct {
	msgTemplateStore   messagetemplatestore.Provider
	applicantStore     applicantstore.Provider
	spaceSettingsStore spacesettingsstore.Provider
	spaceUsersStore    spaceusersstore.Provider
}

func (i impl) SendEmailMessage(spaceID, templateID, applicantID, userID string) error {
	logger := log.WithFields(log.Fields{
		"space_id":     spaceID,
		"template_id":  templateID,
		"applicant_id": applicantID,
	})

	email, err := i.getSenderEmail(spaceID, logger)
	if err != nil {
		return err
	}

	msgTemplate, err := i.getMsgTemplate(spaceID, templateID, logger)
	if err != nil {
		return err
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
	user, err := i.spaceUsersStore.GetByID(userID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения профиля отправителя")
		return errors.New("ошибка получения профиля отправителя")
	}
	textSign := ""
	if user != nil {
		textSign = user.TextSign
	}
	msg := buildMsg(msgTemplate.Message, textSign)

	err = smtp.Instance.SendEMail(email, applicant.Email, msg, msgTemplate.Title)
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

func (i impl) MultiSendEmail(spaceID, userID string, data applicantapimodels.MultiEmailRequest) (failMails []string, err error) {
	logger := log.WithFields(log.Fields{
		"space_id":    spaceID,
		"template_id": data.MsgTemplateID,
		"user_id":     userID,
	})
	if !smtp.Instance.IsConfigured() {
		logger.
			WithError(err).
			Error("smtp клиент не настроен")
		return nil, errors.New("smtp клиент не настроен")
	}

	email, err := i.getSenderEmail(spaceID, logger)
	if err != nil {
		return nil, err
	}

	msgTemplate, err := i.getMsgTemplate(spaceID, data.MsgTemplateID, logger)
	if err != nil {
		return nil, err
	}

	user, err := i.spaceUsersStore.GetByID(userID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения профиля отправителя")
		return nil, errors.New("ошибка получения профиля отправителя")
	}
	textSign := ""
	if user != nil {
		textSign = user.TextSign
	}
	msg := buildMsg(msgTemplate.Message, textSign)

	applicantList, err := i.applicantStore.ListOfApplicantByIDs(spaceID, data.IDs, nil)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения списка кандидатов")
		return nil, err
	}
	failMails = []string{}
	for _, applicant := range applicantList {
		fio := applicant.GetFIO()
		if applicant.Email == "" {
			logger.
				WithField("applicant_id", applicant.ID).
				Warn("у кандидата не указана почта")
			failMails = append(failMails, fio)
			continue
		}
		err = smtp.Instance.SendEMail(email, applicant.Email, msg, msgTemplate.Title)
		if err != nil {
			logger.
				WithError(err).
				Error("ошибка отправки почты кандидату")
			failMails = append(failMails, fio)
			continue
		}
	}

	return failMails, nil
}

func buildMsg(tmpl string, emailTextSign string) string {
	// возможно в будущем будем искать место вставки подписи и заполнение шаблона,
	// пока подпись добавляем в конец
	// если будеи использвать не только text/plain но и text/html, надо будет менять перенос на <br> для последнего
	if emailTextSign == "" {
		return tmpl
	}
	return fmt.Sprintf("%v\r\n%v", tmpl, emailTextSign)
}

func (i impl) getSenderEmail(spaceID string, logger *log.Entry) (string, error) {
	email, err := i.spaceSettingsStore.GetValueByCode(spaceID, models.SpaceSenderEmail)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения почты для отправки из настроек пространства")
		return "", errors.New("ошибка получения почты для отправки из настроек пространства")
	}
	if email == "" {
		logger.
			Error("в настройках пространства не указана почта для отправки")
		return "", errors.New("в настройках пространства не указана почта для отправки")
	}
	return email, nil
}

func (i impl) getMsgTemplate(spaceID, templateID string, logger *log.Entry) (*dbmodels.MessageTemplate, error) {
	msgTemplate, err := i.msgTemplateStore.GetByID(spaceID, templateID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения шаблона сообщения")
		return nil, errors.New("ошибка получения шаблона сообщения")
	}
	if msgTemplate == nil {
		logger.
			Error("шаблон сообщения не найден")
		return nil, errors.New("шаблон сообщения не найден")
	}
	return msgTemplate, nil
}
