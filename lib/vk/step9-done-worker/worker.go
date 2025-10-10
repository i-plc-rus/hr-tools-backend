package vkstep9doneworker

import (
	"context"
	"hr-tools-backend/db"
	"hr-tools-backend/lib/utils/helpers"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	vkvideoanalyzestore "hr-tools-backend/lib/vk/vk-video-analyze-store"
	dbmodels "hr-tools-backend/models/db"
	"runtime/debug"
	"time"

	log "github.com/sirupsen/logrus"
)

// Задача ВК. Шаг 9. семантическая оценка для видео опроса завершена (смена статуса)
func StartWorker(ctx context.Context) {
	i := &impl{
		vkVideoAnalyzeStore: vkvideoanalyzestore.NewInstance(db.DB),
		vkStore:             applicantvkstore.NewInstance(db.DB),
	}
	go i.runSemanticEvaluatedMonitor(ctx)
}

const (
	handlePeriod = 5 * time.Minute
)

type impl struct {
	vkVideoAnalyzeStore vkvideoanalyzestore.Provider
	vkStore             applicantvkstore.Provider
}

func (i impl) getLogger() *log.Entry {
	logger := log.
		WithField("worker_name", "VkStep9ScoreDoneWorker")
	return logger
}

func (i impl) runSemanticEvaluatedMonitor(ctx context.Context) {
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
			i.checkSemanticEvaluated(ctx)
			logger.Info("Задача выполнена")
		}
		period = handlePeriod
	}
}

func (i impl) checkSemanticEvaluated(ctx context.Context) {
	logger := i.getLogger()
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
