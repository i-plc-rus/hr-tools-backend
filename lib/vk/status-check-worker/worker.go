package vkstatuscheckworker

import (
	"context"
	"hr-tools-backend/db"
	baseworker "hr-tools-backend/lib/utils/base-worker"
	"hr-tools-backend/lib/utils/helpers"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

// Задача ВК. Проверка и обновление статуса
func StartWorker(ctx context.Context) {
	i := &impl{
		BaseImpl:         *baseworker.NewInstance("VkStatusCheckWorker", 25*time.Second, time.Hour),
		applicantVkStore: applicantvkstore.NewInstance(db.DB),
	}
	go i.Run(ctx, i.handle)
}

type impl struct {
	baseworker.BaseImpl
	applicantVkStore applicantvkstore.Provider
}

var checkStatusSlice = []models.VideoInterviewStatus{
	models.VideoInterviewStatusUploading,
	models.VideoInterviewStatusProcessing,
	"",
}

func (i impl) handle(ctx context.Context) {
	logger := i.GetLogger()
	list, err := i.applicantVkStore.GetByVideoInterviewStatus(checkStatusSlice)
	if err != nil {
		logger.WithError(err).Error("ВК. ошибка получения списка видео интервью для проверки статуса")
		return
	}

	now := time.Now()
	for _, rec := range list {
		if helpers.IsContextDone(ctx) {
			break
		}

		if rec.VideoInterview.Status == "" {
			// совместимость для старых записей
			i.handleNoStatus(rec)
			continue
		}

		shouldUpdate := false
		switch rec.VideoInterview.Status {
		case models.VideoInterviewStatusUploading:
			if rec.VideoInterview.StartTime == nil {
				rec.VideoInterview.StartTime = &now
				shouldUpdate = true
				break
			}
			// Если загрузка не завершилась за TTL (2 часа) - ERROR
			if rec.VideoInterview.StartTime.Add(time.Hour * 2).Before(time.Now()) {
				rec.VideoInterview.Status = models.VideoInterviewStatusError
				shouldUpdate = true
				break
			}
			break
		case models.VideoInterviewStatusProcessing:
			//  любая ошибка загрузки/обработки
			for _, evaluation := range rec.VideoInterviewEvaluations {
				if evaluation.Error != "" && !evaluation.ManualSkip {
					rec.VideoInterview.Status = models.VideoInterviewStatusError
					shouldUpdate = true
					break
				}
			}
		}
		if shouldUpdate {
			_, err = i.applicantVkStore.Save(rec)
			if err != nil {
				i.GetLogger().
					WithError(err).
					WithField("rec_id", rec.ID).
					Error("ВК. ошибка обновления статуса видео интервью")
			}
		}

	}
}

func (i impl) handleNoStatus(rec dbmodels.ApplicantVkStep) {
	// совместимость для старых записей
	loadedAnswers := len(rec.VideoInterview.Answers)
	switch rec.Status {
	case dbmodels.VkStep11Report:
		rec.VideoInterview.Status = models.VideoInterviewStatusReady
		break

	case dbmodels.VkStep10Filtered:
	case dbmodels.VkStepVideoSemanticEvaluated:
	case dbmodels.VkStepVideoTranscripted:
		rec.VideoInterview.Status = models.VideoInterviewStatusProcessing
		break

	case dbmodels.VkStepVideoSuggestSent:
		if loadedAnswers > 0 {
			rec.VideoInterview.Status = models.VideoInterviewStatusUploading
			now := time.Now()
			rec.VideoInterview.StartTime = &now
		}
		break
	default:
		rec.VideoInterview.Status = models.VideoInterviewStatusAbsent
	}
	_, err := i.applicantVkStore.Save(rec)
	if err != nil {
		i.GetLogger().
			WithError(err).
			WithField("rec_id", rec.ID).
			Error("ВК. ошибка обновления статуса видео интервью")
	}
}
