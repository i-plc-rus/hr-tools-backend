package vkstep9scoreworker

import (
	"context"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/lib/vk"
	vkvideoanalyzestore "hr-tools-backend/lib/vk/vk-video-analyze-store"
	"runtime/debug"
	"time"
)

// Задача ВК. Шаг 9. семантическая оценка ответов для видео опроса
func StartWorker(ctx context.Context) {
	i := &impl{
		vkVideoAnalyzeStore: vkvideoanalyzestore.NewInstance(db.DB),
	}
	go i.run(ctx)
}

const (
	handlePeriod = 5 * time.Minute
)

type impl struct {
	vkVideoAnalyzeStore vkvideoanalyzestore.Provider
}

func (i impl) getLogger() *log.Entry {
	logger := log.
		WithField("worker_name", "VkStep9ScoreWorker")
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
			i.handle(ctx)
			logger.Info("Задача выполнена")
		}
		period = handlePeriod
	}
}

func (i impl) handle(ctx context.Context) {
	logger := i.getLogger()
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
			i.getLogger().
				WithError(err).
				WithField("applicant_vk_step_id", rec.ApplicantVkStepID).
				WithField("question_id", rec.QuestionID).
				Warn("ошибка оценки ответа")
		}
	}
}
