package vacancyreqhandler

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"hr-tools-backend/db"
	aprovalstageshandler "hr-tools-backend/lib/aproval-stages"
	approvalstagestore "hr-tools-backend/lib/aproval-stages/store"
	citystore "hr-tools-backend/lib/dicts/city/store"
	companyprovider "hr-tools-backend/lib/dicts/company"
	companystructprovider "hr-tools-backend/lib/dicts/company-struct"
	departmentprovider "hr-tools-backend/lib/dicts/department"
	jobtitleprovider "hr-tools-backend/lib/dicts/job-title"
	vacancyhandler "hr-tools-backend/lib/vacancy"
	vacancyreqstore "hr-tools-backend/lib/vacancy-req/store"
	"hr-tools-backend/models"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(spaceID, userID string, data vacancyapimodels.VacancyRequestCreateData) (id string, err error)
	GetByID(spaceID, id string) (item vacancyapimodels.VacancyRequestView, err error)
	Update(spaceID, id string, data vacancyapimodels.VacancyRequestEditData) error
	Delete(spaceID, id string) error
	List(spaceID, userID string, filter vacancyapimodels.VrFilter) (list []vacancyapimodels.VacancyRequestView, rowCount int64, err error)
	ChangeStatus(spaceID, id, userID string, status models.VRStatus) error
	Approve(spaceID, id, userID string, data vacancyapimodels.VacancyRequestData) error
	Reject(spaceID, id, userID string, data vacancyapimodels.VacancyRequestData) error
	CreateVacancy(spaceID, id, userID string) error
	ToPin(id, userID string, isSet bool) error
	ToFavorite(id, userID string, isSet bool) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store:                 vacancyreqstore.NewInstance(db.DB),
		approvalStageStore:    approvalstagestore.NewInstance(db.DB),
		companyProvider:       companyprovider.Instance,
		departmentProvider:    departmentprovider.Instance,
		jobTitleProvider:      jobtitleprovider.Instance,
		cityStore:             citystore.NewInstance(db.DB),
		companyStructProvider: companystructprovider.Instance,
		vacancyHandler:        vacancyhandler.Instance,
		aprovalStagesHandler:  aprovalstageshandler.Instance,
	}
}

type impl struct {
	store                 vacancyreqstore.Provider
	approvalStageStore    approvalstagestore.Provider
	companyProvider       companyprovider.Provider
	departmentProvider    departmentprovider.Provider
	jobTitleProvider      jobtitleprovider.Provider
	cityStore             citystore.Provider
	companyStructProvider companystructprovider.Provider
	vacancyHandler        vacancyhandler.Provider
	aprovalStagesHandler  aprovalstageshandler.Provider
}

func (i impl) checkDependency(spaceID string, data vacancyapimodels.VacancyRequestData) (err error) {
	if data.CompanyID != "" {
		_, err = i.companyProvider.Get(spaceID, data.CompanyID)
		if err != nil {
			return err
		}
	}
	if data.DepartmentID != "" {
		_, err = i.departmentProvider.Get(spaceID, data.DepartmentID)
		if err != nil {
			return err
		}
	}
	if data.JobTitleID != "" {
		_, err = i.jobTitleProvider.Get(spaceID, data.JobTitleID)
		if err != nil {
			return err
		}
	}
	if data.CityID != "" {
		cityRec, err := i.cityStore.GetByID(data.CityID)
		if err != nil {
			return err
		}
		if cityRec == nil {
			return errors.New("город не найден")
		}
	}
	if data.CompanyStructID != "" {
		_, err = i.companyStructProvider.Get(spaceID, data.CompanyStructID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i impl) Create(spaceID, userID string, data vacancyapimodels.VacancyRequestCreateData) (id string, err error) {
	logger := log.WithField("space_id", spaceID)
	err = i.checkDependency(spaceID, data.VacancyRequestData)
	if err != nil {
		return "", err
	}
	rec := dbmodels.VacancyRequest{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		AuthorID:        userID,
		VacancyName:     data.VacancyName,
		Confidential:    data.Confidential,
		OpenedPositions: data.OpenedPositions,
		Urgency:         data.Urgency,
		RequestType:     data.RequestType,
		SelectionType:   data.SelectionType,
		PlaceOfWork:     data.PlaceOfWork,
		ChiefFio:        data.ChiefFio,
		Interviewer:     data.Interviewer,
		ShortInfo:       data.ShortInfo,
		Requirements:    data.Requirements,
		OutInteraction:  data.OutInteraction,
		InInteraction:   data.InInteraction,
		Status:          models.VRStatusCreated,
		Employment:      data.Employment,
		Experience:      data.Experience,
		Schedule:        data.Schedule,
	}
	if data.AsTemplate {
		rec.Status = models.VRStatusTemplate
	}
	if data.CompanyID != "" {
		rec.CompanyID = &data.CompanyID
	}
	if data.DepartmentID != "" {
		rec.DepartmentID = &data.DepartmentID
	}
	if data.JobTitleID != "" {
		rec.JobTitleID = &data.JobTitleID
	}
	if data.CityID != "" {
		rec.CityID = &data.CityID
	}
	if data.CompanyStructID != "" {
		rec.CompanyStructID = &data.CompanyStructID
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		store := vacancyreqstore.NewInstance(tx)
		aprovalStagesHandler := aprovalstageshandler.NewHandlerWithTx(tx)
		id, err = store.Create(rec)
		if err != nil {
			logger.
				WithField("request", fmt.Sprintf("%+v", data)).
				WithError(err).
				Error("Ошибка создания заявки")
			return err
		}
		return aprovalStagesHandler.Save(spaceID, id, data.ApprovalStages.ApprovalStages)
	})
	if err != nil {
		return "", err
	}
	logger.
		WithField("rec_id", id).
		Info("Создана заявка")
	return id, nil
}

func (i impl) GetByID(spaceID, id string) (item vacancyapimodels.VacancyRequestView, err error) {
	rec, err := i.getRec(spaceID, id)
	if err != nil {
		return vacancyapimodels.VacancyRequestView{}, err
	}
	return vacancyapimodels.VacancyRequestConvert(*rec), nil
}

func (i impl) Update(spaceID, id string, data vacancyapimodels.VacancyRequestEditData) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		store := vacancyreqstore.NewInstance(tx)
		aprovalStagesHandler := aprovalstageshandler.NewHandlerWithTx(tx)
		err := i.updateVr(store, spaceID, id, data.VacancyRequestData)
		if err != nil {
			return err
		}
		return aprovalStagesHandler.Save(spaceID, id, data.ApprovalStages.ApprovalStages)
	})
	if err != nil {
		return err
	}
	logger.Info("обновлена заявка")
	return nil
}

func (i impl) Delete(spaceID, id string) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	err := i.store.Delete(spaceID, id)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка удаления заявки")
		return err
	}
	logger.Info("удалена заявки")
	return nil
}

func (i impl) List(spaceID, userID string, filter vacancyapimodels.VrFilter) (list []vacancyapimodels.VacancyRequestView, rowCount int64, err error) {
	logger := log.WithField("space_id", spaceID)
	rowCount, err = i.store.ListCount(spaceID, userID, filter)
	if err != nil {
		return nil, 0, err
	}

	page, limit := filter.GetPage()
	offset := (page - 1) * limit
	if int64(offset) > rowCount {
		return []vacancyapimodels.VacancyRequestView{}, rowCount, nil
	}

	recList, err := i.store.List(spaceID, userID, filter)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения списка заявок")
		return nil, 0, err
	}
	result := make([]vacancyapimodels.VacancyRequestView, 0, len(list))
	for _, rec := range recList {
		result = append(result, vacancyapimodels.VacancyRequestConvert(rec))
	}
	return result, rowCount, nil
}

func (i impl) ChangeStatus(spaceID, id, userID string, status models.VRStatus) error {
	logger := log.
		WithField("space_id", spaceID).
		WithField("rec_id", id).
		WithField("new_status", status)
	rec, err := i.GetByID(spaceID, id)
	if err != nil {
		return err
	}
	if !rec.Status.IsAllowChange(status) {
		return errors.Errorf("изменение статуса на %v недопустимо", status)
	}
	updMap := map[string]interface{}{
		"status": status,
	}
	err = i.store.Update(spaceID, id, updMap)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка обновления статуса")
		return err
	}
	logger.Info("статус заявки обновлен")
	return nil
}

func (i impl) checkVacancyExist(spaceID, id, userID string) (bool, error) {
	filter := vacancyapimodels.VacancyFilter{
		VacancyRequestID: id,
	}
	_, rowCount, err := i.vacancyHandler.List(spaceID, userID, filter)
	if err != nil {
		return false, err
	}
	return rowCount > 0, nil
}

func (i impl) Approve(spaceID, id, userID string, data vacancyapimodels.VacancyRequestData) error {
	logger := log.
		WithField("space_id", spaceID).
		WithField("rec_id", id).
		WithField("user_id", userID)
	rec, err := i.getRec(spaceID, id)
	if err != nil {
		return err
	}
	if !rec.Status.AllowAccept() {
		return errors.Errorf("невозможно согласовать заявку в текущем статусе: %v", rec.Status)
	}
	if rec.Status == models.VRStatusAccepted {
		return errors.New("заявка уже согласована")
	}

	isLastStage, stage := rec.GetCurrentApprovalStage()
	if stage != nil {
		if userID != stage.SpaceUserID {
			return errors.New("за текущий этап отвечает другой сотрудник")
		}
		err = i.updateVr(i.store, spaceID, id, data)
		if err != nil {
			logger.WithError(err).Error("ошибка обновления данных заявки при согласовании")
			return err
		}
		updMap := map[string]interface{}{
			"ApprovalStatus": models.AStatusApproved,
		}
		err = i.approvalStageStore.Update(spaceID, stage.ID, updMap)
		if err != nil {
			logger.WithError(err).Error("ошибка обновления статуса согласования")
			return err
		}
	}
	if isLastStage {
		err = i.ChangeStatus(spaceID, id, userID, models.VRStatusAccepted)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i impl) Reject(spaceID, id, userID string, data vacancyapimodels.VacancyRequestData) error {
	logger := log.
		WithField("space_id", spaceID).
		WithField("rec_id", id).
		WithField("user_id", userID)
	rec, err := i.getRec(spaceID, id)
	if err != nil {
		return err
	}
	if !rec.Status.AllowReject() {
		return errors.Errorf("невозможно согласовать заявку в текущем статусе: %v", rec.Status)
	}
	_, stage := rec.GetCurrentApprovalStage()
	if stage != nil {
		if userID != stage.SpaceUserID {
			return errors.New("за текущий этап отвечает другой сотрудник")
		}
		err = i.updateVr(i.store, spaceID, id, data)
		if err != nil {
			logger.WithError(err).Error("ошибка обновления данных заявки при согласовании")
			return err
		}
		updMap := map[string]interface{}{
			"ApprovalStatus": models.AStatusRejected,
		}
		err = i.approvalStageStore.Update(spaceID, stage.ID, updMap)
		if err != nil {
			logger.WithError(err).Error("ошибка обновления статуса согласования")
			return err
		}
	}
	return nil
}

func (i impl) getRec(spaceID, id string) (item *dbmodels.VacancyRequest, err error) {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения заявки")
		return nil, err
	}
	if rec == nil {
		return nil, errors.New("заявка не найдена")
	}
	return rec, nil
}

func (i impl) updateVr(store vacancyreqstore.Provider, spaceID, id string, data vacancyapimodels.VacancyRequestData) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	err := i.checkDependency(spaceID, data)
	if err != nil {
		return err
	}
	updMap := map[string]interface{}{
		"SpaceID":         spaceID,
		"CompanyID":       data.CompanyID,
		"DepartmentID":    data.DepartmentID,
		"JobTitleID":      data.JobTitleID,
		"CityID":          data.CityID,
		"CompanyStructID": data.CompanyStructID,
		"VacancyName":     data.VacancyName,
		"Confidential":    data.Confidential,
		"OpenedPositions": data.OpenedPositions,
		"Urgency":         data.Urgency,
		"RequestType":     data.RequestType,
		"SelectionType":   data.SelectionType,
		"PlaceOfWork":     data.PlaceOfWork,
		"ChiefFio":        data.ChiefFio,
		"Interviewer":     data.Interviewer,
		"ShortInfo":       data.ShortInfo,
		"Requirements":    data.Requirements,
		"Description":     data.Description,
		"OutInteraction":  data.OutInteraction,
		"InInteraction":   data.InInteraction,
		"Employment":      data.Employment,
		"Experience":      data.Experience,
		"Schedule":        data.Schedule,
	}
	err = store.Update(spaceID, id, updMap)
	if err != nil {
		logger.
			WithField("request", fmt.Sprintf("%+v", data)).
			WithError(err).
			Error("ошибка обновления заявки")
		return err
	}
	logger.Info("обновлена заявка")
	return nil
}

func (i impl) publish(spaceID, id, userID string) error {
	rec, err := i.getRec(spaceID, id)
	if err != nil {
		return err
	}
	if rec.Status != models.VRStatusAccepted {
		return errors.New("необходимо согласовать заявку")
	}
	data := vacancyapimodels.VacancyData{
		VacancyRequestID: rec.ID,
		CompanyID:        *rec.CompanyID,
		DepartmentID:     *rec.DepartmentID,
		JobTitleID:       *rec.JobTitleID,
		CityID:           *rec.CityID,
		CompanyStructID:  *rec.CompanyStructID,
		VacancyName:      rec.VacancyName,
		OpenedPositions:  rec.OpenedPositions,
		Urgency:          rec.Urgency,
		RequestType:      rec.RequestType,
		SelectionType:    rec.SelectionType,
		PlaceOfWork:      rec.PlaceOfWork,
		ChiefFio:         rec.ChiefFio,
		Requirements:     rec.Requirements,
		Salary:           vacancyapimodels.Salary{},
		Employment:       rec.Employment,
		Experience:       rec.Experience,
		Schedule:         rec.Schedule,
	}
	err = data.Validate(true)
	if err != nil {
		return err
	}
	_, err = i.vacancyHandler.Create(spaceID, userID, data)
	if err != nil {
		return err
	}
	return nil
}

func (i impl) CreateVacancy(spaceID, id, userID string) error {
	rec, err := i.getRec(spaceID, id)
	if err != nil {
		return err
	}
	if rec.Status != models.VRStatusAccepted {
		return errors.New("для создания вакансии, необходимо согласовать заявку")
	}
	exist, err := i.checkVacancyExist(spaceID, id, userID)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("вакансии уже создана")
	}
	return i.publish(spaceID, id, userID)
}

func (i impl) ToPin(id, userID string, isSet bool) error {
	if isSet {
		return i.store.SetPin(id, userID)
	}
	return i.store.RemovePin(id, userID)
}

func (i impl) ToFavorite(id, userID string, isSet bool) error {
	if isSet {
		return i.store.SetFavorite(id, userID)
	}
	return i.store.RemoveFavorite(id, userID)
}
