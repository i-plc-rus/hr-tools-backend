package vkstep0runworker

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

// Задача отправки ссылки на анкету с типовыми вопросами
// ВК. Шаг 0. Отправка типовых вопросов
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
		WithField("worker_name", "VkStep0Worker")
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
	list, err := i.applicantStore.ListOfActivefNegotiation(false)
	if err != nil {
		logger.WithError(err).Error("ВК. Шаг 0. ошибка получения списка анкет кандидатов для оценки")
		return
	}
	for _, applicant := range list {
		if applicant.NegotiationStatus == models.NegotiationStatusRejected ||
			applicant.NegotiationStatus == models.NegotiationStatusAccepted {
			continue
		}
		if applicant.ApplicantVkStep != nil && applicant.ApplicantVkStep.Status != dbmodels.VkStep0NotSent {
			continue
		}

		ok, err := vk.Instance.RunStep0(applicant)
		if err != nil {
			logger.WithError(err).
				WithField("space_id", applicant.SpaceID).
				WithField("applicant_id", applicant.ID).
				Error("ВК. Шаг 0. Ошибка отправки анкеты кандидату")
			continue
		}
		if ok {
			logger.
				WithField("space_id", applicant.SpaceID).
				WithField("applicant_id", applicant.ID).
				Info("ВК. Шаг 0. Кандидату отправлена ссылка на анкету с типовыми вопросами")
		}
	}
}
