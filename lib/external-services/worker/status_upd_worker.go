package externalserviceworker

import (
	"context"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	avitohandler "hr-tools-backend/lib/external-services/avito"
	hhhandler "hr-tools-backend/lib/external-services/hh"
	spacestore "hr-tools-backend/lib/space/store"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

type StatusCheckJob interface {
	CheckConnected(spaceID string) bool
	GetCheckList(ctx context.Context, spaceID string, status models.VacancyPubStatus) ([]dbmodels.Vacancy, error)
	CheckIsModerationDone(ctx context.Context, spaceID string, list []dbmodels.Vacancy) error
	CheckIsActivePublications(ctx context.Context, spaceID string, list []dbmodels.Vacancy) error
}

func StartWorker(ctx context.Context) {
	i := &impl{
		avito:      avitohandler.Instance.(StatusCheckJob),
		hh:         hhhandler.Instance.(StatusCheckJob),
		spaceStore: spacestore.NewInstance(db.DB),
	}
	go i.run(ctx, "HeadHunter", i.hh)
	go i.run(ctx, "Avito", i.avito)
}

const (
	handlePeriod = 5 * time.Minute
)

type impl struct {
	avito      StatusCheckJob
	hh         StatusCheckJob
	spaceStore spacestore.Provider
}

func (i impl) getLogger(integrationName string) *log.Entry {
	logger := log.
		WithField("integration", integrationName).
		WithField("worker_name", "StatusCheckJob")
	return logger
}

func (i impl) run(ctx context.Context, integrationName string, jobHandler StatusCheckJob) {
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
			i.handle(ctx, integrationName, jobHandler)
			logger.Info("Задача выполнена")
		}
		period = handlePeriod
	}
}

func (i impl) handle(ctx context.Context, integrationName string, jobHandler StatusCheckJob) {
	logger := i.getLogger(integrationName)
	ids, err := i.spaceStore.GetActiveIds()
	if err != nil {
		logger.WithError(err).Error("ошибка получения списка активных спейсов")
		return
	}
	for _, spaceID := range ids {
		if helpers.IsContextDone(ctx) {
			return
		}
		logger = logger.WithField("space_id", spaceID)
		if !jobHandler.CheckConnected(spaceID) {
			logger.Debug("Спейс не подключен к интеграции")
			continue
		}

		list, err := jobHandler.GetCheckList(ctx, spaceID, models.VacancyPubStatusModeration)
		if err != nil {
			logger.
				WithError(err).
				Error("ошибка получения списка вакансий на модерации")
			return
		}
		if len(list) != 0 {
			err = jobHandler.CheckIsModerationDone(ctx, spaceID, list)
			if err != nil {
				logger.
					WithError(err).
					Error("ошибка проверки списка вакансий на модерации")
				return
			}
		}

		list, err = jobHandler.GetCheckList(ctx, spaceID, models.VacancyPubStatusPublished)
		if err != nil {
			logger.
				WithError(err).
				Error("ошибка получения списка активных вакансий")
			return
		}
		if len(list) != 0 {
			err = jobHandler.CheckIsActivePublications(ctx, spaceID, list)
			if err != nil {
				logger.
					WithError(err).
					Error("ошибка проверки списка активных вакансий")
				return
			}
		}
	}
}
