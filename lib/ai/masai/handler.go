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

type impl struct {
	ctx     context.Context
	baseUrl string
	session masaisessionstore.Provider
	busy    atomic.Bool
}

var Instance *impl

func NewHandler(ctx context.Context) {
	log.Infof("Инициализация ИИ: %v, модель: %v", config.Conf.AI.Masai.URL, "masai")
	instance := &impl{
		ctx:     ctx,
		baseUrl: config.Conf.AI.Masai.URL,
		session: masaisessionstore.NewInstance(db.DB),
	}
	initchecker.CheckInit(
		"session", instance.session,
	)
	Instance = instance
}

func GetHandler(ctx context.Context) *impl {
	log.Infof("Инициализация ИИ: %v, модель: %v", config.Conf.AI.Masai.URL, "masai")
	return &impl{
		ctx:     ctx,
		baseUrl: config.Conf.AI.Masai.URL,
		session: masaisessionstore.NewInstance(db.DB),
	}
}

func (i *impl) getLogger() *log.Entry {
	return log.
		WithField("ai", "masai")
}

func (i *impl) AnalyzeAnswer(vkStepID, applicantID, questionID string, reader io.Reader) (result surveyapimodels.VkAiInterviewResponse, err error) {
	//получаем данные по существующим сессиям
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
		// данных по текущему запросу нет, но есть другие, сначала завершим их
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
		WithField("answer_duration_sec", time.Now().Sub(now).Seconds()).
		Info("Ответ AI на запрос QueryMasai")

	return i.convertResponse(response), nil
}

func (i *impl) QueryMasai(reader io.Reader, fileName string, sessionRec dbmodels.MasaiSession) (result masaimodels.GradioResponse, err error) {
	// лочим ресурсы
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
		_, err = i.session.Save(sessionRec)
		if err != nil {
			logger.WithError(err).Error("ошибка сохранения идентификатора события в сессию")
		}
	}

	data, err := i.listenResults(sessionRec.EventID)
	if err != nil {
		i.removeSession(sessionRec.ID, false)
		return masaimodels.GradioResponse{}, errors.Wrap(err, "ошибка анализа видео файла")
	}
	i.removeSession(sessionRec.ID, true)

	var updates []masaimodels.GradioUpdate
	if err := json.Unmarshal(data, &updates); err != nil {
		return masaimodels.GradioResponse{}, errors.Wrap(err, "ошибка сериализации ответа")
	}
	return masaimodels.GradioResponse{Elements: updates}, nil
}

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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result []string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result[0], nil
}

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

	resp, err := http.Post(fmt.Sprintf("%v/call/event_handler_submit", i.baseUrl),
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

	return r["event_id"], nil
}

func (i *impl) listenResults(eventID string) (result []byte, err error) {
	// флаг занятости ИИ
	i.busy.Store(true)
	defer i.busy.Store(false)

	url := fmt.Sprintf("%v/call/event_handler_submit/%s", i.baseUrl, eventID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	var currentEvent string
	var currentData bytes.Buffer

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		line = bytes.TrimSpace(line)
		// Пустая строка означает конец события в SSE
		if len(line) == 0 {
			if currentEvent == "complete" {
				// Убираем лишние пробелы
				result := bytes.TrimSpace(currentData.Bytes())
				return result, nil
			}
			if currentEvent == "error" {
				return nil, errors.Wrap(errors.New(currentData.String()), "ошибка анализа видео файла")
			}
			// Сбрасываем для следующего события
			currentEvent = ""
			currentData.Reset()
			continue
		}

		if bytes.HasPrefix(line, []byte("event:")) {
			currentEvent = string(bytes.TrimSpace(line[len("event:"):]))
			currentData.Reset() // Сбрасываем данные при новом событии
		} else if bytes.HasPrefix(line, []byte("data:")) {
			// Каждая строка начинается с "data:"
			// Многострочные данные объединяются с одним переносом строки
			dataPart := line[len("data:"):]
			dataPart = bytes.TrimLeft(dataPart, " \t") // Убираем пробелы только в начале
			if currentData.Len() > 0 {
				currentData.WriteByte('\n') // Добавляем перенос между многострочными данными
			}
			currentData.Write(dataPart)
		}
	}
	return nil, errors.New("не получено событие complete")
}

func (i *impl) removeSession(id string, force bool) {
	if !force && helpers.IsContextDone(i.ctx) {
		// завершен контекст приложения, не удаляем сессию, тк возможно ИИ еще работает
		return
	}
	err := i.session.Delete(id)
	if err != nil {
		i.getLogger().WithError(err).Error("ошибка удаление сессии")
	}
}

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

func (i *impl) IsVideoAiAvailable() bool {
	return !i.busy.Load()
}
