package vkstep10runworker

import (
	"context"
	"fmt"
	"hr-tools-backend/db"
	applicanthistoryhandler "hr-tools-backend/lib/applicant-history"
	applicantstore "hr-tools-backend/lib/applicant/store"
	baseworker "hr-tools-backend/lib/utils/base-worker"
	"hr-tools-backend/lib/utils/helpers"
	selectionstagestore "hr-tools-backend/lib/vacancy/selection-stage-store"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	vkvideoanalyzestore "hr-tools-backend/lib/vk/vk-video-analyze-store"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	threshold = 60 // Порог для прохождения интервью % 
)

// Задача ВК. Шаг 10. Подсчёт баллов и адаптивный фильтр
func StartWorker(ctx context.Context) {
	i := &impl{
		BaseImpl:            *baseworker.NewInstance("VkStep10RunWorker", 5*time.Second, 5*time.Minute),
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
	// Получаем ответы для оценки
	list, err := i.vkStore.GetByStatus(dbmodels.VkStepVideoSemanticEvaluated)
	if err != nil {
		logger.WithError(err).Error("ВК. Шаг 10. ошибка получения списка анкет")
		return
	}

	for _, rec := range list {
		if helpers.IsContextDone(ctx) {
			break
		}
		err := i.Scoring(rec)
		if err != nil {
			i.GetLogger().
				WithError(err).
				WithField("applicant_id", rec.ApplicantID).
				WithField("space_id", rec.SpaceID).
				Warn("ошибка оценки ответа")
		}
	}
}

func (i impl) Scoring(rec dbmodels.ApplicantVkStep) error {
	qCount := len(rec.Step1.Questions)
	weight := 1.0 / float64(qCount)
	totalScore := 0.0
	for _, answer := range rec.VideoInterviewEvaluations {
		weightI := float64(answer.Similarity) * weight
		totalScore += weightI
	}

	rec.TotalScore = int(totalScore)
	rec.Threshold = threshold
	rec.Pass = totalScore >= threshold
	rec.Status = dbmodels.VkStep10Filtered

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		err := updateSurvay(tx, rec)
		if err != nil {
			return errors.Wrap(err, "ошибка обновления баллов по опросу")
		}
		err = updateApplicant(tx, rec.SpaceID, rec.ApplicantID, rec.Pass)
		if err != nil {
			i.GetLogger().
				WithError(err).
				WithField("applicant_id", rec.ApplicantID).
				WithField("space_id", rec.SpaceID).
				Warn("Не удалось изменить статус отклика кандидата после оценки")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func updateSurvay(tx *gorm.DB, rec dbmodels.ApplicantVkStep) error {
	store := applicantvkstore.NewInstance(tx)
	_, err := store.Save(rec)
	if err != nil {
		return err
	}
	return nil
}

func updateApplicant(tx *gorm.DB, spaceID, applicantID string, isPass bool) error {
	store := applicantstore.NewInstance(tx)
	selectionStageStore := selectionstagestore.NewInstance(tx)
	applicantHistory := applicanthistoryhandler.NewTxHandler(tx)

	status := models.NegotiationStatusAccepted
	if !isPass {
		status = models.NegotiationStatusRejected
	}
	rec, err := store.GetByID(spaceID, applicantID)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("кандидат не найден")
	}

	msg, ok := rec.IsAllowStatusChange(status)
	if msg != "" {
		return errors.New(msg)
	}
	if !ok {
		// смена статуса не требуется
		return nil
	}
	changeMsg := fmt.Sprintf("Перевод отклика кандидата на статус %v", status)
	updMap := map[string]any{
		"negotiation_status":      status,
		"negotiation_accept_date": nil,
	}
	if status == models.NegotiationStatusAccepted {
		updMap["negotiation_accept_date"] = time.Now()
		updMap["status"] = models.ApplicantStatusInProcess
		selectionStages, err := selectionStageStore.List(rec.SpaceID, rec.VacancyID)
		if err != nil {
			return errors.Wrap(err, "ошибка получения списка этапов подбора")
		}
		for _, stage := range selectionStages {
			if stage.Name == dbmodels.AddedStage {
				updMap["selection_stage_id"] = stage.ID
				break
			}
		}
		changeMsg = "Кандидат из отклика, добавлен на вакансию"
	}
	if status == models.NegotiationStatusRejected {
		updMap["negotiation_accept_date"] = time.Now()
		changeMsg = "Отклик кандидата отклонен"
	}
	err = store.Update(applicantID, updMap)
	if err != nil {
		return errors.Wrap(err, "ошибка обновления кандидата")
	}
	changes := applicanthistoryhandler.GetUpdateChanges(changeMsg, rec.Applicant, updMap)
	applicantHistory.Save(rec.SpaceID, applicantID, rec.VacancyID, "", dbmodels.HistoryTypeUpdate, changes)
	return nil
}
