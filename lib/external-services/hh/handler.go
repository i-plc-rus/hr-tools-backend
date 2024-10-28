package hhhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	hhclient "hr-tools-backend/lib/external-services/hh/client"
	extservicestore "hr-tools-backend/lib/external-services/store"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	"hr-tools-backend/models"
	hhapimodels "hr-tools-backend/models/api/hh"
	dbmodels "hr-tools-backend/models/db"
	"strings"
	"sync"
	"time"
)

type Provider interface {
	GetConnectUri(spaceID string) (uri string, err error)
	RequestToken(spaceID, code string)
	CheckConnected(spaceID string) bool
	VacancyPublish(ctx context.Context, spaceID, vacancyID string) (vacancyUrl string, err error)
	VacancyUpdate(ctx context.Context, spaceID, vacancyID string) error
	VacancyClose(ctx context.Context, spaceID, vacancyID string) error
	VacancyAttach(ctx context.Context, spaceID, vacancyID, hhID string) error
}

var Instance Provider

func NewHandler() {
	Instance = &impl{
		client:             hhclient.Instance,
		extStore:           extservicestore.NewInstance(db.DB),
		spaceUserStore:     spaceusersstore.NewInstance(db.DB),
		vacancyStore:       vacancystore.NewInstance(db.DB),
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
		tokenMap:           sync.Map{},
		cityMap:            map[string]string{},
	}
}

type impl struct {
	client             hhclient.Provider
	extStore           extservicestore.Provider
	spaceUserStore     spaceusersstore.Provider
	vacancyStore       vacancystore.Provider
	spaceSettingsStore spacesettingsstore.Provider
	tokenMap           sync.Map
	cityMap            map[string]string
}

const (
	TokenCode     = "HH_TOKEN"
	VacancyUriTpl = "https://hh.ru/vacancy/%v"
)

func (i *impl) GetConnectUri(spaceID string) (uri string, err error) {
	clientID, err := i.getValue(spaceID, models.HhClientIDSetting)
	if err != nil {
		return "", errors.New("ошибка получения настройки ClientID для HH")
	}
	_, err = i.getValue(spaceID, models.HhClientSecretSetting)
	if err != nil {
		return "", errors.New("ошибка получения настройки ClientSecret для HH")
	}
	return i.client.GetLoginUri(clientID, spaceID)
}

func (i *impl) RequestToken(spaceID, code string) {
	logger := log.WithField("space_id", spaceID)
	clientID, err := i.getValue(spaceID, models.HhClientIDSetting)
	if err != nil {
		logger.WithError(err).Error("ошибка получения настройки ClientID для HH")
		return
	}

	clientSecret, err := i.getValue(spaceID, models.HhClientSecretSetting)
	if err != nil {
		logger.WithError(err).Error("ошибка получения настройки ClientSecret для HH")
		return
	}

	req := hhapimodels.RequestToken{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Code:         code,
		RedirectUri:  "",
	}
	token, err := i.client.RequestToken(context.TODO(), req)
	if err != nil {
		logger.WithError(err).Error("ошибка получения токена HH")
		return
	}
	err = i.storeToken(spaceID, *token, true)
	if err != nil {
		logger.Error(err)
	}
}

func (i *impl) CheckConnected(spaceID string) bool {
	_, ok := i.tokenMap.Load(spaceID)
	if ok {
		return true
	}
	data, ok, err := i.extStore.Get(spaceID, TokenCode)
	if err != nil || !ok {
		return false
	}
	token := hhapimodels.ResponseToken{}
	err = json.Unmarshal(data, &token)
	if err != nil {
		return false
	}
	i.storeToken(spaceID, token, false)
	return true
}

func (i *impl) VacancyPublish(ctx context.Context, spaceID, vacancyID string) (vacancyUrl string, err error) {
	logger := log.
		WithField("space_id", spaceID).
		WithField("vacancy_id", vacancyID)

	accessToken, err := i.getToken(ctx, spaceID)
	if err != nil {
		return "", err
	}

	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return "", err
	}
	if rec == nil {
		return "", errors.New("вакансия не найдена")
	}

	if models.VacancyStatusOpened != rec.Status {
		return "", errors.Errorf("неподходящей статус вакансии %v, для публикации в НН", rec.Status)
	}

	if rec.HhID != "" {
		return "", errors.New("вакансия уже опубликованна")
	}

	request, err := i.fillVacancyData(ctx, rec)
	if err != nil {
		return "", err
	}

	id, err := i.client.VacancyPublish(ctx, accessToken, *request)
	if err != nil {
		return "", err
	}
	vacancyUrl = fmt.Sprintf(VacancyUriTpl, id)
	updMap := map[string]interface{}{
		"hh_id":  id,
		"hh_uri": vacancyUrl,
	}
	err = i.vacancyStore.Update(spaceID, vacancyID, updMap)
	if err != nil {
		err = errors.Errorf("не удалось сохранить идентификатор опубликованной вакансии (%v)", id)
		logger.Error(err)
		return vacancyUrl, err
	}
	return vacancyUrl, nil
}

func (i *impl) VacancyUpdate(ctx context.Context, spaceID, vacancyID string) error {
	accessToken, err := i.getToken(ctx, spaceID)
	if err != nil {
		return err
	}

	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("вакансия не найдена")
	}
	if rec.HhID == "" {
		return errors.New("вакансия еще не опубликованна")
	}

	request, err := i.fillVacancyData(ctx, rec)
	if err != nil {
		return err
	}

	return i.client.VacancyUpdate(ctx, accessToken, rec.HhID, *request)
}

func (i *impl) VacancyClose(ctx context.Context, spaceID, vacancyID string) error {
	accessToken, err := i.getToken(ctx, spaceID)
	if err != nil {
		return err
	}
	meResp, err := i.client.Me(ctx, accessToken)
	if err != nil {
		return errors.Wrap(err, "ошибка получения информации о токене HH")
	}
	return i.client.VacancyClose(ctx, accessToken, meResp.Employer.ID, vacancyID)
}

func (i *impl) VacancyAttach(ctx context.Context, spaceID, vacancyID, hhID string) error {
	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("вакансия не найдена")
	}
	if rec.HhID != "" {
		return errors.New("ссылка на вакансию уже добавлена")
	}
	if models.VacancyStatusOpened != rec.Status {
		return errors.Errorf("неподходящей статус вакансии: %v", rec.Status)
	}
	vacancyUrl := fmt.Sprintf(VacancyUriTpl, hhID)
	updMap := map[string]interface{}{
		"hh_id":  hhID,
		"hh_uri": vacancyUrl,
	}
	logger := log.
		WithField("space_id", spaceID).
		WithField("vacancy_id", vacancyID)
	err = i.vacancyStore.Update(spaceID, vacancyID, updMap)
	if err != nil {
		err = errors.Errorf("не удалось сохранить идентификатор опубликованной вакансии (%v)", hhID)
		logger.Error(err)
		return err
	}
	return nil
}

func (i *impl) getValue(spaceID string, code models.SpaceSettingCode) (string, error) {
	return i.spaceSettingsStore.GetValueByCode(spaceID, string(code))
}

func (i *impl) fillAreas(areas []hhapimodels.Area) {
	for _, area := range areas {
		if len(area.Areas) == 0 {
			i.cityMap[area.Name] = area.ID
			continue
		}
		i.fillAreas(area.Areas)
	}
}

func (i *impl) getArea(ctx context.Context, city *dbmodels.City) (hhapimodels.DictItem, error) {
	if len(i.cityMap) == 0 {
		areas, err := i.client.GetAreas(ctx)
		for _, area := range areas {
			if area.Name == "Россия" {
				i.fillAreas(area.Areas)
				break
			}
		}
		return hhapimodels.DictItem{}, err
	}
	id, ok := i.cityMap[city.City]
	if ok {
		return hhapimodels.DictItem{
			ID: id,
		}, nil
	}
	cityRegion := fmt.Sprintf("%v (%v", city.City, city.Region)
	for key, value := range i.cityMap {
		if strings.HasPrefix(key, cityRegion) {
			return hhapimodels.DictItem{
				ID: value,
			}, nil
		}
	}
	return hhapimodels.DictItem{}, errors.New("город публикации не найден в справочнике HH")
}

func (i *impl) storeToken(spaceID string, token hhapimodels.ResponseToken, inDb bool) error {
	tokenData := hhapimodels.TokenData{
		ResponseToken: token,
		ExpiresAt:     time.Now(),
	}
	i.tokenMap.Store(spaceID, tokenData)
	if inDb {
		data, err := json.Marshal(token)
		err = i.extStore.Set(spaceID, TokenCode, data)
		if err != nil {
			return errors.Wrap(err, "ошибка сохранения токена HH в бд")
		}
	}
	return nil
}

func (i *impl) getToken(ctx context.Context, spaceID string) (string, error) {
	if !i.CheckConnected(spaceID) {
		return "", errors.New("HeadHunter не подключен")
	}
	value, ok := i.tokenMap.Load(spaceID)
	if !ok {
		return "", errors.New("HeadHunter не подключен")
	}
	tokenData := value.(hhapimodels.TokenData)
	if time.Now().After(tokenData.ExpiresAt) {
		req := hhapimodels.RefreshToken{
			RefreshToken: tokenData.RefreshToken,
		}
		tokenResp, err := i.client.RefreshToken(ctx, req)
		if err != nil {
			return "", errors.New("ошибка получения токена для HeadHunter")
		}
		err = i.storeToken(spaceID, *tokenResp, true)
		if err != nil {
			return "", errors.New("ошибка сохранения токена для HeadHunter")
		}
	}
	return tokenData.AccessToken, nil
}

func (i *impl) fillVacancyData(ctx context.Context, rec *dbmodels.Vacancy) (*hhapimodels.VacancyPubRequest, error) {
	if rec.City == nil {
		return nil, errors.New("не указан город публикации")
	}
	area, err := i.getArea(ctx, rec.City)
	if err != nil {
		return nil, err
	}

	request := hhapimodels.VacancyPubRequest{
		Description: rec.Requirements,
		Name:        rec.VacancyName,
		Area:        area,
		//Employment:        hhapimodels.DictItem{},
		//Schedule:          hhapimodels.DictItem{},
		//Experience:        hhapimodels.DictItem{},
		//Salary: 			 hhapimodels.Salary{},
		//Contacts:          hhapimodels.Contacts{},
		ProfessionalRoles: nil, //!!todo
		BillingType: hhapimodels.DictItem{
			ID: "free",
		},
		Type: hhapimodels.DictItem{
			ID: "open",
		},
	}
	salary := hhapimodels.Salary{Currency: "RUR"}
	if rec.Salary.InHand != 0 {
		salary.From = rec.Salary.InHand
		salary.To = rec.Salary.InHand
		salary.Gross = false
		request.Salary = &salary
	} else if rec.Salary.From != 0 || rec.Salary.To != 0 {
		salary = hhapimodels.Salary{
			From:  rec.Salary.From,
			To:    rec.Salary.To,
			Gross: true,
		}
		request.Salary = &salary
	}
	return &request, nil
}
