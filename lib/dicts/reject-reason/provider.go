package rejectreasonprovider

import (
	"hr-tools-backend/db"
	rejectreasondictstore "hr-tools-backend/lib/dicts/reject-reason/store"
	"hr-tools-backend/models"
	dictapimodels "hr-tools-backend/models/api/dict"
	dbmodels "hr-tools-backend/models/db"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var hrReasons = []string{
	"Не дозвонились",
	"Отказ образование",
	"Отказ переезд",
	"Отказ опыт работы",
	"Отказ график работы",
	"Отказ пол",
	"Отказ гражданство",
	"Отказ возраст",
	"Отказ на работном сайте",
	"Плохо выполненное тестовое задание",
	"Недостаток мотивации",
	"Отсутствие качеств, необходимых для позиции / компании",
	"Недостаток опыта",
}

var headReasons = []string{
	"Отсутствие качеств, необходимых для позиции / компании",
	"Плохо выполненное тестовое задание",
	"Недостаток мотивации",
	"Недостаток опыта",
}

var applicantReasons = []string{
	"Плохое впечатление от менеджемента",
	"Контроффер",
	"Неинтересна компания/сфера",
	"Неинтересные задачи/обязанности",
}

type Provider interface {
	Create(spaceID string, request dictapimodels.RejectReasonData) (id string, hMsg string, err error)
	Update(spaceID, id string, request dictapimodels.RejectReasonData) (hMsg string, err error)
	Get(spaceID, id string) (item dictapimodels.RejectReasonView, err error)
	List(spaceID string, filter dictapimodels.RejectReasonFind) (list []dictapimodels.RejectReasonView, err error)
	Delete(spaceID, id string) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store: rejectreasondictstore.NewInstance(db.DB),
	}
}

type impl struct {
	store rejectreasondictstore.Provider
}

func (i impl) Create(spaceID string, request dictapimodels.RejectReasonData) (id string, hMsg string, err error) {
	logger := log.WithField("space_id", spaceID)
	rec := dbmodels.RejectReason{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		Initiator: request.Initiator,
		Name:      request.Name,
	}
	found, err := i.store.IsUnique(rec.SpaceID, "", rec.Name, rec.Initiator)
	if err != nil {
		return "", "", err
	}
	if found {
		return "", "причина отказа уже существует", nil
	}

	id, err = i.store.Create(rec)
	if err != nil {
		return "", "", err
	}
	logger.
		WithField("job_title_name", rec.Name).
		WithField("rec_id", rec.ID).
		Info("создана штатная должность")
	return id, "", nil
}

func (i impl) Update(spaceID, id string, request dictapimodels.RejectReasonData) (hMsg string, err error) {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	updMap := map[string]interface{}{
		"name": request.Name,
	}
	found, err := i.store.IsUnique(spaceID, id, request.Name, request.Initiator)
	if err != nil {
		return "", err
	}
	if found {
		return "причина отказа уже существует", nil
	}
	err = i.store.Update(spaceID, id, updMap)
	if err != nil {
		return "", err
	}
	logger.Info("обновлена штатная должность")
	return "", nil
}

func (i impl) Get(spaceID, id string) (item dictapimodels.RejectReasonView, err error) {
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return dictapimodels.RejectReasonView{}, err
	}
	if rec == nil {
		return dictapimodels.RejectReasonView{}, errors.New("штатная должность не найдена")
	}
	return dictapimodels.RejectReasonConvert(*rec), nil
}

func (i impl) List(spaceID string, filter dictapimodels.RejectReasonFind) (list []dictapimodels.RejectReasonView, err error) {
	recList, err := i.store.List(spaceID, filter)
	if err != nil {
		return nil, err
	}
	result := make([]dictapimodels.RejectReasonView, 0, len(list)+len(hrReasons)+len(headReasons)+len(applicantReasons))
	result = append(result, getStatic(filter)...)
	for _, rec := range recList {
		result = append(result, dictapimodels.RejectReasonConvert(rec))
	}
	return result, nil
}

func (i impl) Delete(spaceID, id string) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	err := i.store.Delete(spaceID, id)
	if err != nil {
		return err
	}
	logger.Info("удалена штатная должность")
	return nil
}

func getStatic(filter dictapimodels.RejectReasonFind) []dictapimodels.RejectReasonView {
	result := getStaticView(models.HrReject, hrReasons, filter)
	result = append(result, getStaticView(models.HeadReject, headReasons, filter)...)
	result = append(result, getStaticView(models.ApplicantReject, applicantReasons, filter)...)
	return result
}

func getStaticView(initiator models.RejectInitiator, reasons []string, filter dictapimodels.RejectReasonFind) []dictapimodels.RejectReasonView {
	result := make([]dictapimodels.RejectReasonView, 0, len(reasons))
	if filter.Initiator != "" && initiator != filter.Initiator {
		return []dictapimodels.RejectReasonView{}
	}
	search := strings.ToLower(filter.Search)
	for _, name := range reasons {
		if search != "" {
			if !strings.Contains(strings.ToLower(name), search) {
				continue
			}
		}
		result = append(result, dictapimodels.RejectReasonView{
			RejectReasonData: dictapimodels.RejectReasonData{
				Name:      name,
				Initiator: initiator,
			},
			ID:        "",
			CanChange: false,
		})
	}
	return result
}
