package aprovalstageshandler

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	aprovalstagestore "hr-tools-backend/lib/aproval-stages/store"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Save(spaceID, vacancyRequestID string, stages []vacancyapimodels.ApprovalStageData) (err error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store: aprovalstagestore.NewInstance(db.DB),
	}
}

type impl struct {
	store aprovalstagestore.Provider
}

func (i impl) Save(spaceID, vrID string, stages []vacancyapimodels.ApprovalStageData) (err error) {
	logger := log.WithField("space_id", spaceID).
		WithField("vacancy_request_id", vrID)
	err = i.store.DeleteByVacancyRequest(spaceID, vrID)
	if err != nil {
		logger.
			WithError(err).
			Error("Ошибка сохранения цепочки согласования")
		return errors.Wrap(err, "Ошибка сохранения цепочки согласования")
	}

	if len(stages) == 0 {
		return nil
	}
	for _, stage := range stages {
		rec := dbmodels.ApprovalStage{
			BaseSpaceModel: dbmodels.BaseSpaceModel{
				SpaceID: spaceID,
			},
			VacancyRequestID: vrID,
			Stage:            stage.Stage,
			SpaceUserID:      stage.SpaceUserID,
			ApprovalStatus:   stage.ApprovalStatus,
		}
		_, err = i.store.Create(rec)
		if err != nil {
			logger.
				WithField("stage", fmt.Sprintf("%+v", stage)).
				WithError(err).
				Error("Ошибка сохранения цепочки согласования")
			return errors.Wrap(err, "Ошибка сохранения цепочки согласования")
		}
	}
	return nil
}
