package languagesprovider

import (
	"hr-tools-backend/db"
	languagestore "hr-tools-backend/lib/dicts/languages/store"
	dictapimodels "hr-tools-backend/models/api/dict"
)

type Provider interface {
	FindByName(name string) (list []dictapimodels.LangView, err error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store: languagestore.NewInstance(db.DB),
	}
}

type impl struct {
	store languagestore.Provider
}

func (i impl) FindByName(name string) (list []dictapimodels.LangView, err error) {
	recList, err := i.store.List(name)
	if err != nil {
		return nil, err
	}
	result := make([]dictapimodels.LangView, 0, len(list))
	for _, rec := range recList {
		result = append(result, dictapimodels.LangConvert(rec))
	}
	return result, nil
}
