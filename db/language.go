package db

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	languagestore "hr-tools-backend/lib/dicts/languages/store"
	dbmodels "hr-tools-backend/models/db"
	"os"
)

func fillLanguages() {
	log.Info("предзаполнение языков")
	store := languagestore.NewInstance(DB)
	list, err := store.List("")
	if err != nil {
		log.WithError(err).Error("ошибка предзаполнения языков")
		return
	}
	if len(list) > 0 {
		log.Info("языки заполнены")
		return
	}

	body, err := os.ReadFile("./static_preload/languages.json")
	if err != nil {
		log.WithError(err).Error("ошибка чтения файла с языками")
		return
	}
	type langStruct struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Uid  string `json:"uid"`
	}
	lines := []langStruct{}
	err = json.Unmarshal(body, &lines)
	if err != nil {
		log.WithError(err).Error("ошибка сериализации файла с языком")
		return
	}

	for _, item := range lines {

		rec := dbmodels.LanguageData{
			BaseModel: dbmodels.BaseModel{
				ID: item.Uid,
			},
			Code: item.ID,
			Name: item.Name,
		}
		err = store.Add(rec, true)
		if err != nil {
			log.
				WithError(err).
				Error("ошибка добавления языка")
			return
		}
	}

	log.Info("языки добавлены")
}
