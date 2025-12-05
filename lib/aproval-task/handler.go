package aprovaltaskhandler

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"hr-tools-backend/db"
	approvaltaskhistorystore "hr-tools-backend/lib/aproval-task/history-store"
	approvaltaskstore "hr-tools-backend/lib/aproval-task/store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	"hr-tools-backend/models"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Get(spaceID, requestID string) ([]vacancyapimodels.ApprovalTaskView, error)
	Save(spaceID, requestID string, stages []vacancyapimodels.ApprovalTaskData) (hMsg string, err error)
	History(spaceID, requestID string) ([]vacancyapimodels.ApprovalHistoryView, error)
	Audit(data dbmodels.ApprovalTask)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store:                approvaltaskstore.NewInstance(db.DB),
		spaceUsersStore:      spaceusersstore.NewInstance(db.DB),
		approvalHistoryStore: approvaltaskhistorystore.NewInstance(db.DB),
	}
}

func NewHandlerWithTx(tx *gorm.DB) Provider {
	return impl{
		approvalHistoryStore: approvaltaskhistorystore.NewInstance(tx),
		store:                approvaltaskstore.NewInstance(tx),
		spaceUsersStore:      spaceusersstore.NewInstance(tx),
	}
}

type impl struct {
	store                approvaltaskstore.Provider
	spaceUsersStore      spaceusersstore.Provider
	approvalHistoryStore approvaltaskhistorystore.Provider
}

func (i impl) GetLogger(spaceID, requestID string) *log.Entry {
	logger := log.
		WithField("space_id", spaceID).
		WithField("vacancy_request_id", requestID)
	return logger
}

func (i impl) Get(spaceID, requestID string) ([]vacancyapimodels.ApprovalTaskView, error) {
	currentList, err := i.store.List(spaceID, requestID)
	if err != nil {
		return nil, err
	}
	result := make([]vacancyapimodels.ApprovalTaskView, 0, len(currentList))
	for _, rec := range currentList {
		result = append(result, vacancyapimodels.ApprovalStageConvert(rec))
	}
	return result, nil
}

func (i impl) Save(spaceID, requestID string, stages []vacancyapimodels.ApprovalTaskData) (hMsg string, err error) {
	usersMap := map[string]int{}                     //0-оставить/1-добавить/-1 удалить
	currentMap := map[string]dbmodels.ApprovalTask{} //[userid]rec
	currentList, err := i.store.List(spaceID, requestID)
	if err != nil {
		return "", err
	}
	for _, current := range currentList {
		usersMap[current.AssigneeUserID] = -1
		currentMap[current.AssigneeUserID] = current
	}

	for _, stage := range stages {
		user, err := i.spaceUsersStore.GetByID(stage.AssigneeUserID)
		if err != nil {
			return "", err
		}
		if user == nil || user.SpaceID != spaceID {
			return fmt.Sprintf("Сотрудник %v, не найден в справочнике сотрудников", stage.AssigneeUserID), nil
		}

		what, ok := usersMap[stage.AssigneeUserID]
		if ok {
			if what < 0 {
				usersMap[stage.AssigneeUserID] = 0
			} else {
				return fmt.Sprintf("Сотрудник %v уже был указан на ранних этапах", user.GetFullName()), nil
			}
		} else {
			usersMap[stage.AssigneeUserID] = 1
		}
	}
	for userID, what := range usersMap {
		switch what {
		case -1: // удаляем отсутсвующих
			currentRec, ok := currentMap[userID]
			if ok {
				err = i.store.Delete(spaceID, currentRec.ID)
				if err != nil {
					return "", err
				}
				currentRec.State = models.AStateRemoved
				i.Audit(currentRec)
			}
		case 1: // добавляем новых
			rec := dbmodels.ApprovalTask{
				BaseSpaceModel: dbmodels.BaseSpaceModel{
					SpaceID: spaceID,
				},
				RequestID:      requestID,
				AssigneeUserID: userID,
				State:          models.AStatePending,
			}
			recID, err := i.store.Create(rec)
			if err != nil {
				return "", errors.Wrapf(err, "Ошибка сохранения цепочки согласования, stage=%+v", rec)
			}
			rec.ID = recID
			i.Audit(rec)
		}
	}
	return "", nil
}

func (i impl) History(spaceID, requestID string) ([]vacancyapimodels.ApprovalHistoryView, error) {
	list, err := i.approvalHistoryStore.List(spaceID, requestID)
	if err != nil {
		return nil, err
	}
	result := make([]vacancyapimodels.ApprovalHistoryView, 0, len(list))
	for _, rec := range list {
		result = append(result, vacancyapimodels.ApprovalHistoryConvert(rec))
	}
	return result, nil
}

func (i impl) Audit(data dbmodels.ApprovalTask) {
	rec := dbmodels.ApprovalHistory{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: data.SpaceID,
		},
		RequestID:      data.RequestID,
		TaskID:         data.ID,
		AssigneeUserID: data.AssigneeUserID,
		State:          data.State,
		Comment:        data.Comment,
	}
	_, err := i.approvalHistoryStore.Create(rec)
	if err != nil {
		i.GetLogger(data.SpaceID, data.RequestID).WithError(err).Error("Ошибка добавления истории по задаче согласования заявки")
	}
}
