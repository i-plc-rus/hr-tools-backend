package client

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
	VacancyClose(ctx context.Context, accessToken string, employerID, vacancyID string) error

	//https://api.hh.ru/openapi/redoc#tag/Otklikipriglasheniya-rabotodatelya/operation/get-negotiations
	//https://api.hh.ru/openapi/redoc#tag/Otklikipriglasheniya-rabotodatelya/operation/get-collection-negotiations-list
	Negotiations(ctx context.Context, accessToken, vacancyID string, page, perPage int) (hhapimodels.NegotiationResponse, error)

	//https://api.hh.ru/openapi/redoc#tag/Otklikipriglasheniya-rabotodatelya/operation/post-negotiations-topics-read
	NegotiationMarkRead(ctx context.Context, accessToken, vacancyID, negotiationID string) error

	//https://api.hh.ru/openapi/redoc#tag/Prosmotr-rezyume/operation/get-resume
	GetResume(ctx context.Context, accessToken, resumeUrl string) (hhapimodels.ResumeResponse, error)

	GetAreas(ctx context.Context) ([]hhapimodels.Area, error)
}

var Instance Provider

type impl struct {
	host        string
	redirectUri string
}

func NewProvider(host, redirectUri string) {
	Instance = &impl{
		host:        host,
		redirectUri: redirectUri,
	}
}

const (
	mePath                    string = "/me"
	tokenPath                 string = "/token"
	oAuthPattern              string = "https://hh.ru/oauth/authorize?response_type=code&client_id=%v&state=%v&redirect_uri=%v"
	vPublishPath              string = "/vacancies"
	vUpdatePath               string = "/vacancies/%v"
	vDeletePath               string = "/employers/%v/vacancies/%v"
	negotiationCollectionPath string = "/negotiations?vacancy_id=%v"
	negotiationCollectionTpl  string = "%v&page=%v&per_page=%v"
	negotiationReadPath       string = "/negotiations/read"
	areasPath                 string = "areas"
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
	uri := i.host + tokenPath
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

	err = i.sendRequest(logger, r, &resp, "")
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (i impl) RefreshToken(ctx context.Context, req hhapimodels.RefreshToken) (*hhapimodels.ResponseToken, error) {
	uri := i.host + tokenPath
	data := url.Values{}
	data.Add("refresh_token", req.RefreshToken)
	data.Set("grant_type", "refresh_token")

	r, _ := http.NewRequestWithContext(ctx, "POST", uri, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp := hhapimodels.ResponseToken{}

	logger := log.
		WithField("external_request", uri).
		WithField("request_body", fmt.Sprintf("%+v", data.Encode()))

	err := i.sendRequest(logger, r, &resp, "")
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (i impl) Me(ctx context.Context, accessToken string) (*hhapimodels.MeResponse, error) {
	uri := i.host + mePath
	logger := log.
		WithField("external_request", uri)
	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.MeResponse{}
	err := i.sendRequest(logger, r, &resp, accessToken)
	return &resp, err
}

func (i impl) VacancyPublish(ctx context.Context, accessToken string, request hhapimodels.VacancyPubRequest) (vacancyID string, err error) {
	uri := i.host + vPublishPath
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

	err = i.sendRequest(logger, r, &resp, accessToken)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (i impl) VacancyUpdate(ctx context.Context, accessToken, vacancyID string, request hhapimodels.VacancyPubRequest) error {
	uri := i.host + fmt.Sprintf(vUpdatePath, vacancyID)
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

	return i.sendRequest(logger, r, nil, accessToken)
}

func (i impl) VacancyClose(ctx context.Context, accessToken, employerID, vacancyID string) error {
	uri := i.host + fmt.Sprintf(vDeletePath, employerID, vacancyID)
	logger := log.
		WithField("vacancy_id", vacancyID).
		WithField("employer_id", employerID).
		WithField("external_request", uri)
	r, _ := http.NewRequestWithContext(ctx, "PUT", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	return i.sendRequest(logger, r, nil, accessToken)
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
	uri := i.host + negotiationReadPath
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

	err = i.sendRequest(logger, r, &resp, accessToken)
	if err != nil {
		return err
	}
	return nil
}

func (i impl) GetResume(ctx context.Context, accessToken, resumeUrl string) (hhapimodels.ResumeResponse, error) {
	logger := log.
		WithField("external_request", resumeUrl)

	r, _ := http.NewRequestWithContext(ctx, "GET", resumeUrl, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.ResumeResponse{}

	err := i.sendRequest(logger, r, &resp, accessToken)
	if err != nil {
		return hhapimodels.ResumeResponse{}, err
	}
	return resp, nil
}

func (i impl) getNegotiations(ctx context.Context, accessToken, vacancyID string) (*hhapimodels.NegotiationCollections, error) {
	uri := i.host + fmt.Sprintf(negotiationCollectionPath, vacancyID)
	logger := log.
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := hhapimodels.NegotiationCollections{}

	err := i.sendRequest(logger, r, &resp, accessToken)
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

	err := i.sendRequest(logger, r, &resp, accessToken)
	if err != nil {
		return hhapimodels.NegotiationResponse{}, err
	}
	return resp, nil
}

func (i impl) GetAreas(ctx context.Context) ([]hhapimodels.Area, error) {
	uri := i.host + areasPath
	logger := log.
		WithField("external_request", uri)

	r, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	r.Header.Add("Content-Type", "application/json")
	resp := []hhapimodels.Area{}

	err := i.sendRequest(logger, r, &resp, "")
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

	errorResp := hhapimodels.ErrorData{}
	if response != nil {
		responseBody, _ := io.ReadAll(response.Body)
		logger = logger.WithField("response_body", string(responseBody))
		err = json.Unmarshal(responseBody, &errorResp)
		if err != nil {
			logger.WithError(err).Error("ошибка сериализации ответа")
		}

	}
	logger.Error("ошибка отправки запроса в HH")
	if response.StatusCode == 403 {
		err = errors.Errorf("Ошибка: %v, Причины: %+v", errorResp.Error, errorResp.Errors)
		return errors.New("Необходима повторная ааторизация")
	}
	return errors.New("Некорректный запрос")
}
