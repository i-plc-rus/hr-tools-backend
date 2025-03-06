package applicantsurveyworker

import (
	"context"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	"hr-tools-backend/lib/survey"
	applicantsurveystore "hr-tools-backend/lib/survey/applicant-survey-store"
	"time"

	log "github.com/sirupsen/logrus"
)

// генерация опросов по кандидату
func StartWorker(ctx context.Context) {
	i := &impl{
		applicantStore:       applicantstore.NewInstance(db.DB),
		applicantSurveyStore: applicantsurveystore.NewInstance(db.DB),
		survey:               survey.Instance,
	}
	go i.run(ctx)
}

const (
	handlePeriod = 5 * time.Minute
)

type impl struct {
	applicantStore       applicantstore.Provider
	applicantSurveyStore applicantsurveystore.Provider
	survey               survey.Provider
}

func (i impl) getLogger() *log.Entry {
	logger := log.
		WithField("worker_name", "ApplicantSurveyWorker")
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
	//Получаем список откликов по вакансиям с заполненной анкетой HR
	list, err := i.applicantStore.ListOfActivefNegotiation(true)
	if err != nil {
		logger.WithError(err).Error("ошибка получения списка активных кандидатов")
		return
	}
	for _, applicant := range list {
		if applicant.ApplicantSurvey == nil {
			ok, err := i.survey.GenApplicantSurvey(applicant.SpaceID, applicant.VacancyID, applicant.ID)
			if err != nil {
				logger.WithError(err).
					WithField("space_id", applicant.SpaceID).
					WithField("vacancy_id", applicant.VacancyID).
					WithField("applicant_id", applicant.ID).
					Error("ошибка генерации анкеты для кандидата")
				continue
			}
			if ok {
				logger.
					WithField("space_id", applicant.SpaceID).
					WithField("vacancy_id", applicant.VacancyID).
					WithField("applicant_id", applicant.ID).
					Info("анкеты для кандидата сгенерирована")
			}
		}
	}
}
