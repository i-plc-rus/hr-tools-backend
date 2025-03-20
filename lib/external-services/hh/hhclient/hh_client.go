package hhclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	hhapimodels "hr-tools-backend/models/api/hh"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Provider interface {
	GetLoginUri(clientID, spaceID string) (string, error)
	RequestToken(ctx context.Context, req hhapimodels.RequestToken) (*hhapimodels.ResponseToken, error)
	RefreshToken(ctx context.Context, req hhapimodels.RefreshToken) (*hhapimodels.ResponseToken, error)
	Me(ctx context.Context, accessToken string) (*hhapimodels.MeResponse, error)

	//https://api.hh.ru/openapi/redoc#tag/Upravlenie-vakansiyami/operation/publish-vacancy
	VacancyPublish(ctx context.Context, accessToken string, request hhapimodels.VacancyPubRequest) (vacancyID string, err error)

	//https://api.hh.ru/openapi/redoc#tag/Upravlenie-vakansiyami/operation/edit-vacancy
	VacancyUpdate(ctx context.Context, accessToken string, vacancyID string, request hhapimodels.VacancyPubRequest) error

	//https://api.hh.ru/openapi/redoc#tag/Upravlenie-vakansiyami/operation/add-vacancy-to-hidden
	//https://api.hh.ru/openapi/redoc#tag/Upravlenie-vakansiyami/operation/add-vacancy-to-archive
	VacancyClose(ctx context.Context, accessToken string, employerID, vacancyID string) error

	//https://api.hh.ru/openapi/redoc#tag/Otklikipriglasheniya-rabotodatelya/operation/get-negotiations
	//https://api.hh.ru/openapi/redoc#tag/Otklikipriglasheniya-rabotodatelya/operation/get-collection-negotiations-list
	Negotiations(ctx context.Context, accessToken, vacancyID string, page, perPage int) (hhapimodels.NegotiationResponse, error)

	//https://api.hh.ru/openapi/redoc#tag/Otklikipriglasheniya-rabotodatelya/operation/post-negotiations-topics-read
	NegotiationMarkRead(ctx context.Context, accessToken, vacancyID, negotiationID string) error

	//https://api.hh.ru/openapi/redoc#tag/Prosmotr-rezyume/operation/get-resume
	GetResume(ctx context.Context, accessToken, resumeUrl string) (hhapimodels.ResumeResponse, error)

	GetAreas(ctx context.Context) ([]hhapimodels.Area, error)

	GetVacancy(ctx context.Context, accessToken, vacancyID string) (*hhapimodels.VacancyInfo, error)

	DownloadResume(ctx context.Context, accessToken, resumeUrl string) ([]byte, error)

	//https://api.hh.ru/openapi/redoc#tag/Otklikipriglasheniya-rabotodatelya/operation/get-negotiation-messages
	GetMessages(ctx context.Context, accessToken, vacancyID, negotiationID string) (hhapimodels.NegotiationMessagesResponse, error)

	// https://api.hh.ru/openapi/redoc#tag/Otklikipriglasheniya-rabotodatelya/operation/send-negotiation-message
	SendNewMessage(ctx context.Context, accessToken, vacancyID, negotiationID, message string) error
}

var Instance Provider

type impl struct {
	host        string
	redirectUri string
}

func NewProvider(redirectUri string) {
	Instance = &impl{
		host:        host,
		redirectUri: redirectUri,
	}
}

const (
	host                      string = "https://api.hh.ru"
	mePath                    string = "%s/me"
	tokenPath                 string = "%s/token"
	oAuthPattern              string = "https://hh.ru/oauth/authorize?response_type=code&client_id=%v&state=%v&redirect_uri=%v"
	vPublishPath              string = "%s/vacancies"
	vUpdatePath               string = "%s/vacancies/%v"
	vGetPath                  string = "%s/vacancies/%v"
	vDeletePath               string = "%s/employers/%v/vacancies/%v"
	vArchivePath              string = "%s/employers/%v/vacancies/archived/%v"
	negotiationCollectionPath string = "%s/negotiations?vacancy_id=%v"
	negotiationCollectionTpl  string = "%v&page=%v&per_page=%v"
	negotiationReadPath       string = "%s/negotiations/read"
	areasPath                 string = "%s/areas"
	messagesListPath          string = "%s/negotiations/%v/messages"
	messageNewPath            string = "%s/negotiations/%v/messages"
)

func (i impl) GetLoginUri(clientID, spaceID string) (string, error) {
	redirectUri, err := url.QueryUnescape(i.redirectUri)
	if err != nil {
		return "", errors.Wrap(err, "ошибка формирования ссылки")
	}
	uri := fmt.Sprintf(oAuthPattern, clientID, spaceID, redirectUri)
	return uri, nil
}

func (i impl) RequestToken(ctx context.Context, req hhapimodels.RequestToken) (*hhapimodels.ResponseToken, error) {
	redirectUri, err := url.QueryUnescape(i.redirectUri)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка формирования ссылки")
	}
	uri := fmt.Sprintf(tokenPath, i.host)
	data := url.Values{}
	data.Set("client_id", req.ClientID)
	data.Set("client_secret", req.ClientSecret)
	data.Set("code", req.Code)
	data.Add("redirect_uri", redirectUri)
	data.Set("grant_type", "authorization_code")

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp := hhapimodels.ResponseToken{}

	logger := log.
		WithField("external_request", uri).
		WithField("request_body", fmt.Sprintf("%+v", data.Encode()))

	err = i.sendRequest(logger, r, &resp, "", true)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (i impl) RefreshToken(ctx context.Context, req hhapimodels.RefreshToken) (*hhapimodels.ResponseToken, error) {
	uri := fmt.Sprintf(tokenPath, i.host)
	data := url.Values{}
	data.Add("refresh_token", req.RefreshToken)
	data.Set("grant_type", "refresh_token")

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp := hhapimodels.ResponseToken{}

	logger := log.
		WithField("external_request", uri).
		WithField("request_body", fmt.Sprintf("%+v", data.Encode()))

	err := i.sendRequest(logger, r, &resp, "", true)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (i impl) Me(ctx context.Context, accessToken string) (*hhapimodels.MeResponse, error) {
	uri := fmt.Sprintf(mePath, i.host)
	logger := log.
		WithField("external_request", uri)
	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.MeResponse{}
	err := i.sendRequest(logger, r, &resp, accessToken, true)
	return &resp, err
}

func (i impl) VacancyPublish(ctx context.Context, accessToken string, request hhapimodels.VacancyPubRequest) (vacancyID string, err error) {
	uri := fmt.Sprintf(vPublishPath, i.host)
	logger := log.
		WithField("external_request", uri)
	body, err := json.Marshal(request)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации запроса")
	}

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, bytes.NewBuffer(body))
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.VacancyResponse{}

	logger = logger.
		WithField("request_body", string(body))

	err = i.sendRequest(logger, r, &resp, accessToken, true)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (i impl) VacancyUpdate(ctx context.Context, accessToken, vacancyID string, request hhapimodels.VacancyPubRequest) error {
	uri := fmt.Sprintf(vUpdatePath, i.host, vacancyID)
	logger := log.
		WithField("vacancy_id", vacancyID).
		WithField("external_request", uri)
	body, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "ошибка десериализации запроса")
	}

	r, _ := http.NewRequestWithContext(ctx, "PUT", uri, bytes.NewBuffer(body))
	r.Header.Add("Content-Type", "application/json")

	logger = logger.
		WithField("request_body", string(body))

	return i.sendRequest(logger, r, nil, accessToken, true)
}

func (i impl) VacancyClose(ctx context.Context, accessToken, employerID, vacancyID string) error {
	uri := fmt.Sprintf(vArchivePath, i.host, employerID, vacancyID)
	logger := log.
		WithField("vacancy_id", vacancyID).
		WithField("employer_id", employerID).
		WithField("external_request", uri)
	r, _ := http.NewRequestWithContext(ctx, "PUT", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	return i.sendRequest(logger, r, nil, accessToken, true)
}

func (i impl) Negotiations(ctx context.Context, accessToken, vacancyID string, page, perPage int) (hhapimodels.NegotiationResponse, error) {
	collections, err := i.getNegotiations(ctx, accessToken, vacancyID)
	if err != nil {
		return hhapimodels.NegotiationResponse{}, errors.Wrap(err, "ошибка запроса списака откликов")
	}
	if len(collections.Collections) == 0 {
		return hhapimodels.NegotiationResponse{}, nil
	}
	return i.getNegotiationCollection(ctx, accessToken, collections.Collections[0].Url, page, perPage)
}

func (i impl) NegotiationMarkRead(ctx context.Context, accessToken, vacancyID, negotiationID string) error {
	uri := fmt.Sprintf(negotiationReadPath, i.host)
	logger := log.
		WithField("vacancy_id", vacancyID).
		WithField("negotiation_id", negotiationID).
		WithField("external_request", uri)

	request := hhapimodels.NegotiationReadRequest{
		TopicID: negotiationID,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "ошибка десериализации запроса")
	}

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, bytes.NewBuffer(body))
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.ResumeResponse{}

	err = i.sendRequest(logger, r, &resp, accessToken, true)
	if err != nil {
		return err
	}
	return nil
}

func (i impl) GetResume(ctx context.Context, accessToken, resumeUrl string) (hhapimodels.ResumeResponse, error) {
	resumeUrl = resumeUrl + "&with_job_search_status=true"
	logger := log.
		WithField("external_request", resumeUrl)

	r, _ := http.NewRequestWithContext(ctx, "GET", resumeUrl, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.ResumeResponse{}

	err := i.sendRequest(logger, r, &resp, accessToken, true)
	if err != nil {
		return hhapimodels.ResumeResponse{}, err
	}
	return resp, nil
}

func (i impl) GetVacancy(ctx context.Context, accessToken, vacancyID string) (*hhapimodels.VacancyInfo, error) {
	uri := fmt.Sprintf(vGetPath, i.host, vacancyID)
	logger := log.
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.VacancyInfo{}

	err := i.sendRequest(logger, r, &resp, accessToken, true)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (i impl) GetMessages(ctx context.Context, accessToken, vacancyID, negotiationID string) (hhapimodels.NegotiationMessagesResponse, error) {
	uri := fmt.Sprintf(messagesListPath, i.host, negotiationID)
	logger := log.
		WithField("vacancy_id", vacancyID).
		WithField("negotiation_id", negotiationID).
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.NegotiationMessagesResponse{}

	err := i.sendRequest(logger, r, &resp, accessToken, true)
	if err != nil {
		return hhapimodels.NegotiationMessagesResponse{}, err
	}
	return resp, nil
}

func (i impl) SendNewMessage(ctx context.Context, accessToken, vacancyID, negotiationID, message string) error {
	uri := fmt.Sprintf(messageNewPath, i.host, negotiationID)

	data := url.Values{}
	data.Set("message", message)

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	logger := log.
		WithField("vacancy_id", vacancyID).
		WithField("negotiation_id", negotiationID).
		WithField("external_request", uri).
		WithField("request_body", fmt.Sprintf("%+v", data.Encode()))

	return i.sendRequest(logger, r, nil, "", true)
}

func (i impl) getNegotiations(ctx context.Context, accessToken, vacancyID string) (*hhapimodels.NegotiationCollections, error) {
	uri := fmt.Sprintf(negotiationCollectionPath, i.host, vacancyID)
	logger := log.
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.NegotiationCollections{}

	err := i.sendRequest(logger, r, &resp, accessToken, true)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (i impl) getNegotiationCollection(ctx context.Context, accessToken, originUri string, page, perPage int) (hhapimodels.NegotiationResponse, error) {
	uri := fmt.Sprintf(negotiationCollectionTpl, originUri, page, perPage)
	logger := log.
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.NegotiationResponse{}

	err := i.sendRequest(logger, r, &resp, accessToken, true)
	if err != nil {
		return hhapimodels.NegotiationResponse{}, err
	}
	return resp, nil
}

func (i impl) GetAreas(ctx context.Context) ([]hhapimodels.Area, error) {
	uri := fmt.Sprintf(areasPath, i.host)
	logger := log.
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := []hhapimodels.Area{}

	err := i.sendRequest(logger, r, &resp, "", true)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i impl) DownloadResume(ctx context.Context, accessToken, resumeUrl string) ([]byte, error) {
	r, _ := http.NewRequestWithContext(ctx, "GET", resumeUrl, nil)
	r.Header.Add("Content-Type", "application/pdf")
	logger := log.
		WithField("external_request", resumeUrl)
	var fileBody []byte
	err := i.sendRequest(logger, r, &fileBody, accessToken, false)
	if err != nil {
		return nil, err
	}
	return fileBody, nil
}

func (i impl) sendRequest(logger *log.Entry, r *http.Request, resp interface{}, accessToken string, needUnmarshalResponse bool) error {
	r.Header.Add("User-Agent", "HRTools/1.0")
	if accessToken != "" {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	}
	client := &http.Client{}
	response, err := client.Do(r)
	addResponseBody(logger, response)
	addStatusCode(logger, response)
	if err != nil {
		logger.WithError(err).Error("ошибка отправки запроса в HH")
		return errors.Wrap(err, "ошибка отправки запроса в HH")
	}
	if response != nil && (response.StatusCode >= 200 && response.StatusCode <= 300) {
		if resp != nil {
			responseBody, _ := io.ReadAll(response.Body)
			if needUnmarshalResponse {
				err = json.Unmarshal(responseBody, resp)
				if err != nil {
					logger.WithError(err).Error("ошибка сериализации ответа")
					return errors.Wrap(err, "ошибка сериализации ответа")
				}
			}
		}
		return nil
	}

	errorResp := hhapimodels.ErrorData{}
	if response != nil {
		responseBody, _ := io.ReadAll(response.Body)
		err = json.Unmarshal(responseBody, &errorResp)
		if err != nil {
			logger.WithError(err).Error("ошибка сериализации ответа с ошибкой")
		}
	}
	logger.Error("Некорректный запрос в HH")
	return errors.Errorf("Некорректный запрос. Ошибка: %+v", errorResp)
}

func addResponseBody(logger *log.Entry, response *http.Response) *log.Entry {
	if response != nil {
		responseBody, _ := io.ReadAll(response.Body)
		return logger.WithField("response_body", string(responseBody))
	}
	return logger
}

func addStatusCode(logger *log.Entry, response *http.Response) *log.Entry {
	if response != nil {
		return logger.WithField("response_status_code", response.StatusCode)
	}
	return logger
}

func getErrorData(response *http.Response) (data hhapimodels.ErrorData, statusCode int, err error) {
	errorResp := hhapimodels.ErrorData{}
	if response != nil {
		responseBody, _ := io.ReadAll(response.Body)
		err = json.Unmarshal(responseBody, &errorResp)
		if err != nil {
			return hhapimodels.ErrorData{}, response.StatusCode, errors.Wrap(err, "ошибка сериализации ответа")
		}
		return errorResp, response.StatusCode, nil
	}
	return hhapimodels.ErrorData{}, 0, nil
}
