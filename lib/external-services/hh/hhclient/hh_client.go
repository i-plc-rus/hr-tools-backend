package hhclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hr-tools-backend/db"
	externalservices "hr-tools-backend/lib/external-services"
	extapiauditstore "hr-tools-backend/lib/external-services/ext-api-audit-store"
	initchecker "hr-tools-backend/lib/utils/init-checker"
	hhapimodels "hr-tools-backend/models/api/hh"
	dbmodels "hr-tools-backend/models/db"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	GetLoginUri(clientID, spaceID string) (string, error)
	RequestToken(ctx context.Context, req hhapimodels.RequestToken) (*hhapimodels.ResponseToken, error)
	RefreshToken(ctx context.Context, req hhapimodels.RefreshToken) (token *hhapimodels.ResponseToken, isDeactivated bool, err error)
	Me(ctx context.Context, accessToken string) (me *hhapimodels.MeResponse, isExpired bool, err error)

	//https://api.hh.ru/openapi/redoc#tag/Upravlenie-vakansiyami/operation/publish-vacancy
	VacancyPublish(ctx context.Context, accessToken string, request hhapimodels.VacancyPubRequest) (vacancyID string, hMsg string, err error)

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
	auditStore  extapiauditstore.Provider
}

func NewProvider(redirectUri string) {
	instance := &impl{
		host:        host,
		redirectUri: redirectUri,
		auditStore:  extapiauditstore.NewInstance(db.DB),
	}
	initchecker.CheckInit(
		"auditStore", instance.auditStore,
	)
	Instance = instance
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
const (
	tokenExpiredError     string = "token-expired"
	tokenDeactivatedError string = "token deactivated"
	serviceName           string = "HH"
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

	err = i.sendRequest(ctx, logger, r, &resp, "", true)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (i impl) RefreshToken(ctx context.Context, req hhapimodels.RefreshToken) (token *hhapimodels.ResponseToken, isDeactivated bool, err error) {
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

	errData, err := i.sendRequestWithErrorData(ctx, logger, r, &resp, "", true)
	if err != nil {
		if errData != nil && errData.ErrorDescription == tokenDeactivatedError {
			return nil, true, nil
		}
		return nil, false, err
	}
	return &resp, false, nil
}

func (i impl) Me(ctx context.Context, accessToken string) (me *hhapimodels.MeResponse, isExpired bool, err error) {
	uri := fmt.Sprintf(mePath, i.host)
	logger := log.
		WithField("external_request", uri)
	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.MeResponse{}
	errData, err := i.sendRequestWithErrorData(ctx, logger, r, &resp, accessToken, true)
	if err != nil {
		if errData != nil && errData.OauthError == tokenExpiredError {
			return nil, true, nil
		}
		return nil, false, err
	}
	return &resp, false, nil
}

func (i impl) VacancyPublish(ctx context.Context, accessToken string, request hhapimodels.VacancyPubRequest) (vacancyID string, hMsg string, err error) {
	uri := fmt.Sprintf(vPublishPath, i.host)
	logger := log.
		WithField("external_request", uri)
	body, err := json.Marshal(request)
	if err != nil {
		return "", "", errors.Wrap(err, "ошибка десериализации запроса")
	}

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, bytes.NewBuffer(body))
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.VacancyResponse{}

	logger = logger.
		WithField("request_body", string(body))

	rCtx := externalservices.GetAuditContext(ctx, uri, body)
	errData, err := i.sendRequestWithErrorData(rCtx, logger, r, &resp, accessToken, true)
	if err != nil {
		if errData != nil {
			return "", errData.GetPublishErrorReason(), nil
		}
		return "", "", err
	}
	return resp.ID, "", nil
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

	rCtx := externalservices.GetAuditContext(ctx, uri, body)
	return i.sendRequest(rCtx, logger, r, nil, accessToken, true)
}

func (i impl) VacancyClose(ctx context.Context, accessToken, employerID, vacancyID string) error {
	uri := fmt.Sprintf(vArchivePath, i.host, employerID, vacancyID)
	logger := log.
		WithField("vacancy_id", vacancyID).
		WithField("employer_id", employerID).
		WithField("external_request", uri)
	r, _ := http.NewRequestWithContext(ctx, "PUT", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	rCtx := externalservices.GetAuditContext(ctx, uri, nil)
	return i.sendRequest(rCtx, logger, r, nil, accessToken, true)
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

	err = i.sendRequest(ctx, logger, r, &resp, accessToken, true)
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

	err := i.sendRequest(ctx, logger, r, &resp, accessToken, true)
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

	err := i.sendRequest(ctx, logger, r, &resp, accessToken, true)
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

	err := i.sendRequest(ctx, logger, r, &resp, accessToken, true)
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

	return i.sendRequest(ctx, logger, r, nil, "", true)
}

func (i impl) getNegotiations(ctx context.Context, accessToken, vacancyID string) (*hhapimodels.NegotiationCollections, error) {
	uri := fmt.Sprintf(negotiationCollectionPath, i.host, vacancyID)
	logger := log.
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.NegotiationCollections{}

	err := i.sendRequest(ctx, logger, r, &resp, accessToken, true)
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

	err := i.sendRequest(ctx, logger, r, &resp, accessToken, true)
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

	err := i.sendRequest(ctx, logger, r, &resp, "", true)
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
	err := i.sendRequest(ctx, logger, r, &fileBody, accessToken, false)
	if err != nil {
		return nil, err
	}
	return fileBody, nil
}

func (i impl) sendRequest(ctx context.Context, logger *log.Entry, r *http.Request, resp interface{}, accessToken string, needUnmarshalResponse bool) error {
	_, err := i.sendRequestWithErrorData(ctx, logger, r, resp, accessToken, needUnmarshalResponse)
	return err
}

func (i impl) sendRequestWithErrorData(ctx context.Context, logger *log.Entry, r *http.Request, resp interface{}, accessToken string, needUnmarshalResponse bool) (errData *hhapimodels.ErrorData, err error) {
	r.Header.Add("User-Agent", "HRTools/1.0")
	if accessToken != "" {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	}
	client := &http.Client{}
	response, err := client.Do(r)
	// читаем Body только 1 раз
	responseBody, logger := getResponseBody(logger, response)
	logger = addStatusCode(logger, response)
	if err != nil {
		logger.WithError(err).Error("ошибка отправки запроса в HH")
		return nil, errors.Wrap(err, "ошибка отправки запроса в HH")
	}
	if response != nil && (response.StatusCode >= 200 && response.StatusCode <= 300) {
		if resp != nil && needUnmarshalResponse {
			if responseBody != nil {
				err = json.Unmarshal(responseBody, resp)
				if err != nil {
					logger.WithError(err).Error("ошибка сериализации ответа")
					return nil, errors.Wrap(err, "ошибка сериализации ответа")
				}
			}
		}
		return nil, nil
	}
	logger.Error("Некорректный запрос в HH")
	errorResp := hhapimodels.ErrorData{}
	if response != nil && responseBody != nil {
		i.auditError(ctx, string(responseBody), response.StatusCode)
		err = json.Unmarshal(responseBody, &errorResp)
		if err != nil {
			logger.WithError(err).Error("ошибка сериализации ответа с ошибкой")
		} else {
			return &errorResp, errors.Errorf("Некорректный запрос. Ошибка: %+v", errorResp)
		}
	}
	return nil, errors.Errorf("Некорректный запрос. Ошибка: %+v", errorResp)
}

func getResponseBody(logger *log.Entry, response *http.Response) ([]byte, *log.Entry) {
	if response != nil {
		responseBody, _ := io.ReadAll(response.Body)
		return responseBody, logger.WithField("response_body", string(responseBody))
	}
	return nil, logger
}

func addStatusCode(logger *log.Entry, response *http.Response) *log.Entry {
	if response != nil {
		return logger.WithField("response_status_code", response.StatusCode)
	}
	return logger
}

func (i impl) auditError(ctx context.Context, response string, status int) {
	ctxData := externalservices.ExtractAuditData(ctx)
	if !ctxData.WithAudit {
		return
	}
	rec := dbmodels.ExtApiAudit{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: ctxData.SpaceID,
		},
		RecID:    ctxData.RecID,
		Service:  serviceName,
		Uri:      ctxData.Uri,
		Request:  ctxData.Request,
		Response: response,
		Status:   status,
	}
	i.auditStore.Create(rec)
}
