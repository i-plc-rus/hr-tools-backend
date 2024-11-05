package cityprovider

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	store "hr-tools-backend/lib/dicts/city/store"
	dictapimodels "hr-tools-backend/models/api/dict"
)

type Provider interface {
	Get(id string) (item dictapimodels.CityView, err error)
	FindByName(request dictapimodels.CityData) (list []dictapimodels.CityView, err error)
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

func (i impl) Get(id string) (item dictapimodels.CityView, err error) {
	logger := log.WithField("rec_id", id)
	rec, err := i.store.GetByID(id)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения города")
		return dictapimodels.CityView{}, err
	}
	if rec == nil {
		return dictapimodels.CityView{}, errors.New("город не найдена")
	}
	return dictapimodels.CityConvert(*rec), nil
}

func (i impl) FindByName(request dictapimodels.CityData) (list []dictapimodels.CityView, err error) {
	recList, err := i.store.List(request.Address)
	if err != nil {
		log.
			WithError(err).
			Error("ошибка получения списка городов")
		return nil, err
	}
	result := make([]dictapimodels.CityView, 0, len(list))
	for _, rec := range recList {
		result = append(result, dictapimodels.CityConvert(rec))
	}
	return result, nil
}
