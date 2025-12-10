package avitoclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hr-tools-backend/db"
	externalservices "hr-tools-backend/lib/external-services"
	extapiauditstore "hr-tools-backend/lib/external-services/ext-api-audit-store"
	avitoapimodels "hr-tools-backend/models/api/avito"
	dbmodels "hr-tools-backend/models/db"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	//https://developers.avito.ru/api-catalog/auth/documentation#ApiDescriptionBlock
	GetLoginUri(clientID, spaceID string) (string, error)
	RequestToken(ctx context.Context, req avitoapimodels.RequestToken) (*avitoapimodels.ResponseToken, error)
	RefreshToken(ctx context.Context, req avitoapimodels.RefreshToken) (*avitoapimodels.ResponseToken, error)
	Self(ctx context.Context, accessToken string) (*avitoapimodels.SelfData, error)

	//https://developers.avito.ru/api-catalog/job/documentation#operation/vacancyCreateV2
	VacancyPublish(ctx context.Context, accessToken string, request avitoapimodels.VacancyPubRequest) (publishID string, err error)

	//https://developers.avito.ru/api-catalog/job/documentation#operation/vacancyGetStatuses
	VacancyStatus(ctx context.Context, accessToken string, request avitoapimodels.StatusRequest) (resp *avitoapimodels.StatusResponse, err error)

	//https://developers.avito.ru/api-catalog/job/documentation#operation/vacancyUpdateV2
	VacancyUpdate(ctx context.Context, accessToken, vacancyPublishID string, vacancyID int, request avitoapimodels.VacancyPubRequest) (publishID string, err error)

	//https://developers.avito.ru/api-catalog/job/documentation#operation/vacancyArchive
	VacancyClose(ctx context.Context, accessToken string, vacancyID int) error

	//https://developers.avito.ru/api-catalog/job/documentation#operation/vacancyGetItem
	GetVacancy(ctx context.Context, accessToken string, vacancyID int) (*avitoapimodels.VacancyInfo, error)

	//https://developers.avito.ru/api-catalog/job/documentation#operation/applicationsGetIds
	GetApplicationIDs(ctx context.Context, accessToken string, updatedAt, lastID string, vacancyID int) (resp *avitoapimodels.AppliesIDResponse, err error)

	//https://developers.avito.ru/api-catalog/job/documentation#operation/applicationsGetByIds
	GetApplicationByIDs(ctx context.Context, accessToken string, ids []string, vacancyID int) (resp *avitoapimodels.AppliesResponse, err error)

	//https://developers.avito.ru/api-catalog/job/documentation#operation/resumeGetItem
	GetResume(ctx context.Context, accessToken string, vacancyID, resumeID int) (resp *avitoapimodels.Resume, err error)

	//https://developers.avito.ru/api-catalog/messenger/documentation#operation/postSendMessage
	SendNewMessage(ctx context.Context, accessToken string, userID int64, chatID, msg string) error

	//https://developers.avito.ru/api-catalog/messenger/documentation#operation/chatRead
	MarkReadMessage(ctx context.Context, accessToken string, userID int64, chatID string) error

	// https://developers.avito.ru/api-catalog/messenger/documentation#operation/getMessagesV3
	GetMessages(ctx context.Context, accessToken string, userID int64, chatID string) (avitoapimodels.MessageResponse, error)

	// https://developers.avito.ru/api-catalog/messenger/documentation#operation/getChatByIdV2
	GetChatInfo(ctx context.Context, accessToken string, userID int64, chatID string) (avitoapimodels.ChatInfo, error)
}

var Instance Provider

type impl struct {
	host       string
	auditStore extapiauditstore.Provider
}

func NewProvider() {
	Instance = &impl{
		host:       host,
		auditStore: extapiauditstore.NewInstance(db.DB),
	}
}

const (
	serviceName          string = "avito"
	host                 string = "https://api.avito.ru"
	tokenPath            string = "%s/token"
	tokenScope           string = "job:cv,job:write,job:applications,messenger:read,messenger:write,stats:read,job:vacancy,user:read"
	oAuthPattern         string = "https://avito.ru/oauth?response_type=code&client_id=%v&scope=%v&state=%v"
	vPublishPath         string = "%s/job/v2/vacancies"
	selfPath             string = "%s/core/v1/accounts/self"
	vPublishStatusPath   string = "%s/job/v2/vacancies/statuses"
	vUpdatePath          string = "%s/job/v2/vacancies/%v"
	vArchivePath         string = "%s/job/v1/vacancies/archived/%v"
	vGetPath             string = "%s/job/v2/vacancies/%v"
	vGetListPath         string = "%s/core/v1/items?category=111&status=active&page=%v&per_page=50"
	vGetApplicationIDs   string = "%s/job/v1/applications/get_ids?updatedAtFrom=%v&vacancyIds=%v"
	vGetApplicationByIDs string = "%s/job/v1/applications/get_by_ids"
	vGetResume           string = "%s/job/v2/resumes/%v"
	messagesNew          string = "%s/messenger/v1/accounts/%v/chats/%v/messages"
	messagesRead         string = "%s/messenger/v1/accounts/%v/chats/%v/read"
	messagesList         string = "%s/messenger/v3/accounts/%v/chats/%v/messages"
	chatInfo             string = "%s/messenger/v2/accounts/%v/chats/%v"
)

func (i impl) GetLoginUri(clientID, spaceID string) (string, error) {
	uri := fmt.Sprintf(oAuthPattern, clientID, tokenScope, spaceID)
	return uri, nil
}

func (i impl) RequestToken(ctx context.Context, req avitoapimodels.RequestToken) (*avitoapimodels.ResponseToken, error) {
	uri := fmt.Sprintf(tokenPath, i.host)
	data := url.Values{}
	data.Set("client_id", req.ClientID)
	data.Set("client_secret", req.ClientSecret)
	data.Set("code", req.Code)
	data.Set("grant_type", "authorization_code")

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp := avitoapimodels.ResponseToken{}

	logger := log.
		WithField("external_request", uri).
		WithField("request_body", fmt.Sprintf("%+v", data.Encode()))

	err := i.sendRequest(ctx, logger, r, &resp, "")
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (i impl) RefreshToken(ctx context.Context, req avitoapimodels.RefreshToken) (*avitoapimodels.ResponseToken, error) {
	uri := fmt.Sprintf(tokenPath, i.host)
	data := url.Values{}
	data.Set("client_id", req.ClientID)
	data.Set("client_secret", req.ClientSecret)
	data.Add("refresh_token", req.RefreshToken)
	data.Set("grant_type", "refresh_token")

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp := avitoapimodels.ResponseToken{}

	logger := log.
		WithField("external_request", uri).
		WithField("request_body", fmt.Sprintf("%+v", data.Encode()))

	err := i.sendRequest(ctx, logger, r, &resp, "")
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (i impl) Self(ctx context.Context, accessToken string) (*avitoapimodels.SelfData, error) {
	uri := fmt.Sprintf(selfPath, i.host)
	logger := log.
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := new(avitoapimodels.SelfData)

	err := i.sendRequest(ctx, logger, r, resp, accessToken)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i impl) VacancyPublish(ctx context.Context, accessToken string, request avitoapimodels.VacancyPubRequest) (publishID string, err error) {
	uri := fmt.Sprintf(vPublishPath, i.host)
	logger := log.
		WithField("external_request", uri)
	body, err := json.Marshal(request)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации запроса")
	}

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, bytes.NewBuffer(body))
	r.Header.Add("Content-Type", "application/json")
	resp := avitoapimodels.VacancyPubResponse{}

	logger = logger.
		WithField("request_body", string(body))

	rCtx := externalservices.GetAuditContext(ctx, uri, body)
	err = i.sendRequest(rCtx, logger, r, &resp, accessToken)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (i impl) VacancyStatus(ctx context.Context, accessToken string, request avitoapimodels.StatusRequest) (resp *avitoapimodels.StatusResponse, err error) {
	uri := fmt.Sprintf(vPublishStatusPath, i.host)
	logger := log.
		WithField("external_request", uri)
	body, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка десериализации запроса")
	}

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, bytes.NewBuffer(body))
	r.Header.Add("Content-Type", "application/json")
	resp = new(avitoapimodels.StatusResponse)

	logger = logger.
		WithField("request_body", string(body))

	err = i.sendRequest(ctx, logger, r, resp, accessToken)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i impl) VacancyUpdate(ctx context.Context, accessToken, vacancyPublishID string, vacancyID int, request avitoapimodels.VacancyPubRequest) (publishID string, err error) {
	uri := fmt.Sprintf(vUpdatePath, i.host, vacancyPublishID)
	logger := log.
		WithField("avito_vacancy_id", vacancyID).
		WithField("vacancy_publish_id", vacancyPublishID).
		WithField("external_request", uri)
	body, err := json.Marshal(request)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации запроса")
	}

	r, _ := http.NewRequestWithContext(ctx, "PUT", uri, bytes.NewBuffer(body))
	r.Header.Add("Content-Type", "application/json")

	logger = logger.
		WithField("request_body", string(body))
	resp := avitoapimodels.VacancyPubResponse{}
	rCtx := externalservices.GetAuditContext(ctx, uri, body)
	err = i.sendRequest(rCtx, logger, r, &resp, accessToken)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (i impl) VacancyClose(ctx context.Context, accessToken string, vacancyID int) error {
	uri := fmt.Sprintf(vArchivePath, i.host, vacancyID)
	logger := log.
		WithField("avito_vacancy_id", vacancyID).
		WithField("external_request", uri)
	r, _ := http.NewRequestWithContext(ctx, "PUT", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	rCtx := externalservices.GetAuditContext(ctx, uri, nil)
	return i.sendRequest(rCtx, logger, r, nil, accessToken)
}

func (i impl) GetVacancy(ctx context.Context, accessToken string, vacancyID int) (resp *avitoapimodels.VacancyInfo, err error) {
	uri := fmt.Sprintf(vGetPath, i.host, vacancyID)
	logger := log.
		WithField("avito_vacancy_id", vacancyID).
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp = new(avitoapimodels.VacancyInfo)

	err = i.sendRequest(ctx, logger, r, resp, accessToken)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i impl) GetApplicationIDs(ctx context.Context, accessToken string, updatedAt, lastID string, vacancyID int) (resp *avitoapimodels.AppliesIDResponse, err error) {
	uri := fmt.Sprintf(vGetApplicationIDs, i.host, updatedAt, vacancyID)
	if lastID != "" {
		uri = fmt.Sprintf("%v&cursor=%v", uri, lastID)
	}
	logger := log.
		WithField("avito_vacancy_id", vacancyID).
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp = new(avitoapimodels.AppliesIDResponse)

	err = i.sendRequest(ctx, logger, r, resp, accessToken)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i impl) GetApplicationByIDs(ctx context.Context, accessToken string, ids []string, vacancyID int) (resp *avitoapimodels.AppliesResponse, err error) {
	uri := fmt.Sprintf(vGetApplicationByIDs, i.host)
	logger := log.
		WithField("avito_vacancy_id", vacancyID).
		WithField("external_request", uri)

	request := avitoapimodels.ApplicationRequest{
		IDs: ids,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка десериализации запроса")
	}

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, bytes.NewBuffer(body))
	r.Header.Add("Content-Type", "application/json")
	resp = new(avitoapimodels.AppliesResponse)

	err = i.sendRequest(ctx, logger, r, resp, accessToken)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i impl) GetResume(ctx context.Context, accessToken string, vacancyID, resumeID int) (resp *avitoapimodels.Resume, err error) {
	uri := fmt.Sprintf(vGetResume, i.host, resumeID)
	logger := log.
		WithField("avito_vacancy_id", vacancyID).
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp = new(avitoapimodels.Resume)

	err = i.sendRequest(ctx, logger, r, resp, accessToken)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i impl) SendNewMessage(ctx context.Context, accessToken string, userID int64, chatID, msg string) error {
	uri := fmt.Sprintf(messagesNew, i.host, userID, chatID)
	logger := log.
		WithField("external_request", uri)
	request := avitoapimodels.NewMessageRequest{}
	body, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "ошибка десериализации запроса")
	}

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, bytes.NewBuffer(body))
	r.Header.Add("Content-Type", "application/json")

	logger = logger.
		WithField("request_body", string(body))

	return i.sendRequest(ctx, logger, r, nil, accessToken)
}

func (i impl) MarkReadMessage(ctx context.Context, accessToken string, userID int64, chatID string) error {
	uri := fmt.Sprintf(messagesRead, i.host, userID, chatID)
	logger := log.
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	return i.sendRequest(ctx, logger, r, nil, accessToken)
}

func (i impl) GetMessages(ctx context.Context, accessToken string, userID int64, chatID string) (avitoapimodels.MessageResponse, error) {
	uri := fmt.Sprintf(messagesList, i.host, userID, chatID)
	logger := log.
		WithField("external_request", uri)

	resp := avitoapimodels.MessageResponse{}
	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")

	err := i.sendRequest(ctx, logger, r, &resp, accessToken)
	if err != nil {
		return avitoapimodels.MessageResponse{}, err
	}
	return resp, nil
}

func (i impl) GetChatInfo(ctx context.Context, accessToken string, userID int64, chatID string) (avitoapimodels.ChatInfo, error) {
	uri := fmt.Sprintf(chatInfo, i.host, userID, chatID)
	logger := log.
		WithField("external_request", uri)

	resp := avitoapimodels.ChatInfo{}
	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")

	err := i.sendRequest(ctx, logger, r, &resp, accessToken)
	if err != nil {
		return avitoapimodels.ChatInfo{}, err
	}
	return resp, nil
}

func (i impl) sendRequest(ctx context.Context, logger *log.Entry, r *http.Request, resp interface{}, accessToken string) error {
	r.Header.Add("User-Agent", "HRTools/1.0")
	if accessToken != "" {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	}
	client := &http.Client{}
	response, err := client.Do(r)
	// читаем Body только 1 раз
	responseBody, logger := getResponseBody(logger, response)
	if response != nil && (response.StatusCode >= 200 && response.StatusCode <= 300) {
		if resp != nil {
			err = json.Unmarshal(responseBody, resp)
			if err != nil {
				return errors.Wrap(err, "ошибка сериализации ответа")
			}
		}
		return nil
	}

	responseError := ""
	if response != nil {
		responseError = string(responseBody)
		i.auditError(ctx, responseError, response.StatusCode)
	}
	logger.Error("ошибка отправки запроса в Avito")
	return errors.Errorf("Некорректный запрос. Ошибка: %v", responseError)
}

func getResponseBody(logger *log.Entry, response *http.Response) ([]byte, *log.Entry) {
	if response != nil {
		responseBody, _ := io.ReadAll(response.Body)
		return responseBody, logger.WithField("response_body", string(responseBody))
	}
	return nil, logger
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
