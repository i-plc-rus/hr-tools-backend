package vkstep9scoreworker

import (
	"context"
	"hr-tools-backend/db"
	baseworker "hr-tools-backend/lib/utils/base-worker"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/lib/vk"
	vkvideoanalyzestore "hr-tools-backend/lib/vk/vk-video-analyze-store"
	"time"
)

// Задача ВК. Шаг 9. семантическая оценка ответов для видео опроса
func StartWorker(ctx context.Context) {
	i := &impl{
		BaseImpl:            *baseworker.NewInstance("VkStep9ScoreWorker", 5*time.Second, 5*time.Minute),
		vkVideoAnalyzeStore: vkvideoanalyzestore.NewInstance(db.DB),
	}
	go i.Run(ctx, i.handle)
}

type impl struct {
	baseworker.BaseImpl
	vkVideoAnalyzeStore vkvideoanalyzestore.Provider
}

func (i impl) handle(ctx context.Context) {
	logger := i.GetLogger()
	// Получаем ответы для оценки
	list, err := i.vkVideoAnalyzeStore.GetForScore()
	if err != nil {
		logger.WithError(err).Error("ВК. Шаг 9. ошибка получения списка ответов требующих оценки")
		return
	}

	for _, rec := range list {
		if rec.Error != "" {
			continue
		}
		if helpers.IsContextDone(ctx) {
			break
		}
		err := vk.Instance.ScoreAnswer(rec)
		if err != nil {
			i.GetLogger().
				WithError(err).
				WithField("applicant_vk_step_id", rec.ApplicantVkStepID).
				WithField("question_id", rec.QuestionID).
				Warn("ошибка оценки ответа")
		}
	}
}
