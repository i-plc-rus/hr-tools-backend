package applicantsurveysuggestworker

import (
	"context"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	companystore "hr-tools-backend/lib/dicts/company/store"
	negotiationchathandler "hr-tools-backend/lib/external-services/negotiation-chat"
	messagetemplate "hr-tools-backend/lib/message-template"
	"hr-tools-backend/lib/smtp"
	applicantsurveystore "hr-tools-backend/lib/survey/applicant-survey-store"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"
	dbmodels "hr-tools-backend/models/db"
	"runtime/debug"
	"time"

	log "github.com/sirupsen/logrus"
)

// отправка ссылок на опрос кандидатам
func StartWorker(ctx context.Context) {
	i := &impl{
		applicantStore:         applicantstore.NewInstance(db.DB),
		negotiationChatHandler: negotiationchathandler.Instance,
		companyStore:           companystore.NewInstance(db.DB),
		messageTemplate:        messagetemplate.Instance,
		applicantSurveyStore:   applicantsurveystore.NewInstance(db.DB),
	}
	go i.run(ctx)
}

const (
	handlePeriod       = 5 * time.Minute
	defaultCompanyName = "HR-Tools"
)

type impl struct {
	applicantStore         applicantstore.Provider
	negotiationChatHandler negotiationchathandler.Provider
	companyStore           companystore.Provider
	messageTemplate        messagetemplate.Provider
	applicantSurveyStore   applicantsurveystore.Provider
}

func (i impl) getLogger() *log.Entry {
	logger := log.
		WithField("worker_name", "ApplicantSurveySuggestWorker")
	return logger
}

func (i impl) run(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			i.getLogger().
				WithField("panic_stack", string(debug.Stack())).
				Errorf("panic: (%v)", r)
		}
	}()
	period := 5 * time.Second
	logger := i.getLogger()
	for {
		select {
		// проверяем не завершён ли ещё контекст и выходим, если завершён
		case <-ctx.Done():
			logger.Info("Задача остановлена")
			return
		case <-time.After(period):
			logger.Info("Задача запущена")
			i.handle()
			logger.Info("Задача выполнена")
		}
		period = handlePeriod
	}
}

func (i impl) handle() {
	logger := i.getLogger()
	//Получаем список не отправленных анкет
	list, err := i.applicantStore.ListForSurveySend()
	if err != nil {
		logger.WithError(err).Error("ошибка получения списка не отправленных анкет")
		return
	}
	for _, applicant := range list {
		isSend := i.informApplicant(applicant)
		err = i.applicantSurveyStore.SetIsSend(applicant.ApplicantSurvey.ID, isSend)
		if err != nil {
			logger.WithError(err).Error("ошибка установки признака отправки анкеты")
		}
	}
}

func (i impl) informApplicant(applicantRec dbmodels.Applicant) (isSend bool) {
	spaceID := applicantRec.SpaceID
	applicantID := applicantRec.ID
	logger := log.
		WithField("space_id", spaceID).
		WithField("applicant_id", applicantID)

	isChatAvailable := false
	availability, err := i.negotiationChatHandler.IsVailable(spaceID, applicantID)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка проверки доступности чата с кандидатом")
	} else {
		isChatAvailable = availability.IsAvailable
	}

	companyName := i.getCompanyName(spaceID, applicantRec.Vacancy.CompanyID)
	link := config.Conf.UIParams.SurveyPath + applicantRec.ApplicantSurvey.ID
	if isChatAvailable {
		if i.sendToChat(spaceID, applicantID, companyName, link, logger) {
			isSend = true
		}
	}
	if applicantRec.Email != "" {
		emailFrom, err := i.messageTemplate.GetSenderEmail(spaceID)
		if err != nil {
			logger.
				WithError(err).
				Warn("ошибка получения почты компании для отправки сообщения с ссылкой на анкету на email кандидату")
			return isSend
		}
		if emailFrom == "" {
			return isSend
		}

		if i.sendToEmail(emailFrom, applicantRec.Email, companyName, link, logger) {
			isSend = true
		}
	}
	return isSend
}

func (i impl) sendToChat(spaceID, applicantID, companyName, link string, logger *log.Entry) bool {
	text, err := messagetemplate.GetSurvaySuggestMessage(companyName, link, false)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка получения сообщения с ссылкой на анкету для отправки кандидату через чат")
		return false
	}
	req := negotiationapimodels.NewMessageRequest{
		ApplicantID: applicantID,
		Text:        text,
	}
	err = i.negotiationChatHandler.SendMessage(spaceID, req)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка отправки сообщения с ссылкой на анкету в чат с кандидатом")
		return false
	}
	return true
}

func (i impl) sendToEmail(emailFrom, mailTo, companyName, link string, logger *log.Entry) bool {
	text, err := messagetemplate.GetSurvaySuggestMessage(companyName, link, true)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка получения сообщения с ссылкой на анкету для отправки кандидату через email")
		return false
	}
	title := messagetemplate.GetSurvaySuggestTitle()
	err = smtp.Instance.SendHtmlEMail(emailFrom, mailTo, text, title, nil)
	if err != nil {
		logger.
			WithError(err).
			Warn("ошибка отправки сообщения с ссылкой на анкету на email кандидату")
		return false
	}
	return true
}

func (i impl) getCompanyName(spaceID string, companyID *string) string {
	if companyID != nil {
		company, err := i.companyStore.GetByID(spaceID, *companyID)
		if err == nil && company != nil && company.Name != "" {
			return company.Name
		}
	}
	return defaultCompanyName
}
