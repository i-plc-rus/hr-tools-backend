package jobtitleprovider

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	"hr-tools-backend/lib/dicts/job-title/store"
	dictapimodels "hr-tools-backend/models/api/dict"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(spaceID string, request dictapimodels.JobTitleData) (id string, err error)
	Update(spaceID, id string, request dictapimodels.JobTitleData) error
	Get(spaceID, id string) (item dictapimodels.JobTitleView, err error)
	FindByName(spaceID string, request dictapimodels.JobTitleData) (list []dictapimodels.JobTitleView, err error)
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

func (i impl) Create(spaceID string, request dictapimodels.JobTitleData) (id string, err error) {
	logger := log.WithField("space_id", spaceID)
	rec := dbmodels.JobTitle{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		DepartmentID: request.DepartmentID,
		Name:         request.Name,
	}
	id, err = i.store.Create(rec)
	if err != nil {
		logger.
			WithField("request", fmt.Sprintf("%+v", request)).
			WithError(err).
			Error("ошибка создания штатной должности")
		return "", err
	}
	logger.
		WithField("job_title_name", rec.Name).
		WithField("rec_id", rec.ID).
		Info("создана штатная должность")
	return id, nil
}

func (i impl) Update(spaceID, id string, request dictapimodels.JobTitleData) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	updMap := map[string]interface{}{
		"name": request.Name,
	}
	err := i.store.Update(spaceID, id, updMap)
	if err != nil {
		logger.
			WithField("request", fmt.Sprintf("%+v", request)).
			WithError(err).
			Error("ошибка обновления штатной должности")
		return err
	}
	logger.Info("обновлена штатная должность")
	return nil
}

func (i impl) Get(spaceID, id string) (item dictapimodels.JobTitleView, err error) {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения штатной должности")
		return dictapimodels.JobTitleView{}, err
	}
	if rec == nil {
		return dictapimodels.JobTitleView{}, errors.New("штатная должность не найдена")
	}
	return dictapimodels.JobTitleConvert(*rec), nil
}

func (i impl) FindByName(spaceID string, request dictapimodels.JobTitleData) (list []dictapimodels.JobTitleView, err error) {
	logger := log.WithField("space_id", spaceID)
	recList, err := i.store.FindByName(spaceID, request.Name, request.DepartmentID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения списка штатных должностей")
		return nil, err
	}
	result := make([]dictapimodels.JobTitleView, 0, len(list))
	for _, rec := range recList {
		result = append(result, dictapimodels.JobTitleConvert(rec))
	}
	return result, nil
}

func (i impl) Delete(spaceID, id string) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	err := i.store.Delete(spaceID, id)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка удаления штатной должности")
		return err
	}
	logger.Info("удалена штатная должность")
	return nil
}
