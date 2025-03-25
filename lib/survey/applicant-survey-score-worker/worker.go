package applicantsurveyscoreworker

import (
	"context"
	"hr-tools-backend/db"
	"hr-tools-backend/lib/survey"
	applicantsurveystore "hr-tools-backend/lib/survey/applicant-survey-store"
	"time"

	log "github.com/sirupsen/logrus"
)

// оценка кандидата
func StartWorker(ctx context.Context) {
	i := &impl{
		applicantSurveyStore: applicantsurveystore.NewInstance(db.DB),
		survey:               survey.Instance,
	}
	go i.run(ctx)
}

const (
	handlePeriod = 5 * time.Minute
)

type impl struct {
	applicantSurveyStore applicantsurveystore.Provider
	survey               survey.Provider
}

func (i impl) getLogger() *log.Entry {
	logger := log.
		WithField("worker_name", "ApplicantSurveyScoreWorker")
	return logger
}

func (i impl) run(ctx context.Context) {
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
	//Получаем список анкет кандидатов для оценки
	list, err := i.applicantSurveyStore.GetSurveyForScore()
	if err != nil {
		logger.WithError(err).Error("ошибка получения списка анкет кандидатов для оценки")
		return
	}
	for _, survey := range list {
		ok, err := i.survey.AIScore(survey)
		if err != nil {
			logger.WithError(err).
				WithField("space_id", survey.SpaceID).
				WithField("applicant_id", survey.ApplicantID).
				Error("ошибка оценки кандидата")
			continue
		}
		if ok {
			logger.
				WithField("space_id", survey.SpaceID).
				WithField("applicant_id", survey.ApplicantID).
				Info("произведена оценка кандидата")
		}
	}
}
