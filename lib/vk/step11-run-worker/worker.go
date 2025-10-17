package vkstep11runworker

import (
	"context"
	"hr-tools-backend/db"
	baseworker "hr-tools-backend/lib/utils/base-worker"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/lib/vk"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	vkvideoanalyzestore "hr-tools-backend/lib/vk/vk-video-analyze-store"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

// Задача ВК. Шаг 11. Генерация отчёта и рекомендаций
func StartWorker(ctx context.Context) {
	i := &impl{
		BaseImpl:            *baseworker.NewInstance("VkStep11RunWorker", 21*time.Second, 5*time.Minute),
		vkVideoAnalyzeStore: vkvideoanalyzestore.NewInstance(db.DB),
		vkStore:             applicantvkstore.NewInstance(db.DB),
	}
	go i.Run(ctx, i.jobHandler)
}

type impl struct {
	baseworker.BaseImpl
	vkVideoAnalyzeStore vkvideoanalyzestore.Provider
	vkStore             applicantvkstore.Provider
}

func (i impl) jobHandler(ctx context.Context) {
	logger := i.GetLogger()
	// Получаем анкеты готовые для генерации отчета
	list, err := i.vkStore.GetByStatus(dbmodels.VkStep10Filtered)
	if err != nil {
		logger.WithError(err).Error("ВК. Шаг 11. Генерация отчёта и рекомендаций")
		return
	}

	for _, rec := range list {
		if helpers.IsContextDone(ctx) {
			break
		}
		ok, err := vk.Instance.GenerateReport(rec)
		if err != nil {
			i.GetLogger().
				WithError(err).
				WithField("applicant_id", rec.ApplicantID).
				WithField("space_id", rec.SpaceID).
				WithField("applicant_vk_step_id", rec.ID).
				Error("ВК. Шаг 11. Ошибка генерации отчёта")
			continue
		}
		if ok {
			logger.
				WithField("applicant_id", rec.ApplicantID).
				WithField("space_id", rec.SpaceID).
				WithField("applicant_vk_step_id", rec.ID).
				Info("ВК. Шаг 11. Генерация отчета завершена")
		}
	}
}
