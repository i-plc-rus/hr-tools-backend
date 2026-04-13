package masaihandler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	masaisessionstore "hr-tools-backend/lib/ai/masai/session-store"
	"hr-tools-backend/lib/utils/helpers"
	initchecker "hr-tools-backend/lib/utils/init-checker"
	"hr-tools-backend/lib/utils/lock"
	masaimodels "hr-tools-backend/models/api/masai"
	surveyapimodels "hr-tools-backend/models/api/survey"
	dbmodels "hr-tools-backend/models/db"
	"io"
	"mime/multipart"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	shortRequestTimeout = 30 * time.Second
	uploadTimeout       = 10 * time.Minute // до 700 МБ - 10 минут
	listenTimeout       = 60 * time.Minute // таймаут на всю операцию ожидания – 60 минут
	sessionTTL          = 30 * time.Minute // tll сессии
)

type impl struct {
	ctx              context.Context
	baseUrl          string
	session          masaisessionstore.Provider
	busy             atomic.Bool
	shortHttpClient  *http.Client // для быстрых операций (submit)
	uploadHttpClient *http.Client // для отправки видео (upload)
	longHttpClient   *http.Client // для долгого listenResults (без таймаута, полагаемся на контекст)

}

var Instance *impl

func NewHandler(ctx context.Context) {
	log.Infof("Инициализация ИИ: %v, модель: %v", config.Conf.AI.Masai.URL, "masai")
	instance := &impl{
		ctx:     ctx,
		baseUrl: config.Conf.AI.Masai.URL,
		session: masaisessionstore.NewInstance(db.DB),
		shortHttpClient: &http.Client{
			Timeout: shortRequestTimeout,
		},
		uploadHttpClient: &http.Client{
			Timeout: uploadTimeout,
		},
		longHttpClient: &http.Client{
			Timeout: 0, // тут таймаутом управляем через контекст
		},
	}
	initchecker.CheckInit("session", instance.session)
	Instance = instance
}

func GetHandler(ctx context.Context) *impl {
	log.Infof("Инициализация ИИ: %v, модель: %v", config.Conf.AI.Masai.URL, "masai")
	return &impl{
		ctx:     ctx,
		baseUrl: config.Conf.AI.Masai.URL,
		session: masaisessionstore.NewInstance(db.DB),
		shortHttpClient: &http.Client{
			Timeout: shortRequestTimeout,
		},
		uploadHttpClient: &http.Client{
			Timeout: uploadTimeout,
		},
		longHttpClient: &http.Client{
			Timeout: 0,
		},
	}
}

func (i *impl) getLogger() *log.Entry {
	return log.WithField("ai", "masai")
}

// AnalyzeAnswer основной метод анализа видео
func (i *impl) AnalyzeAnswer(vkStepID, applicantID, questionID string, reader io.Reader) (result surveyapimodels.VkAiInterviewResponse, err error) {
	sessionRecs, err := i.session.GetAll()
	if err != nil {
		return surveyapimodels.VkAiInterviewResponse{}, err
	}

	var sessionRec dbmodels.MasaiSession
	if len(sessionRecs) == 0 {
		sessionRec = dbmodels.MasaiSession{
			VkStepID:    vkStepID,
			QuestionID:  questionID,
			ApplicantID: applicantID,
			VideoPath:   "",
			EventID:     "",
			ExpiresAt:   timePtr(time.Now().Add(30 * time.Minute)),
		}
		id, err := i.session.Save(sessionRec)
		if err != nil {
			return surveyapimodels.VkAiInterviewResponse{}, err
		}
		sessionRec.ID = id
	} else {
		for _, rec := range sessionRecs {
			if rec.VkStepID == vkStepID && rec.QuestionID == questionID {
				sessionRec = rec
				break
			}
		}
		if sessionRec.ID == "" {
			return surveyapimodels.VkAiInterviewResponse{}, errors.New("найден незавершенный запрос")
		}
	}

	now := time.Now()
	response, err := i.QueryMasai(reader, fmt.Sprintf("%v.mp4", questionID), sessionRec)
	if err != nil {
		return surveyapimodels.VkAiInterviewResponse{}, err
	}
	i.getLogger().
		WithField("applicant_id", applicantID).
		WithField("question_id", questionID).
		WithField("answer", response).
		WithField("answer_duration_sec", time.Since(now).Seconds()).
		Info("Ответ AI на запрос QueryMasai")

	return i.convertResponse(response), nil
}

// QueryMasai выполняет полный цикл: загрузка, запуск, ожидание результатов
func (i *impl) QueryMasai(reader io.Reader, fileName string, sessionRec dbmodels.MasaiSession) (result masaimodels.GradioResponse, err error) {
	if !lock.Resource.Acquire(i.ctx, "QueryMasai") {
		return masaimodels.GradioResponse{}, errors.New("ошибка доступа к ресурсам - контекст завершен")
	}
	defer lock.Resource.Release("QueryMasai")

	logger := i.getLogger()
	if sessionRec.VideoPath == "" {
		videoPath, err := i.uploadVideo(reader, fileName)
		if err != nil {
			i.removeSession(sessionRec.ID, false)
			return masaimodels.GradioResponse{}, errors.Wrap(err, "ошибка отправки видео файла на анализ")
		}
		logger.Info("Видео файл загружен, путь к файлу (VideoPath):", videoPath)
		sessionRec.VideoPath = videoPath
		sessionRec.ExpiresAt = timePtr(time.Now().Add(30 * time.Minute))
		_, err = i.session.Save(sessionRec)
		if err != nil {
			logger.WithError(err).Error("ошибка сохранения пути видео файла в сессию")
		}
	}

	if sessionRec.EventID == "" {
		eventID, err := i.submitJob(sessionRec.VideoPath)
		if err != nil {
			i.removeSession(sessionRec.ID, false)
			return masaimodels.GradioResponse{}, errors.Wrap(err, "ошибка запуска анализа видео файла")
		}
		logger.Info("Задание отправлено, идентификатор события (EventID):", eventID)
		sessionRec.EventID = eventID
		sessionRec.ExpiresAt = timePtr(time.Now().Add(30 * time.Minute))
		_, err = i.session.Save(sessionRec)
		if err != nil {
			logger.WithError(err).Error("ошибка сохранения идентификатора события в сессию")
		}
	}

	data, err := i.listenResults(i.ctx, sessionRec.EventID)
	if err != nil {
		// если ошибка связана с обрывом соединения, не удаляем сессию – дадим шанс повторить
		if errors.Is(err, io.ErrUnexpectedEOF) {
			i.getLogger().WithError(err).Warn("неожиданный обрыв соединения при ожидании результатов, сессия сохранена")
			return masaimodels.GradioResponse{}, errors.Wrap(err, "временная ошибка сети, повторите позже")
		}
		i.removeSession(sessionRec.ID, false)
		return masaimodels.GradioResponse{}, errors.Wrap(err, "ошибка анализа видео файла")
	}

	var updates []masaimodels.GradioUpdate
	if err := json.Unmarshal(data, &updates); err != nil {
		return masaimodels.GradioResponse{}, errors.Wrap(err, "ошибка сериализации ответа")
	}
	return masaimodels.GradioResponse{Elements: updates}, nil
}

// uploadVideo загружает видео на сервер AI (быстрая операция)
func (i *impl) uploadVideo(reader io.Reader, fileName string) (videoPath string, err error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("files", fileName)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, reader)
	if err != nil {
		return "", err
	}
	writer.Close()

	req, err := http.NewRequestWithContext(i.ctx, "POST", fmt.Sprintf("%v/upload", i.baseUrl), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := i.uploadHttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result []string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", errors.New("пустой ответ от сервера при загрузке видео")
	}
	return result[0], nil
}

// submitJob запускает задачу анализа (быстрая операция)
func (i *impl) submitJob(videoPath string) (string, error) {
	payload := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"video": map[string]interface{}{
					"path": videoPath,
					"meta": map[string]string{
						"_type": "gradio.FileData",
					},
				},
				"subtitles": nil,
			},
		},
	}

	data, _ := json.Marshal(payload)

	resp, err := i.shortHttpClient.Post(fmt.Sprintf("%v/call/event_handler_submit", i.baseUrl),
		"application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var r map[string]string
	if err := json.Unmarshal(body, &r); err != nil {
		return "", err
	}
	eventID, ok := r["event_id"]
	if !ok || eventID == "" {
		return "", errors.New("не получен event_id в ответе")
	}
	return eventID, nil
}

// listenResults ожидает завершения обработки через SSE (долгая операция, до 60 минут)
func (i *impl) listenResults(ctx context.Context, eventID string) (result []byte, err error) {
	i.busy.Store(true)
	defer i.busy.Store(false)

	// Таймаут на всю операцию ожидания – 60 минут
	ctx, cancel := context.WithTimeout(ctx, listenTimeout)
	defer cancel()

	url := fmt.Sprintf("%v/call/event_handler_submit/%s", i.baseUrl, eventID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Используем longHttpClient без встроенного таймаута
	resp, err := i.longHttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	var currentEvent string
	var currentData bytes.Buffer

	// Канал для сигнала о завершении по контексту
	done := make(chan struct{})
	defer close(done)

	// Горутина для принудительного закрытия соединения при отмене контекста
	go func() {
		select {
		case <-ctx.Done():
			resp.Body.Close()
		case <-done:
		}
	}()

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, err
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			if currentEvent == "complete" {
				result := bytes.TrimSpace(currentData.Bytes())
				return result, nil
			}
			if currentEvent == "error" {
				errMsg := currentData.String()
				i.getLogger().WithField("event_id", eventID).Error("ошибка от AI: ", errMsg)
				return nil, errors.Errorf("ошибка анализа видео файла: %s", errMsg)
			}
			currentEvent = ""
			currentData.Reset()
			continue
		}

		if bytes.HasPrefix(line, []byte("event:")) {
			currentEvent = string(bytes.TrimSpace(line[len("event:"):]))
			currentData.Reset()
		} else if bytes.HasPrefix(line, []byte("data:")) {
			dataPart := line[len("data"):]
			dataPart = bytes.TrimLeft(dataPart, " \t")
			if currentData.Len() > 0 {
				currentData.WriteByte('\n')
			}
			currentData.Write(dataPart)
		}
	}
	return nil, errors.New("не получено событие complete")
}

// removeSession удаляет сессию, если force == true или контекст не завершён
func (i *impl) removeSession(id string, force bool) {
	if !force && helpers.IsContextDone(i.ctx) {
		return
	}
	err := i.session.Delete(id)
	if err != nil {
		i.getLogger().WithError(err).Error("ошибка удаления сессии")
	}
}

// convertResponse преобразует ответ от Masai в нужный формат
func (i *impl) convertResponse(response masaimodels.GradioResponse) (result surveyapimodels.VkAiInterviewResponse) {
	result.RecognizedText = response.GetRecognizedText()
	for k, elem := range response.Elements {
		if elem.IsPlotValue() {
			p, _ := elem.ToPlotValue()
			contentType, body, err := p.ToByteArr()
			if err == nil {
				switch k {
				case 1:
					result.VoiceAmplitude = &surveyapimodels.VkResponseFileData{
						Body:        body,
						ContentType: contentType,
					}
				case 2:
					result.Frames = &surveyapimodels.VkResponseFileData{
						Body:        body,
						ContentType: contentType,
					}
				case 3:
					result.Emotion = &surveyapimodels.VkResponseFileData{
						Body:        body,
						ContentType: contentType,
					}
				case 4:
					result.Sentiment = &surveyapimodels.VkResponseFileData{
						Body:        body,
						ContentType: contentType,
					}
				}
			}
		}
	}
	return result
}

// IsVideoAiAvailable возвращает true, если AI не занят
func (i *impl) IsVideoAiAvailable() bool {
	return !i.busy.Load()
}

func timePtr(t time.Time) *time.Time {
	return &t
}
