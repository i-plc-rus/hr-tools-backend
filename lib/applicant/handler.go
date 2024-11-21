package applicant

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	selectionstagestore "hr-tools-backend/lib/vacancy/selection-stage-store"
	"hr-tools-backend/models"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

type Provider interface {
	ListOfNegotiation(spaceID string, filter dbmodels.NegotiationFilter) (list []negotiationapimodels.NegotiationView, err error)
	UpdateComment(id string, comment string) error
	UpdateStatus(spaceID, id string, status models.NegotiationStatus) error
	GetByID(spaceID, id string) (negotiationapimodels.NegotiationView, error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store:               applicantstore.NewInstance(db.DB),
		selectionStageStore: selectionstagestore.NewInstance(db.DB),
	}
}

type impl struct {
	store               applicantstore.Provider
	selectionStageStore selectionstagestore.Provider
}

func (i impl) ListOfNegotiation(spaceID string, filter dbmodels.NegotiationFilter) ([]negotiationapimodels.NegotiationView, error) {
	list, err := i.store.ListOfNegotiation(spaceID, filter)
	if err != nil {
		return nil, err
	}
	result := make([]negotiationapimodels.NegotiationView, 0, len(list))
	for _, rec := range list {
		result = append(result, negotiationapimodels.NegotiationConvert(rec))
	}
	return result, nil
}

func (i impl) UpdateComment(id string, comment string) error {
	updMap := map[string]interface{}{
		"comment": comment,
	}
	return i.store.Update(id, updMap)
}

func (i impl) UpdateStatus(spaceID, id string, status models.NegotiationStatus) error {
	logger := log.
		WithField("space_id", spaceID).
		WithField("applicant_id", id)
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("запись не найдена")
	}
	ok, err := rec.IsAllowStatusChange(status)
	if err != nil || !ok {
		return err
	}
	updMap := map[string]interface{}{
		"negotiation_status":      status,
		"negotiation_accept_date": nil,
	}
	if status == models.NegotiationStatusAccepted {
		updMap["negotiation_accept_date"] = time.Now()
		updMap["status"] = models.ApplicantStatusInProcess
		selectionStages, err := i.selectionStageStore.List(rec.SpaceID, rec.VacancyID)
		if err != nil {
			logger.WithError(err).Error("ошибка получения списка этапов подбора")
			return errors.New("ошибка получения списка этапов подбора")
		}
		for _, stage := range selectionStages {
			if stage.Name == dbmodels.AddedStage {
				updMap["selection_stage_id"] = stage.ID
				break
			}
		}
	}
	if status == models.NegotiationStatusRejected {
		updMap["negotiation_accept_date"] = time.Now()
	}
	err = i.store.Update(id, updMap)
	if err != nil {
		logger.WithError(err).Error("ошибка обновления кандидата")
		return errors.New("ошибка обновления кандидата")
	}
	return nil
}

func (i impl) GetByID(spaceID, id string) (negotiationapimodels.NegotiationView, error) {
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return negotiationapimodels.NegotiationView{}, err
	}
	if rec == nil {
		return negotiationapimodels.NegotiationView{}, errors.New("отклил не найден")
	}
	return negotiationapimodels.NegotiationConvertExt(*rec), nil
}
