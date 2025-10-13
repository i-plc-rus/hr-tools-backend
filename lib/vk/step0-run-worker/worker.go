package vkstep0runworker

import (
	"context"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	baseworker "hr-tools-backend/lib/utils/base-worker"
	"hr-tools-backend/lib/vk"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

// Задача отправки ссылки на анкету с типовыми вопросами
// ВК. Шаг 0. Отправка типовых вопросов
func StartWorker(ctx context.Context) {
	i := &impl{
		BaseImpl:       *baseworker.NewInstance("VkStep0Worker", 5*time.Second, 5*time.Minute),
		applicantStore: applicantstore.NewInstance(db.DB),
		vkStore:        applicantvkstore.NewInstance(db.DB),
	}
	go i.Run(ctx, i.handle)
}

const (
	handlePeriod = 5 * time.Minute
)

type impl struct {
	baseworker.BaseImpl
	applicantStore applicantstore.Provider
	vkStore        applicantvkstore.Provider
}

func (i impl) handle(ctx context.Context) {
	logger := i.GetLogger()
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
