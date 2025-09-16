package vkstep1runworker

import (
	"context"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	"hr-tools-backend/lib/vk"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"runtime/debug"
	"time"
)

// Задача ВК. Шаг 1. Генерация черновика скрипта
func StartWorker(ctx context.Context) {
	i := &impl{
		applicantStore: applicantstore.NewInstance(db.DB),
		vkStore:        applicantvkstore.NewInstance(db.DB),
	}
	go i.run(ctx)
}

const (
	handlePeriod = 5 * time.Minute
)

type impl struct {
	applicantStore applicantstore.Provider
	vkStore        applicantvkstore.Provider
}

func (i impl) getLogger() *log.Entry {
	logger := log.
		WithField("worker_name", "VkStep1Worker")
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
	//Получаем список анкет кандидатов для отпрвыки типовых вопросов
	list, err := i.applicantStore.ListOfActiveApplicants()
	if err != nil {
		logger.WithError(err).Error("ВК. Шаг 1. ошибка получения списка анкет кандидатов для генерации черновика скрипта")
		return
	}
	for _, applicant := range list {
		if applicant.Status != models.ApplicantStatusInProcess {
			continue
		}
		if applicant.ApplicantVkStep == nil {
			continue
		}
		if applicant.ApplicantVkStep.Status == dbmodels.VkStep0Done {
			// первичная генерация вопросов
			ok, err := vk.Instance.RunStep1(applicant)
			if err != nil {
				logger.WithError(err).
					WithField("space_id", applicant.SpaceID).
					WithField("applicant_id", applicant.ID).
					Error("ВК. Шаг 1. Ошибка генерации черновика скрипта")
				continue
			}
			if ok {
				logger.
					WithField("space_id", applicant.SpaceID).
					WithField("applicant_id", applicant.ID).
					Info("ВК. Шаг 1. Черновик скрипта сгенерирован")
			}
		} else if applicant.ApplicantVkStep.Status == dbmodels.VkStep1Regen {
			// перегенерация вопросов
			ok, err := vk.Instance.RunRegenStep1(applicant)
			if err != nil {
				logger.WithError(err).
					WithField("space_id", applicant.SpaceID).
					WithField("applicant_id", applicant.ID).
					Error("ВК. Шаг 1. Ошибка перегенерации черновика скрипта")
				continue
			}
			if ok {
				logger.
					WithField("space_id", applicant.SpaceID).
					WithField("applicant_id", applicant.ID).
					Info("ВК. Шаг 1. Черновик скрипта перегенерирован")
			}
		}
	}
}
