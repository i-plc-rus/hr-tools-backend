package companyprovider

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	"hr-tools-backend/lib/dicts/company/store"
	initchecker "hr-tools-backend/lib/utils/init-checker"
	dictapimodels "hr-tools-backend/models/api/dict"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(spaceID string, request dictapimodels.CompanyData) (id string, err error)
	Update(spaceID, id string, request dictapimodels.CompanyData) error
	Get(spaceID, id string) (item dictapimodels.CompanyView, err error)
	FindByName(spaceID string, request dictapimodels.CompanyData) (list []dictapimodels.CompanyView, err error)
	Delete(spaceID, id string) error
}

var Instance Provider

func NewHandler() {
	instance := impl{
		store: store.NewInstance(db.DB),
	}
	initchecker.CheckInit(
		"store", instance.store,
	)
	Instance = instance
}

type impl struct {
	store store.Provider
}

func (i impl) Create(spaceID string, request dictapimodels.CompanyData) (id string, err error) {
	logger := log.WithField("space_id", spaceID)
	rec := dbmodels.Company{
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
		WithField("company_name", rec.Name).
		WithField("rec_id", rec.ID).
		Info("создана компания")
	return id, nil
}

func (i impl) Update(spaceID, id string, request dictapimodels.CompanyData) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	updMap := map[string]interface{}{
		"name": request.Name,
	}
	err := i.store.Update(spaceID, id, updMap)
	if err != nil {
		return err
	}
	logger.Info("обновлена компания")
	return nil
}

func (i impl) Get(spaceID, id string) (item dictapimodels.CompanyView, err error) {
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return dictapimodels.CompanyView{}, err
	}
	if rec == nil {
		return dictapimodels.CompanyView{}, errors.New("компания не найдена")
	}
	return dictapimodels.CompanyConvert(*rec), nil
}

func (i impl) FindByName(spaceID string, request dictapimodels.CompanyData) (list []dictapimodels.CompanyView, err error) {
	recList, err := i.store.FindByName(spaceID, request.Name)
	if err != nil {
		return nil, err
	}
	result := make([]dictapimodels.CompanyView, 0, len(list))
	for _, rec := range recList {
		result = append(result, dictapimodels.CompanyConvert(rec))
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
	logger.Info("удалена компания")
	return nil
}
