package companystructprovider

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	"hr-tools-backend/lib/dicts/company-struct/store"
	dictapimodels "hr-tools-backend/models/api/dict"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(spaceID string, request dictapimodels.CompanyStructData) (id string, err error)
	Update(spaceID, id string, request dictapimodels.CompanyStructData) error
	Get(spaceID, id string) (item dictapimodels.CompanyStructView, err error)
	FindByName(spaceID string, request dictapimodels.CompanyStructData) (list []dictapimodels.CompanyStructView, err error)
	Delete(spaceID, id string) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store: store.NewInstance(db.DB),
	}
}

type impl struct {
	store store.Provider
}

func (i impl) Create(spaceID string, request dictapimodels.CompanyStructData) (id string, err error) {
	logger := log.WithField("space_id", spaceID)
	rec := dbmodels.CompanyStruct{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		Name: request.Name,
	}
	id, err = i.store.Create(rec)
	if err != nil {
		return "", err
	}
	logger.
		WithField("company_struct_name", rec.Name).
		WithField("rec_id", rec.ID).
		Info("создана структура компании")
	return id, nil
}

func (i impl) Update(spaceID, id string, request dictapimodels.CompanyStructData) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	updMap := map[string]interface{}{
		"name": request.Name,
	}
	err := i.store.Update(spaceID, id, updMap)
	if err != nil {
		return err
	}
	logger.Info("обновлена структура компании")
	return nil
}

func (i impl) Get(spaceID, id string) (item dictapimodels.CompanyStructView, err error) {
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return dictapimodels.CompanyStructView{}, err
	}
	if rec == nil {
		return dictapimodels.CompanyStructView{}, errors.New("структура компании не найдена")
	}
	return dictapimodels.CompanyStructConvert(*rec), nil
}

func (i impl) FindByName(spaceID string, request dictapimodels.CompanyStructData) (list []dictapimodels.CompanyStructView, err error) {
	recList, err := i.store.FindByName(spaceID, request.Name)
	if err != nil {
		return nil, err
	}
	result := make([]dictapimodels.CompanyStructView, 0, len(list))
	for _, rec := range recList {
		result = append(result, dictapimodels.CompanyStructConvert(rec))
	}
	return result, nil
}

func (i impl) Delete(spaceID, id string) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	err := i.store.Delete(spaceID, id)
	if err != nil {
		return err
	}
	logger.Info("удалена структура компании")
	return nil
}
