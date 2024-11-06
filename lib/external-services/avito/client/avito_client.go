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

	//https://developers.avito.ru/api-catalog/job/documentation#operation/applicationsGetIds
	GetApplicationIDs(ctx context.Context, accessToken string, updatedAt, lastID string, vacancyID int) (resp *avitoapimodels.AppliesIDResponse, err error)

	//https://developers.avito.ru/api-catalog/job/documentation#operation/applicationsGetByIds
	GetApplicationByIDs(ctx context.Context, accessToken string, ids []string, vacancyID int) (resp *avitoapimodels.AppliesResponse, err error)

	//https://developers.avito.ru/api-catalog/job/documentation#operation/resumeGetItem
	GetResume(ctx context.Context, accessToken string, vacancyID, resumeID int) (resp *avitoapimodels.Resume, err error)
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
	host                 string = "https://api.avito.ru"
	tokenPath            string = "%s/token"
	oAuthPattern         string = "https://avito.ru/oauth?response_type=code&client_id=%v&scope=job:cv,job:applications,job:vacancy,job:write&state=%v"
	vPublishPath         string = "%s/job/v2/vacancies"
	vPublishStatusPath   string = "%s/job/v2/vacancies/statuses"
	vUpdatePath          string = "%s/job/v2/vacancies/%v"
	vArchivePath         string = "%s/job/v1/vacancies/archived/%v"
	vGetPath             string = "%s/job/v2/vacancies/%v"
	vGetListPath         string = "%s/core/v1/items?category=111&status=active&page=%v&per_page=50"
	vGetApplicationIDs   string = "%s/job/v1/applications/get_ids?updatedAtFrom=%v&vacancyIds=%v"
	vGetApplicationByIDs string = "%s/job/v1/applications/get_by_ids"
	vGetResume           string = "%s/job/v2/resumes/%v"
)

func (i impl) GetLoginUri(clientID, spaceID string) (string, error) {
	uri := fmt.Sprintf(oAuthPattern, clientID, spaceID)
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

	err := i.sendRequest(logger, r, &resp, "")
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

	err := i.sendRequest(logger, r, &resp, "")
	if err != nil {
		return nil, err
	}
	return &resp, nil
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

	err = i.sendRequest(logger, r, &resp, accessToken)
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

	err = i.sendRequest(logger, r, resp, accessToken)
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
	err = i.sendRequest(logger, r, &resp, accessToken)
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
	return i.sendRequest(logger, r, nil, accessToken)
}

func (i impl) GetVacancy(ctx context.Context, accessToken string, vacancyID int) (resp *avitoapimodels.VacancyInfo, err error) {
	uri := fmt.Sprintf(vGetPath, i.host, vacancyID)
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

	err = i.sendRequest(logger, r, resp, accessToken)
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

	err = i.sendRequest(logger, r, resp, accessToken)
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

	responseError := ""
	if response != nil {
		responseBody, _ := io.ReadAll(response.Body)
		logger = logger.WithField("response_body", string(responseBody))
		responseError = string(responseBody)
	}
	logger.Error("ошибка отправки запроса в Avito")
	return errors.Errorf("Некорректный запрос. Ошибка: %v", responseError)
}
