package aprovalstageshandler

import (
	"fmt"
	"hr-tools-backend/db"
	aprovalstagestore "hr-tools-backend/lib/aproval-stages/store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	"hr-tools-backend/models"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Provider interface {
	Save(spaceID, vacancyRequestID string, stages []vacancyapimodels.ApprovalStageData) (hMsg string, err error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store:           aprovalstagestore.NewInstance(db.DB),
		spaceUsersStore: spaceusersstore.NewInstance(db.DB),
	}
}

func NewHandlerWithTx(tx *gorm.DB) Provider {
	return impl{
		store:           aprovalstagestore.NewInstance(tx),
		spaceUsersStore: spaceusersstore.NewInstance(tx),
	}
}

type impl struct {
	store           aprovalstagestore.Provider
	spaceUsersStore spaceusersstore.Provider
}

func (i impl) Save(spaceID, vrID string, stages []vacancyapimodels.ApprovalStageData) (hMsg string, err error) {
	usersMap := map[string]bool{}
	for _, stage := range stages {
		user, err := i.spaceUsersStore.GetByID(stage.SpaceUserID)
		if err != nil {
			return "", err
		}
		if user == nil || user.SpaceID != spaceID {
			return fmt.Sprintf("сотрудник с этапа %v, не найден в справочнике сотрудников", stage.Stage), nil
		}
		if usersMap[stage.SpaceUserID] {
			return fmt.Sprintf("сотрудник с этапа %v уже был указан на ранних этапах", stage.Stage), nil
		}
		usersMap[stage.SpaceUserID] = true
	}
	err = i.store.DeleteByVacancyRequest(spaceID, vrID)
	if err != nil {
		return "", err
	}

	if len(stages) == 0 {
		return "", nil
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
			return "", errors.Wrapf(err, "Ошибка сохранения цепочки согласования, stage=%+v", stage)
		}
	}
	return "", nil
}
