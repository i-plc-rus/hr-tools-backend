package negotiationworker

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
	"runtime/debug"
	"time"
)

type NegotiationCheckJob interface {
	CheckConnected(ctx context.Context, spaceID string) bool
	GetCheckList(ctx context.Context, spaceID string, status models.VacancyPubStatus) ([]dbmodels.Vacancy, error)
	HandleNegotiations(ctx context.Context, data dbmodels.Vacancy) error
}

func StartWorker(ctx context.Context) {
	i := &impl{
		avito:      avitohandler.Instance.(NegotiationCheckJob),
		hh:         hhhandler.Instance.(NegotiationCheckJob),
		spaceStore: spacestore.NewInstance(db.DB),
	}
	go i.run(ctx, "HeadHunter", i.hh)
	go i.run(ctx, "Avito", i.avito)
}

const (
	handlePeriod = 5 * time.Minute
)

type impl struct {
	avito      NegotiationCheckJob
	hh         NegotiationCheckJob
	spaceStore spacestore.Provider
}

func (i impl) getLogger(integrationName string) *log.Entry {
	logger := log.
		WithField("integration", integrationName).
		WithField("worker_name", "NegotiationCheckJob")
	return logger
}

func (i impl) run(ctx context.Context, integrationName string, jobHandler NegotiationCheckJob) {
	defer func() {
		if r := recover(); r != nil {
			i.getLogger(integrationName).
				WithField("panic_stack", string(debug.Stack())).
				Errorf("panic: (%v)", r)
		}
	}()
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

func (i impl) handle(ctx context.Context, integrationName string, jobHandler NegotiationCheckJob) {
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
		if !jobHandler.CheckConnected(ctx, spaceID) {
			continue
		}

		list, err := jobHandler.GetCheckList(ctx, spaceID, models.VacancyPubStatusPublished)
		if err != nil {
			logger.
				WithError(err).
				Error("ошибка получения списка активных вакансий")
			return
		}
		for _, vacancy := range list {
			err = jobHandler.HandleNegotiations(ctx, vacancy)
			if err != nil {
				logger.
					WithError(err).
					WithField("vacancy_id", vacancy.ID).
					Error("ошибка получения откликов по вакансии")
				return
			}
		}
	}
}
