package newmsgworker

import (
	"context"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	externalservices "hr-tools-backend/lib/external-services"
	avitohandler "hr-tools-backend/lib/external-services/avito"
	hhhandler "hr-tools-backend/lib/external-services/hh"
	extservicestore "hr-tools-backend/lib/external-services/store"
	pushhandler "hr-tools-backend/lib/space/push/handler"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

func StartWorker(ctx context.Context) {
	i := &impl{
		avito:          avitohandler.Instance.(externalservices.JobSiteProvider),
		hh:             hhhandler.Instance.(externalservices.JobSiteProvider),
		applicantStore: applicantstore.NewInstance(db.DB),
		extStore:       extservicestore.NewInstance(db.DB),
		vacancyStore:   vacancystore.NewInstance(db.DB),
	}
	go i.run(ctx, "HeadHunter", i.hh)
	go i.run(ctx, "Avito", i.avito)
}

const (
	handlePeriod = 5 * time.Minute
)

type impl struct {
	avito          externalservices.JobSiteProvider
	hh             externalservices.JobSiteProvider
	applicantStore applicantstore.Provider
	extStore       extservicestore.Provider
	vacancyStore   vacancystore.Provider
}

func (i impl) getLogger(integrationName string) *log.Entry {
	logger := log.
		WithField("integration", integrationName).
		WithField("worker_name", "NewMsgCheckJob")
	return logger
}

func (i impl) run(ctx context.Context, integrationName string, provider externalservices.JobSiteProvider) {
	period := time.Second
	logger := i.getLogger(integrationName)
	for {
		select {
		// проверяем не завершён ли ещё контекст и выходим, если завершён
		case <-ctx.Done():
			logger.Info("Задача остановлена")
			return
		case <-time.After(period):
			logger.Info("Задача запущена")
			i.handle(ctx, integrationName, provider)
			logger.Info("Задача выполнена")
		}
		period = handlePeriod
	}
}

func (i impl) handle(ctx context.Context, integrationName string, provider externalservices.JobSiteProvider) {
	logger := i.getLogger(integrationName)
	list, err := i.applicantStore.ListOfActiveApplicants()
	if err != nil {
		logger.WithError(err).Error("ошибка получения списка активных кандидатов")
		return
	}
	for _, applicant := range list {
		if !provider.CheckConnected(applicant.SpaceID) {
			continue
		}
		msg, err := provider.GetLastInMessage(ctx, applicant)
		if err != nil {
			logger.WithError(err).
				WithField("space_id", applicant.SpaceID).
				WithField("vacancy_id", applicant.VacancyID).
				WithField("applicant_id", applicant.ID).
				Error("ошибка получения последнего сообщения от кандидата")
			continue
		}
		if msg == nil || msg.SelfMessage {
			continue
		}
		b, ok, err := i.extStore.Get(applicant.SpaceID, getChatKey(applicant.ID))
		if err != nil {
			logger.WithError(err).
			WithField("space_id", applicant.SpaceID).
			WithField("vacancy_id", applicant.VacancyID).
			WithField("applicant_id", applicant.ID).
			Warn("не удалось получить дату последнего обновления")
		} else if ok {
			lastMsgTime := time.Unix(int64(binary.LittleEndian.Uint64(b)), 0)
			if !msg.MessageDateTime.After(lastMsgTime) {
				continue
			}
		}
		//новое сообщение
		b = make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(msg.MessageDateTime.Unix()))
		i.extStore.Set(applicant.SpaceID, getChatKey(applicant.ID), b)

		notification := models.GetPushApplicantMsg(applicant.Vacancy.VacancyName, applicant.GetFIO(), integrationName)
		go func(spaceID, vacancyID, integrationName string, nData models.NotificationData) {
			i.newMsg(spaceID, vacancyID, integrationName, nData)
		}(applicant.SpaceID, applicant.VacancyID, integrationName, notification)
	}
}

func (i impl) newMsg(spaceID, vacancyID, integrationName string, data models.NotificationData) {
	logger := i.getLogger(integrationName).
		WithField("space_id", spaceID).
		WithField("vacancy_id", vacancyID)
	vacancy := i.getVacancy(spaceID, vacancyID, logger)
	if vacancy == nil {
		return
	}
	//отправляем автору
	pushhandler.Instance.SendNotification(vacancy.AuthorID, data)
	for _, teamMember := range vacancy.VacancyTeam {
		//отправляем команде
		if vacancy.AuthorID == teamMember.ID {
			continue
		}
		pushhandler.Instance.SendNotification(teamMember.ID, data)
	}
}

func getChatKey(applicantID string) string {
	return fmt.Sprintf("JOB_CHAT:%v", applicantID)
}

func (i *impl) getVacancy(spaceID, vacancyID string, logger *log.Entry) *dbmodels.Vacancy {
	vacancyRec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		logger.WithError(err).Error("ошибка получения вакансии")
		return nil
	}
	if vacancyRec == nil {
		logger.Error("вакансия не найдена")
		return nil
	}
	return vacancyRec
}
