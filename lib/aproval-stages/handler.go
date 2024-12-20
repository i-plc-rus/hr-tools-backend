package aprovalstageshandler

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"hr-tools-backend/db"
	aprovalstagestore "hr-tools-backend/lib/aproval-stages/store"
	"hr-tools-backend/models"
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

func NewHandlerWithTx(tx *gorm.DB) Provider {
	return impl{
		store: aprovalstagestore.NewInstance(tx),
	}
}

type impl struct {
	store aprovalstagestore.Provider
}

func (i impl) Save(spaceID, vrID string, stages []vacancyapimodels.ApprovalStageData) (err error) {
	err = i.store.DeleteByVacancyRequest(spaceID, vrID)
	if err != nil {
		return err
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
		if rec.ApprovalStatus == "" {
			rec.ApprovalStatus = models.AStatusAwaiting
		}
		_, err = i.store.Create(rec)
		if err != nil {
			return errors.Wrapf(err, "Ошибка сохранения цепочки согласования, stage=%+v", stage)
		}
	}
	return nil
}
