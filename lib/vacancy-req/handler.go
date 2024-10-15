package vacancyreqhandler

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	citystore "hr-tools-backend/lib/dicts/city/store"
	companyprovider "hr-tools-backend/lib/dicts/company"
	companystructprovider "hr-tools-backend/lib/dicts/company-struct"
	departmentprovider "hr-tools-backend/lib/dicts/department"
	jobtitleprovider "hr-tools-backend/lib/dicts/job-title"
	vacancyreqstore "hr-tools-backend/lib/vacancy-req/store"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(spaceID string, data vacancyapimodels.VacancyRequestData) (id string, err error)
	GetByID(spaceID, id string) (item vacancyapimodels.VacancyRequestView, err error)
	Update(spaceID, id string, data vacancyapimodels.VacancyRequestData) error
	Delete(spaceID, id string) error
	List(spaceID string) (list []vacancyapimodels.VacancyRequestView, err error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store:                 vacancyreqstore.NewInstance(db.DB),
		companyProvider:       companyprovider.Instance,
		departmentProvider:    departmentprovider.Instance,
		jobTitleProvider:      jobtitleprovider.Instance,
		cityStore:             citystore.NewInstance(db.DB),
		companyStructProvider: companystructprovider.Instance,
	}
}

type impl struct {
	store                 vacancyreqstore.Provider
	companyProvider       companyprovider.Provider
	departmentProvider    departmentprovider.Provider
	jobTitleProvider      jobtitleprovider.Provider
	cityStore             citystore.Provider
	companyStructProvider companystructprovider.Provider
}

func (i impl) checkDependency(spaceID string, data vacancyapimodels.VacancyRequestData) (err error) {
	if data.CompanyID != "" {
		_, err = i.companyProvider.Get(spaceID, data.CompanyID)
		if err != nil {
			return err
		}
	}
	if data.DepartmentID != "" {
		_, err = i.departmentProvider.Get(spaceID, data.DepartmentID)
		if err != nil {
			return err
		}
	}
	if data.JobTitleID != "" {
		_, err = i.jobTitleProvider.Get(spaceID, data.JobTitleID)
		if err != nil {
			return err
		}
	}
	if data.CityID != "" {
		cityRec, err := i.cityStore.GetByID(data.CityID)
		if err != nil {
			return err
		}
		if cityRec == nil {
			return errors.New("город не найден")
		}
	}
	if data.CompanyStructID != "" {
		_, err = i.companyStructProvider.Get(spaceID, data.CompanyStructID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i impl) Create(spaceID string, data vacancyapimodels.VacancyRequestData) (id string, err error) {
	logger := log.WithField("space_id", spaceID)
	err = i.checkDependency(spaceID, data)
	if err != nil {
		return "", err
	}
	rec := dbmodels.VacancyRequest{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		VacancyName:     data.VacancyName,
		Confidential:    data.Confidential,
		OpenedPositions: data.OpenedPositions,
		Urgency:         data.Urgency,
		RequestType:     data.RequestType,
		PlaceOfWork:     data.PlaceOfWork,
		ChiefFio:        data.ChiefFio,
		Interviewer:     data.Interviewer,
		ShortInfo:       data.ShortInfo,
		Requirements:    data.Requirements,
		Description:     data.Description,
		OutInteraction:  data.OutInteraction,
		InInteraction:   data.InInteraction,
	}
	if data.CompanyID != "" {
		rec.CompanyID = &data.CompanyID
	}
	if data.DepartmentID != "" {
		rec.DepartmentID = &data.DepartmentID
	}
	if data.JobTitleID != "" {
		rec.JobTitleID = &data.JobTitleID
	}
	if data.CityID != "" {
		rec.CityID = &data.CityID
	}
	if data.CompanyStructID != "" {
		rec.CompanyStructID = &data.CompanyStructID
	}
	recID, err := i.store.Create(rec)
	if err != nil {
		logger.
			WithField("request", fmt.Sprintf("%+v", data)).
			WithError(err).
			Error("Ошибка создания заявки")
		return "", err
	}
	logger.
		WithField("rec_id", recID).
		Info("Создана заявка")
	return recID, nil
}

func (i impl) GetByID(spaceID, id string) (item vacancyapimodels.VacancyRequestView, err error) {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения заявки")
		return vacancyapimodels.VacancyRequestView{}, err
	}
	if rec == nil {
		return vacancyapimodels.VacancyRequestView{}, errors.New("заявка не найдена")
	}
	return vacancyapimodels.VacancyRequestConvert(*rec), nil
}

func (i impl) Update(spaceID, id string, data vacancyapimodels.VacancyRequestData) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	err := i.checkDependency(spaceID, data)
	if err != nil {
		return err
	}
	updMap := map[string]interface{}{
		"SpaceID":         spaceID,
		"CompanyID":       data.CompanyID,
		"DepartmentID":    data.DepartmentID,
		"JobTitleID":      data.JobTitleID,
		"CityID":          data.CityID,
		"CompanyStructID": data.CompanyStructID,
		"VacancyName":     data.VacancyName,
		"Confidential":    data.Confidential,
		"OpenedPositions": data.OpenedPositions,
		"Urgency":         data.Urgency,
		"RequestType":     data.RequestType,
		"PlaceOfWork":     data.PlaceOfWork,
		"ChiefFio":        data.ChiefFio,
		"Interviewer":     data.Interviewer,
		"ShortInfo":       data.ShortInfo,
		"Requirements":    data.Requirements,
		"Description":     data.Description,
		"OutInteraction":  data.OutInteraction,
		"InInteraction":   data.InInteraction,
	}
	err = i.store.Update(spaceID, id, updMap)
	if err != nil {
		logger.
			WithField("request", fmt.Sprintf("%+v", data)).
			WithError(err).
			Error("ошибка обновления заявки")
		return err
	}
	logger.Info("обновлена заявка")
	return nil
}

func (i impl) Delete(spaceID, id string) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	err := i.store.Delete(spaceID, id)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка удаления заявки")
		return err
	}
	logger.Info("удалена заявки")
	return nil
}

func (i impl) List(spaceID string) (list []vacancyapimodels.VacancyRequestView, err error) {
	logger := log.WithField("space_id", spaceID)
	recList, err := i.store.List(spaceID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения списка заявок")
		return nil, err
	}
	result := make([]vacancyapimodels.VacancyRequestView, 0, len(list))
	for _, rec := range recList {
		result = append(result, vacancyapimodels.VacancyRequestConvert(rec))
	}
	return result, nil
}
