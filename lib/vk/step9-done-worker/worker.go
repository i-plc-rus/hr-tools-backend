package vkstep9doneworker

import (
	"context"
	"hr-tools-backend/db"
	baseworker "hr-tools-backend/lib/utils/base-worker"
	"hr-tools-backend/lib/utils/helpers"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	vkvideoanalyzestore "hr-tools-backend/lib/vk/vk-video-analyze-store"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

// Задача ВК. Шаг 9. семантическая оценка для видео опроса завершена (смена статуса)
func StartWorker(ctx context.Context) {
	i := &impl{
		BaseImpl:            *baseworker.NewInstance("VkStep9ScoreDoneWorker", 5*time.Second, 5*time.Minute),
		vkVideoAnalyzeStore: vkvideoanalyzestore.NewInstance(db.DB),
		vkStore:             applicantvkstore.NewInstance(db.DB),
	}
	go i.Run(ctx, i.checkSemanticEvaluated)
}

type impl struct {
	baseworker.BaseImpl
	vkVideoAnalyzeStore vkvideoanalyzestore.Provider
	vkStore             applicantvkstore.Provider
}

func (i impl) checkSemanticEvaluated(ctx context.Context) {
	logger := i.GetLogger()
	list, err := i.vkStore.GetByStatus(dbmodels.VkStepVideoTranscripted)
	if err != nil {
		logger.WithError(err).Error("ошибка получения списка анкет")
		return
	}
	for _, rec := range list {
		if helpers.IsContextDone(ctx) {
			break
		}
		scoredRowsCount, err := i.vkVideoAnalyzeStore.GetScoredCount(rec.ID)
		if err != nil {
			logger.WithError(err).Error("ошибка получения количества оцененных ответов")
			continue
		}
		if len(rec.VideoInterview.Answers) <= int(scoredRowsCount) {
			rec.Status = dbmodels.VkStepVideoSemanticEvaluated
			_, err = i.vkStore.Save(rec)
			if err != nil {
				logger.WithError(err).Error("ошибка обновления статуса анкеты")
			}
		}
	}
}
