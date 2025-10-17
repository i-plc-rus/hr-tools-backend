package vkstep9runworker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"hr-tools-backend/db"
	masaihandler "hr-tools-backend/lib/ai/masai"
	masaisessionstore "hr-tools-backend/lib/ai/masai/session-store"
	filestorage "hr-tools-backend/lib/file-storage"
	baseworker "hr-tools-backend/lib/utils/base-worker"
	"hr-tools-backend/lib/utils/helpers"
	applicantvkstore "hr-tools-backend/lib/vk/applicant-vk-store"
	vkvideoanalyzestore "hr-tools-backend/lib/vk/vk-video-analyze-store"
	surveyapimodels "hr-tools-backend/models/api/survey"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

// Задача ВК. Шаг 9. Транскрибация
func StartWorker(ctx context.Context) {
	i := &impl{
		BaseImpl:              *baseworker.NewInstance("VkStep9Worker", 5*time.Second, 5*time.Minute),
		vkStore:               applicantvkstore.NewInstance(db.DB),
		vkAiInterviewProvider: masaihandler.GetHandler(ctx),
		vkVideoAnalyzeStore:   vkvideoanalyzestore.NewInstance(db.DB),
		session:               masaisessionstore.NewInstance(db.DB),
		fileStorage:           filestorage.Instance,
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
}

func (i impl) handle(ctx context.Context) {
	logger := i.GetLogger()
	// Получаем не завершенные запросы
	sessionRecs, err := i.session.GetAll()
	if err != nil {
		logger.WithError(err).Error("ВК. Шаг 9. ошибка получения списка незавершенных запросов")
	} else {
		for _, sessionRec := range sessionRecs {
			if helpers.IsContextDone(ctx) {
				break
			}
			vkStepRec, err := i.vkStore.GetByID(sessionRec.VkStepID)
			if err != nil || vkStepRec == nil {
				logger.Error("ВК. Шаг 9. ошибка получения данных по незавершенному запросу")
				i.session.Delete(sessionRec.ID)
				continue
			}
			answer, ok := vkStepRec.VideoInterview.Answers[sessionRec.QuestionID]
			if !ok {
				// данных по видео ответу нет (в принципе такого не должно быть)
				i.session.Delete(sessionRec.ID)
				continue
			}

			done, err := i.analyzeVideoAnswer(ctx, *vkStepRec, sessionRec.QuestionID, answer)
			if err != nil {
				i.GetLogger().
					WithError(err).
					WithField("applicant_id", vkStepRec.ApplicantID).
					WithField("file_id", answer.FileID).
					Warn("ошибка анализа видео файла")
			}
			if done {
				i.session.Delete(sessionRec.ID)
			}
		}
	}
	//Получаем список анкет кандидатов для отпрвыки типовых вопросов
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

func (i impl) analyzeVideoAnswers(ctx context.Context, vkStepRec dbmodels.ApplicantVkStep) (ok bool, err error) {
	handledCount := 0
	for questionID, answer := range vkStepRec.VideoInterview.Answers {
		done, err := i.analyzeVideoAnswer(ctx, vkStepRec, questionID, answer)
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
		if done {
			handledCount++
		}
		continue
	}

	if handledCount == len(vkStepRec.Step1.Questions) {
		// все видео обработаны
		vkStepRec.Status = dbmodels.VkStepVideoTranscripted
		_, err := i.vkStore.Save(vkStepRec)
		if err != nil {
			return false, errors.Wrap(err, "ошибка обновления статуса")
		}
		return true, nil
	}
	return false, nil
}

func (i impl) analyzeVideoAnswer(ctx context.Context, vkStepRec dbmodels.ApplicantVkStep, questionID string, answer dbmodels.VkVideoAnswer) (done bool, err error) {
	if answer.FileID == "" {
		return false, nil
	}
	videoAnalize, err := i.vkVideoAnalyzeStore.GetByStepQuestion(vkStepRec.ID, questionID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения данных о проанализированном ответе")
	}
	if videoAnalize != nil {
		return true, nil
	}

	reader, err := i.fileStorage.GetFileObject(ctx, vkStepRec.SpaceID, answer.FileID)
	if err != nil {
		i.saveFailAnalize(vkStepRec.ID, questionID, "ошибка загрузки видео файла из S3")
		return true, errors.Wrap(err, "ошибка загрузки видео файла из S3")
	}
	defer reader.Close()
	result, err := i.vkAiInterviewProvider.AnalyzeAnswer(vkStepRec.ID, vkStepRec.ApplicantID, questionID, reader)
	if err != nil {
		if helpers.IsContextDone(ctx) {
			return false, nil
		}
		i.saveFailAnalize(vkStepRec.ID, questionID, "ошибка анализа видео файла")
		return true, errors.Wrap(err, "ошибка анализа видео файла")
	}
	// анализ заверешен, сохраняем результат
	rec := dbmodels.ApplicantVkVideoSurvey{
		ApplicantVkStepID: vkStepRec.ID,
		QuestionID:        questionID,
		TranscriptText:    result.RecognizedText,
		Error:             "",
	}

	logger := i.GetLogger().
		WithField("applicant_id", vkStepRec.ApplicantID).
		WithField("question_id", questionID)

	// VoiceAmplitude
	fileID, err := i.saveImageFile(ctx, vkStepRec, result.VoiceAmplitude, getImageFileName(questionID, "voice"))
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения изображения с амплитудой голоса")
	} else {
		rec.VoiceAmplitudeFileID = fileID
	}

	// Frames
	fileID, err = i.saveImageFile(ctx, vkStepRec, result.Frames, getImageFileName(questionID, "frames"))
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения изображения с видео кадрами")
	} else {
		rec.FramesFileID = fileID
	}

	// Emotion
	fileID, err = i.saveImageFile(ctx, vkStepRec, result.Emotion, getImageFileName(questionID, "emotion"))
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения изображения с графиком эмоциий")
	} else {
		rec.EmotionFileID = fileID
	}

	// Sentiment
	fileID, err = i.saveImageFile(ctx, vkStepRec, result.Sentiment, getImageFileName(questionID, "sentiment"))
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения изображения с графиком настроения")
	} else {
		rec.SentimentFileID = fileID
	}

	_, err = i.vkVideoAnalyzeStore.Save(rec)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка сохранения результата видео анализа")
	}
	return true, nil
}

func (i impl) saveFailAnalize(vkStepsID, questionID string, errMsg string) {
	rec := dbmodels.ApplicantVkVideoSurvey{
		ApplicantVkStepID: vkStepsID,
		QuestionID:        questionID,
		TranscriptText:    "",
		Error:             errMsg,
	}
	_, err := i.vkVideoAnalyzeStore.Save(rec)
	if err != nil {
		i.GetLogger().
			WithError(err).
			Error("ошибка сохранения результата видео анализа")
	}
}

func (i impl) saveImageFile(ctx context.Context, vkStepRec dbmodels.ApplicantVkStep, fileData *surveyapimodels.VkResponseFileData, fileName string) (fileID string, err error) {
	if fileData == nil {
		return "", nil
	}

	// анализ заверешен, сохраняем результат
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

func getImageFileName(questionID, imageName string) string {
	return fmt.Sprintf("%v_%v.jpeg", questionID, imageName)
}
