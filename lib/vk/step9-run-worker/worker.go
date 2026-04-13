package vkstep9runworker

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"hr-tools-backend/db"
	masaihandler "hr-tools-backend/lib/ai/masai"
	masaisessionstore "hr-tools-backend/lib/ai/masai/session-store"
	filestorage "hr-tools-backend/lib/file-storage"
	ailogstore "hr-tools-backend/lib/gpt/store"
	baseworker "hr-tools-backend/lib/utils/base-worker"
	botnotify "hr-tools-backend/lib/utils/bot-notify"
	"hr-tools-backend/lib/utils/helpers"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	vkvideoanalyzestore "hr-tools-backend/lib/vk/vk-video-analyze-store"
	surveyapimodels "hr-tools-backend/models/api/survey"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
)

const (
	maxAutoRetries = 2               // всего 2 автоматические попытки (первая + один повтор)
	autoRetryDelay = 1 * time.Minute // задержка перед автоматическим повтором
)

// StartWorker запускает воркер для транскрибации видео
func StartWorker(ctx context.Context) {
	i := &impl{
		BaseImpl:              *baseworker.NewInstance("VkStep9Worker", 5*time.Second, 5*time.Minute),
		vkStore:               applicantvkstore.NewInstance(db.DB),
		vkAiInterviewProvider: masaihandler.Instance,
		vkVideoAnalyzeStore:   vkvideoanalyzestore.NewInstance(db.DB),
		session:               masaisessionstore.NewInstance(db.DB),
		fileStorage:           filestorage.Instance,
		logStore:              ailogstore.NewInstance(db.DB),
	}
	go i.Run(ctx, i.handle)
}

type impl struct {
	baseworker.BaseImpl
	vkStore               applicantvkstore.Provider
	vkAiInterviewProvider surveyapimodels.VkAiInterviewProvider
	vkVideoAnalyzeStore   vkvideoanalyzestore.Provider
	session               masaisessionstore.Provider
	fileStorage           filestorage.Provider
	logStore              ailogstore.Provider
}

// handle основной цикл воркера
func (i impl) handle(ctx context.Context) {
	logger := i.GetLogger()
	// Получаем не завершенные запросы
	sessionRecs, err := i.session.GetAll()
	if err != nil {
		logger.WithError(err).Error("ВК. Шаг 9. ошибка получения списка незавершенных запросов")
	} else {
		now := time.Now()
		for _, sessionRec := range sessionRecs {
			if helpers.IsContextDone(ctx) {
				break
			}
			// FIX: удаляем просроченные сессии
			if sessionRec.ExpiresAt != nil && sessionRec.ExpiresAt.Before(now) {
				_ = i.session.Delete(sessionRec.ID)
				continue
			}

			vkStepRec, err := i.vkStore.GetByID(sessionRec.VkStepID)
			if err != nil || vkStepRec == nil {
				logger.Error("ВК. Шаг 9. ошибка получения данных по незавершенному запросу")
				_ = i.session.Delete(sessionRec.ID)
				continue
			}
			answer, ok := vkStepRec.VideoInterview.Answers[sessionRec.QuestionID]
			if !ok {
				_ = i.session.Delete(sessionRec.ID)
				continue
			}

			done, err := i.analyzeVideoAnswer(ctx, *vkStepRec, sessionRec.QuestionID, answer, true)
			if err != nil {
				i.GetLogger().
					WithError(err).
					WithField("applicant_id", vkStepRec.ApplicantID).
					WithField("file_id", answer.FileID).
					Warn("ошибка анализа видео файла")
			}
			if done {
				_ = i.session.Delete(sessionRec.ID)
			}
		}
	}

	// Получение анкет для обработки
	list, err := i.vkStore.GetByStatus(dbmodels.VkStepVideoSuggestSent)
	if err != nil {
		logger.WithError(err).Error("ВК. Шаг 9. ошибка получения списка анкет кандидатов для анализа ответов")
		return
	}
	for _, vkStepRec := range list {
		if helpers.IsContextDone(ctx) {
			break
		}
		if len(vkStepRec.VideoInterview.Answers) == 0 {
			continue
		}
		ok, err := i.analyzeVideoAnswers(ctx, vkStepRec)
		if err != nil {
			logger.WithError(err).
				WithField("space_id", vkStepRec.SpaceID).
				WithField("applicant_id", vkStepRec.ApplicantID).
				Error("ВК. Шаг 9. Ошибка транскрибации видео ответов")
			continue
		}
		if ok {
			logger.
				WithField("space_id", vkStepRec.SpaceID).
				WithField("applicant_id", vkStepRec.ApplicantID).
				Info("ВК. Шаг 9. Транскрибация видео ответов завершена")
		}
	}
}

// analyzeVideoAnswers обрабатываем все ответы кандидата
func (i impl) analyzeVideoAnswers(ctx context.Context, vkStepRec dbmodels.ApplicantVkStep) (ok bool, err error) {
	for questionID, answer := range vkStepRec.VideoInterview.Answers {
		done, err := i.analyzeVideoAnswer(ctx, vkStepRec, questionID, answer, false)
		if !done && err != nil {
			return false, err
		}
		if err != nil {
			i.GetLogger().
				WithError(err).
				WithField("applicant_id", vkStepRec.ApplicantID).
				WithField("file_id", answer.FileID).
				Warn("ошибка анализа видео файла")
		}
	}

	answers, err := i.vkVideoAnalyzeStore.GetByApplicantVkStep(vkStepRec.ID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка подсчета обработанных видео")
	}
	handledCount := 0
	for _, answer := range answers {
		if answer.Error == "" || answer.ManualSkip {
			handledCount++
		}
	}

	if handledCount == len(vkStepRec.Step1.Questions) {
		vkStepRec.Status = dbmodels.VkStepVideoTranscripted
		_, err := i.vkStore.Save(vkStepRec)
		if err != nil {
			return false, errors.Wrap(err, "ошибка обновления статуса")
		}
		return true, nil
	}
	return false, nil
}

// analyzeVideoAnswer анализируем одно видео с учётом авто- и ручных повторов
func (i impl) analyzeVideoAnswer(ctx context.Context, vkStepRec dbmodels.ApplicantVkStep, questionID string, answer dbmodels.VkVideoAnswer, withSession bool) (done bool, err error) {
	if answer.FileID == "" {
		return false, nil
	}
	rec, err := i.vkVideoAnalyzeStore.GetByStepQuestion(vkStepRec.ID, questionID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения данных о проанализированном ответе")
	}

	// успешно или пропущено вручную
	if rec != nil && (rec.Error == "" || rec.ManualSkip) {
		return true, nil
	}

	// проверяем нужно ли выполнять попытку
	if rec != nil && rec.Error != "" {
		// ручной ретрай
		if rec.ManualRetry {
			i.GetLogger().WithField("analyze_id", rec.ID).Info("ручной ретрай, запускаем анализ")
		} else {
			// авто режим
			if rec.RetryCount >= maxAutoRetries {
				i.GetLogger().
					WithField("applicant_id", vkStepRec.ApplicantID).
					WithField("question_id", questionID).
					WithField("retry_count", rec.RetryCount).
					Info("превышено максимальное количество автоматических попыток, ожидаем ручного ретрая")
				return true, nil
			}
			if rec.LastAttemptAt != nil && time.Since(*rec.LastAttemptAt) < autoRetryDelay {
				return true, nil
			}
		}
	}

	// загрузка видео из S3 – фатальная ошибка, возвращаем true (сессия будет удалена)
	reader, err := i.fileStorage.GetFileObject(ctx, vkStepRec.SpaceID, answer.FileID)
	if err != nil {
		i.saveFailAnalize(rec, vkStepRec.ID, questionID, "ошибка загрузки видео файла из S3")
		return true, errors.Wrap(err, "ошибка загрузки видео файла из S3")
	}
	defer reader.Close()

	// Вызов AI (может быть временная ошибка)
	result, err := i.vkAiInterviewProvider.AnalyzeAnswer(vkStepRec.ID, vkStepRec.ApplicantID, questionID, reader)
	if err != nil {
		if helpers.IsContextDone(ctx) {
			i.saveFailAnalize(rec, vkStepRec.ID, questionID, "сервис прервал выполнение")
			return false, nil // контекст завершён – не удаляем сессию
		}
		i.saveFailAnalize(rec, vkStepRec.ID, questionID, "ошибка анализа видео файла")

		// сохраним ошибку в бд
		i.saveLog(vkStepRec, questionID, rec.RetryCount, err)

		// Отправка уведомлений в Telegram
		if rec != nil && rec.RetryCount >= maxAutoRetries {
			// все автоматические попытки закончились, предлагаем ручной ретрай
			retryLink := getRertyLink(rec.ID)
			skipLink := getSkipLink(rec.ID)
			botnotify.SendAiRetry("ошибка анализа видео файла, возможна повторная попытка", vkStepRec.SpaceID, vkStepRec.ApplicantID, err.Error(), retryLink, skipLink, i.GetLogger())
		} else {
			botnotify.SendAiResult("ошибка анализа видео файла, будет предпринята еще одна попытка", vkStepRec.SpaceID, vkStepRec.ApplicantID, err.Error(), i.GetLogger())
		}

		// Временная ошибка – возвращаем false, сессия остаётся для повтора
		return false, errors.Wrap(err, "ошибка анализа видео файла")
	}

	if rec == nil {
		rec = &dbmodels.ApplicantVkVideoSurvey{
			ApplicantVkStepID: vkStepRec.ID,
			QuestionID:        questionID,
			TranscriptText:    result.RecognizedText,
		}
	}
	rec.ManualRetry = false
	rec.Error = ""
	rec.RetryCount = 0
	rec.LastAttemptAt = nil

	logger := i.GetLogger().
		WithField("applicant_id", vkStepRec.ApplicantID).
		WithField("question_id", questionID)
	botnotify.SendAiResult("video analyze done", vkStepRec.SpaceID, vkStepRec.ApplicantID, "", logger)

	// сохраняем графики
	fileID, err := i.saveImageFile(ctx, vkStepRec, result.VoiceAmplitude, getImageFileName(questionID, "voice"))
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения изображения с амплитудой голоса")
	} else {
		rec.VoiceAmplitudeFileID = fileID
	}
	fileID, err = i.saveImageFile(ctx, vkStepRec, result.Frames, getImageFileName(questionID, "frames"))
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения изображения с видео кадрами")
	} else {
		rec.FramesFileID = fileID
	}
	fileID, err = i.saveImageFile(ctx, vkStepRec, result.Emotion, getImageFileName(questionID, "emotion"))
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения изображения с графиком эмоций")
	} else {
		rec.EmotionFileID = fileID
	}
	fileID, err = i.saveImageFile(ctx, vkStepRec, result.Sentiment, getImageFileName(questionID, "sentiment"))
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения изображения с графиком настроения")
	} else {
		rec.SentimentFileID = fileID
	}

	_, err = i.vkVideoAnalyzeStore.Save(*rec)
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения результата видео анализа")
	}
	return true, nil
}

// saveFailAnalize сохраняем неудачную попытку
func (i impl) saveFailAnalize(rec *dbmodels.ApplicantVkVideoSurvey, vkStepsID, questionID string, errMsg string) {
	now := time.Now()
	if rec == nil {
		rec = &dbmodels.ApplicantVkVideoSurvey{
			ApplicantVkStepID: vkStepsID,
			QuestionID:        questionID,
			TranscriptText:    "",
		}
	}
	rec.RetryCount++
	rec.LastAttemptAt = &now
	rec.Error = errMsg
	rec.ManualRetry = false // сбрасываем ручной ретрай после неудачи

	_, err := i.vkVideoAnalyzeStore.Save(*rec)
	if err != nil {
		i.GetLogger().WithError(err).Error("ошибка сохранения результата видео анализа")
	}
}

// saveImageFile сохраняем изображение в S3
func (i impl) saveImageFile(ctx context.Context, vkStepRec dbmodels.ApplicantVkStep, fileData *surveyapimodels.VkResponseFileData, fileName string) (string, error) {
	if fileData == nil {
		return "", nil
	}
	fileInfo := dbmodels.UploadFileInfo{
		SpaceID:        vkStepRec.SpaceID,
		ApplicantID:    vkStepRec.ApplicantID,
		FileName:       fileName,
		FileType:       dbmodels.ApplicantEmotions,
		ContentType:    fileData.ContentType,
		IsUniqueByName: true,
	}
	return i.fileStorage.UploadObject(ctx, fileInfo, bytes.NewReader(fileData.Body), len(fileData.Body))
}

func (i impl) saveLog(vkStepRec dbmodels.ApplicantVkStep, questionID string, retryCount int, executionErr error) {

	var sysPromtBuilder strings.Builder
	sysPromtBuilder.WriteString(fmt.Sprintf("vkStepRec.ID: %v\n", vkStepRec.ID))
	sysPromtBuilder.WriteString(fmt.Sprintf("questionID: %v\n", questionID))
	sysPromtBuilder.WriteString(fmt.Sprintf("applicantID: %v\n", vkStepRec.ApplicantID))
	sysPromtBuilder.WriteString(fmt.Sprintf("retryCount: %v\n", retryCount))

	rec := dbmodels.AiLog{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: vkStepRec.SpaceID,
		},
		SysPromt:   sysPromtBuilder.String(),
		UserPromt:  "",
		Answer:     fmt.Sprintf("%+v", executionErr),
		VacancyID:  "",
		ReqestType: dbmodels.AiVideoAnalyze,
		AiName:     dbmodels.AiMasaiType,
	}
	_, err := i.logStore.Save(rec)
	if err != nil {
		i.GetLogger().
			WithField("space_id", vkStepRec.SpaceID).
			WithError(err).
			Error("ошибка сохранения лога ИИ")
	}
}

func getImageFileName(questionID, imageName string) string {
	return fmt.Sprintf("%v_%v.jpeg", questionID, imageName)
}

func getRertyLink(videoSurveyRecID string) string {
	return fmt.Sprintf("/api/v1/space/applicant/analyze-retry/video/%v", videoSurveyRecID)
}

func getSkipLink(videoSurveyRecID string) string {
	return fmt.Sprintf("/api/v1/space/applicant/analyze-skip/video/%v", videoSurveyRecID)
}
