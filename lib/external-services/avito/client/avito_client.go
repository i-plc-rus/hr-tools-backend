package avitoclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	avitoapimodels "hr-tools-backend/models/api/avito"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Provider interface {
	//https://developers.avito.ru/api-catalog/auth/documentation#ApiDescriptionBlock
	GetLoginUri(clientID, spaceID string) (string, error)
	RequestToken(ctx context.Context, req avitoapimodels.RequestToken) (*avitoapimodels.ResponseToken, error)
	RefreshToken(ctx context.Context, req avitoapimodels.RefreshToken) (*avitoapimodels.ResponseToken, error)

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
}

var Instance Provider

type impl struct {
	host string
}

func NewProvider() {
	Instance = &impl{
		host: host,
	}
}

const (
	host               string = "https://api.avito.ru"
	tokenPath          string = "/token"
	oAuthPattern       string = "https://avito.ru/oauth?response_type=code&client_id=%v&scope=job:cv,job:applications,job:vacancy,job:write&state=%v"
	vPublishPath       string = "/job/v2/vacancies"
	vPublishStatusPath string = "/job/v2/vacancies/statuses"
	vUpdatePath        string = "/job/v2/vacancies/%v"
	vArchivePath       string = "/job/v1/vacancies/archived/%v"
	vGetPath           string = "/job/v2/vacancies/%v"
)

func (i impl) GetLoginUri(clientID, spaceID string) (string, error) {
	uri := fmt.Sprintf(oAuthPattern, clientID, spaceID)
	return uri, nil
}

func (i impl) RequestToken(ctx context.Context, req avitoapimodels.RequestToken) (*avitoapimodels.ResponseToken, error) {
	uri := i.host + tokenPath
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

	err := i.sendRequest(logger, r, &resp, "")
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (i impl) RefreshToken(ctx context.Context, req avitoapimodels.RefreshToken) (*avitoapimodels.ResponseToken, error) {
	uri := i.host + tokenPath
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

	err := i.sendRequest(logger, r, &resp, "")
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (i impl) VacancyPublish(ctx context.Context, accessToken string, request avitoapimodels.VacancyPubRequest) (publishID string, err error) {
	uri := i.host + vPublishPath
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

	err = i.sendRequest(logger, r, &resp, accessToken)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (i impl) VacancyStatus(ctx context.Context, accessToken string, request avitoapimodels.StatusRequest) (resp *avitoapimodels.StatusResponse, err error) {
	uri := i.host + vPublishStatusPath
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

	err = i.sendRequest(logger, r, resp, accessToken)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i impl) VacancyUpdate(ctx context.Context, accessToken, vacancyPublishID string, vacancyID int, request avitoapimodels.VacancyPubRequest) (publishID string, err error) {
	uri := i.host + fmt.Sprintf(vUpdatePath, vacancyPublishID)
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
	err = i.sendRequest(logger, r, &resp, accessToken)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (i impl) VacancyClose(ctx context.Context, accessToken string, vacancyID int) error {
	uri := i.host + fmt.Sprintf(vArchivePath, vacancyID)
	logger := log.
		WithField("avito_vacancy_id", vacancyID).
		WithField("external_request", uri)
	r, _ := http.NewRequestWithContext(ctx, "PUT", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	return i.sendRequest(logger, r, nil, accessToken)
}

func (i impl) GetVacancy(ctx context.Context, accessToken string, vacancyID int) (resp *avitoapimodels.VacancyInfo, err error) {
	uri := i.host + fmt.Sprintf(vGetPath, vacancyID)
	logger := log.
		WithField("avito_vacancy_id", vacancyID).
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp = new(avitoapimodels.VacancyInfo)

	err = i.sendRequest(logger, r, resp, accessToken)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i impl) sendRequest(logger *log.Entry, r *http.Request, resp interface{}, accessToken string) error {
	r.Header.Add("User-Agent", "HRTools/1.0")
	if accessToken != "" {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	}
	client := &http.Client{}
	response, err := client.Do(r)
	if response != nil && (response.StatusCode >= 200 && response.StatusCode <= 300) {
		if resp != nil {
			responseBody, _ := io.ReadAll(response.Body)
			err = json.Unmarshal(responseBody, resp)
			if err != nil {
				return errors.Wrap(err, "ошибка сериализации ответа")
			}
		}
		return nil
	}

	errorResp := avitoapimodels.ErrorData{}
	if response != nil {
		responseBody, _ := io.ReadAll(response.Body)
		logger = logger.WithField("response_body", string(responseBody))
		err = json.Unmarshal(responseBody, &errorResp)
		if err != nil {
			logger.WithError(err).Error("ошибка сериализации ответа")
		}

	}
	logger.Error("ошибка отправки запроса в Avito")
	switch response.StatusCode {
	case 400:
		return errors.Errorf("Некорректный запрос. Ошибка: %+v", errorResp.Error.Err400)
	case 401:
		return errors.Errorf("Некорректный запрос. Ошибка: %+v", errorResp.Error.Err401)
	case 402 - 404:
		return errors.Errorf("Некорректный запрос. Ошибка: %+v", errorResp.Error.Err402X)
	case 500 - 503:
		return errors.Errorf("Некорректный запрос. Ошибка: %+v", errorResp.Error.Err5XX)
	}
	return errors.Errorf("Некорректный запрос. Ошибка: %+v", errorResp)
}
