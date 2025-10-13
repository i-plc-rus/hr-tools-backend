package vkstep1runworker

import (
	"context"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	baseworker "hr-tools-backend/lib/utils/base-worker"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/lib/vk"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

// Задача ВК. Шаг 1. Генерация черновика скрипта
func StartWorker(ctx context.Context) {
	i := &impl{
		BaseImpl:       *baseworker.NewInstance("VkStep1Worker", 5*time.Second, 5*time.Minute),
		applicantStore: applicantstore.NewInstance(db.DB),
		vkStore:        applicantvkstore.NewInstance(db.DB),
	}
	go i.Run(ctx, i.handle)
}

type impl struct {
	baseworker.BaseImpl
	applicantStore applicantstore.Provider
	vkStore        applicantvkstore.Provider
}

func (i impl) handle(ctx context.Context) {
	logger := i.GetLogger()
	//Получаем список анкет кандидатов для отпрвыки типовых вопросов
	list, err := i.applicantStore.ListOfActiveApplicants()
	if err != nil {
		logger.WithError(err).Error("ВК. Шаг 1. ошибка получения списка анкет кандидатов для генерации черновика скрипта")
		return
	}
	for _, applicant := range list {
		if helpers.IsContextDone(ctx) {
			break
		}
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
