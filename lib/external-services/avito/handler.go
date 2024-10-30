package avitohandler

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	externalservices "hr-tools-backend/lib/external-services"
	avitoclient "hr-tools-backend/lib/external-services/avito/client"
	extservicestore "hr-tools-backend/lib/external-services/store"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	"hr-tools-backend/models"
	avitoapimodels "hr-tools-backend/models/api/avito"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
	"strconv"
	"sync"
	"time"
)

var Instance externalservices.JobSiteProvider

func NewHandler() {
	Instance = &impl{
		client:             avitoclient.Instance,
		extStore:           extservicestore.NewInstance(db.DB),
		spaceUserStore:     spaceusersstore.NewInstance(db.DB),
		vacancyStore:       vacancystore.NewInstance(db.DB),
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
		tokenMap:           sync.Map{},
	}
}

type impl struct {
	client             avitoclient.Provider
	extStore           extservicestore.Provider
	spaceUserStore     spaceusersstore.Provider
	vacancyStore       vacancystore.Provider
	spaceSettingsStore spacesettingsstore.Provider
	tokenMap           sync.Map
}

const (
	TokenCode = "AVITO_TOKEN"
)

func (i *impl) getLogger(spaceID, vacancyID string) *log.Entry {
	logger := log.WithField("integration", "Avito")
	if spaceID != "" {
		logger = logger.WithField("space_id", spaceID)
	}
	if vacancyID != "" {
		logger = logger.WithField("vacancy_id", vacancyID)
	}
	return logger
}

func (i *impl) GetConnectUri(spaceID string) (uri string, err error) {
	clientID, err := i.getValue(spaceID, models.AvitoClientIDSetting)
	if err != nil {
		return "", errors.New("ошибка получения настройки ClientID для Avito")
	}
	_, err = i.getValue(spaceID, models.AvitoClientSecretSetting)
	if err != nil {
		return "", errors.New("ошибка получения настройки ClientSecret для Avito")
	}
	return i.client.GetLoginUri(clientID, spaceID)
}

func (i *impl) RequestToken(spaceID, code string) {
	logger := i.getLogger(spaceID, "")
	clientID, err := i.getValue(spaceID, models.AvitoClientIDSetting)
	if err != nil {
		logger.WithError(err).Error("ошибка получения настройки ClientID для Avito")
		return
	}

	clientSecret, err := i.getValue(spaceID, models.AvitoClientSecretSetting)
	if err != nil {
		logger.WithError(err).Error("ошибка получения настройки ClientSecret для Avito")
		return
	}

	req := avitoapimodels.RequestToken{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Code:         code,
	}
	token, err := i.client.RequestToken(context.TODO(), req)
	if err != nil {
		logger.WithError(err).Error("ошибка получения токена Avito")
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
	token := avitoapimodels.ResponseToken{}
	err = json.Unmarshal(data, &token)
	if err != nil {
		return false
	}
	i.storeToken(spaceID, token, false)
	return true
}

func (i *impl) VacancyPublish(ctx context.Context, spaceID, vacancyID string) error {
	logger := i.getLogger(spaceID, vacancyID)

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

	if models.VacancyStatusOpened != rec.Status {
		return errors.Errorf("неподходящей статус вакансии %v, для публикации в НН", rec.Status)
	}

	if rec.AvitoID != 0 || rec.AvitoPublishID != "" {
		if rec.AvitoStatus != models.VacancyPubStatusNone && rec.AvitoStatus != models.VacancyPubStatusClosed {
			return errors.New("вакансия уже размещена")
		}
	}

	request, err := i.fillVacancyData(ctx, rec)
	if err != nil {
		return err
	}

	id, err := i.client.VacancyPublish(ctx, accessToken, *request)
	if err != nil {
		return err
	}
	updMap := map[string]interface{}{
		"avito_publish_id": id,
		"avito_id":         nil,
		"avito_uri":        nil,
		"avito_reasons":    nil,
		"avito_status":     models.VacancyPubStatusModeration,
	}
	err = i.vacancyStore.Update(spaceID, vacancyID, updMap)
	if err != nil {
		errMsg := errors.Errorf("не удалось сохранить идентификатор опубликованной вакансии (%v)", id)
		logger.WithError(err).Error(errMsg)
		return errMsg
	}
	return nil
}

func (i *impl) VacancyUpdate(ctx context.Context, spaceID, vacancyID string) error {
	logger := i.getLogger(spaceID, vacancyID)
	accessToken, err := i.getToken(ctx, spaceID)
	if err != nil {
		return err
	}

	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return err
	}
	err = allowChange(rec, true)
	if err != nil {
		return err
	}

	request, err := i.fillVacancyData(ctx, rec)
	if err != nil {
		return err
	}

	id, err := i.client.VacancyUpdate(ctx, accessToken, rec.AvitoPublishID, rec.AvitoID, *request)
	if err != nil {
		return err
	}
	updMap := map[string]interface{}{
		"avito_publish_id": id,
		"avito_reasons":    nil,
		"avito_status":     models.VacancyPubStatusModeration,
	}
	err = i.vacancyStore.Update(spaceID, vacancyID, updMap)
	if err != nil {
		errMsg := errors.Errorf("не удалось сохранить идентификатор опубликованной вакансии (%v)", id)
		logger.WithError(err).Error(errMsg)
		return errMsg
	}
	return nil
}

func (i *impl) VacancyClose(ctx context.Context, spaceID, vacancyID string) error {
	accessToken, err := i.getToken(ctx, spaceID)
	if err != nil {
		return err
	}

	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return err
	}
	err = allowChange(rec, false)
	if err != nil {
		return err
	}
	return i.client.VacancyClose(ctx, accessToken, rec.AvitoID)
}

func (i *impl) VacancyAttach(ctx context.Context, spaceID, vacancyID string, extID string) error {
	avitoID, err := strconv.Atoi(extID)
	if err != nil {
		return errors.New("указане некорректный идентификатор вакансии")
	}
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
	if rec.AvitoID != 0 {
		return errors.New("ссылка на вакансию уже добавлена")
	}
	if models.VacancyStatusOpened != rec.Status {
		return errors.Errorf("неподходящей статус вакансии: %v", rec.Status)
	}
	data, err := i.client.GetVacancy(ctx, accessToken, avitoID)
	if err != nil {
		return err
	}
	if !data.IsActive {
		return errors.New("указанная вакансия уже не активна")
	}
	updMap := map[string]interface{}{
		"avito_id":         data.ID,
		"avito_uri":        data.Url,
		"avito_publish_id": nil,
		"avito_reasons":    nil,
		"avito_status":     models.VacancyPubStatusPublished,
	}
	logger := i.getLogger(spaceID, vacancyID)
	err = i.vacancyStore.Update(spaceID, vacancyID, updMap)
	if err != nil {
		err = errors.Errorf("не удалось обновить данные опубликованной вакансии (%v)", data.ID)
		logger.Error(err)
		return err
	}
	return nil
}

func (i *impl) GetVacancyInfo(ctx context.Context, spaceID, vacancyID string) (*vacancyapimodels.ExtVacancyInfo, error) {
	rec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errors.New("вакансия не найдена")
	}
	result := vacancyapimodels.ExtVacancyInfo{
		Url:    rec.AvitoUri,
		Status: models.VacancyPubStatusNone,
	}
	if rec.AvitoID == 0 && rec.AvitoPublishID == "" {
		return &result, nil
	}
	result.Status = rec.AvitoStatus
	result.Reason = rec.AvitoReasons
	return &result, nil
}

func (i *impl) getValue(spaceID string, code models.SpaceSettingCode) (string, error) {
	return i.spaceSettingsStore.GetValueByCode(spaceID, code)
}

func (i *impl) storeToken(spaceID string, token avitoapimodels.ResponseToken, inDb bool) error {
	tokenData := avitoapimodels.TokenData{
		ResponseToken: token,
		ExpiresAt:     time.Now(),
	}
	i.tokenMap.Store(spaceID, tokenData)
	if inDb {
		data, err := json.Marshal(token)
		err = i.extStore.Set(spaceID, TokenCode, data)
		if err != nil {
			return errors.Wrap(err, "ошибка сохранения токена Avito в бд")
		}
	}
	return nil
}

func (i *impl) getToken(ctx context.Context, spaceID string) (string, error) {
	if !i.CheckConnected(spaceID) {
		return "", errors.New("Avito не подключен")
	}
	value, ok := i.tokenMap.Load(spaceID)
	if !ok {
		return "", errors.New("Avito не подключен")
	}
	tokenData := value.(avitoapimodels.TokenData)
	if time.Now().After(tokenData.ExpiresAt) {
		clientID, err := i.getValue(spaceID, models.AvitoClientIDSetting)
		if err != nil {
			return "", errors.New("ошибка получения настройки ClientID для Avito")
		}

		clientSecret, err := i.getValue(spaceID, models.AvitoClientSecretSetting)
		if err != nil {
			return "", errors.New("ошибка получения настройки ClientSecret для Avito")
		}
		req := avitoapimodels.RefreshToken{
			RefreshToken: tokenData.RefreshToken,
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}
		tokenResp, err := i.client.RefreshToken(ctx, req)
		if err != nil {
			return "", errors.New("ошибка получения токена для Avito")
		}
		err = i.storeToken(spaceID, *tokenResp, true)
		if err != nil {
			return "", errors.New("ошибка сохранения токена для Avito")
		}
	}
	return tokenData.AccessToken, nil
}

func (i *impl) fillVacancyData(ctx context.Context, rec *dbmodels.Vacancy) (*avitoapimodels.VacancyPubRequest, error) {
	if rec.City == nil {
		return nil, errors.New("не указан город публикации")
	}

	businessArea := 0 //todo добавить заполение из справочника
	request := avitoapimodels.VacancyPubRequest{
		ApplyProcessing: avitoapimodels.ApplyProcessing{
			ApplyType: avitoapimodels.ApplyTypeWithResume,
		},
		BillingType:  "package",
		BusinessArea: businessArea,
		Description:  rec.Requirements,
		Employment:   rec.Employment,
		Experience:   rec.Experience,
		Location: avitoapimodels.Location{
			Address: avitoapimodels.LocationAddress{
				Locality: rec.City.City,
			},
		},
		Schedule: rec.Schedule,
		Title:    rec.VacancyName,
	}
	if rec.Salary.From != 0 || rec.Salary.To != 0 {
		request.SalaryRange = &avitoapimodels.SalaryRange{
			From: rec.From,
			To:   rec.To,
		}
	} else if rec.Salary.InHand != 0 {
		request.SalaryRange = &avitoapimodels.SalaryRange{
			From: rec.Salary.InHand,
			To:   rec.Salary.InHand,
		}
	}
	return &request, nil
}

func allowChange(rec *dbmodels.Vacancy, isEdit bool) error {
	if rec == nil {
		return errors.New("вакансия не найдена")
	}

	if rec.AvitoID == 0 && rec.AvitoPublishID == "" {
		return errors.New("вакансия еще не размещалась")
	}

	if rec.AvitoID == 0 {
		return errors.New("вакансия размещена, но еще не опубликованна")
	}

	if isEdit && rec.AvitoPublishID == "" {
		return errors.New("вакансия недоступна для редактирования")
	}
	return nil
}
