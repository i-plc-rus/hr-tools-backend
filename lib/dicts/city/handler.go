package cityprovider

import (
	"github.com/pkg/errors"
	"hr-tools-backend/db"
	store "hr-tools-backend/lib/dicts/city/store"
	initchecker "hr-tools-backend/lib/utils/init-checker"
	dictapimodels "hr-tools-backend/models/api/dict"
)

type Provider interface {
	Get(id string) (item dictapimodels.CityView, err error)
	FindByName(request dictapimodels.CityData) (list []dictapimodels.CityView, err error)
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

func (i impl) Get(id string) (item dictapimodels.CityView, err error) {
	rec, err := i.store.GetByID(id)
	if err != nil {
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
		return nil, err
	}
	result := make([]dictapimodels.CityView, 0, len(list))
	for _, rec := range recList {
		result = append(result, dictapimodels.CityConvert(rec))
	}
	return result, nil
}
