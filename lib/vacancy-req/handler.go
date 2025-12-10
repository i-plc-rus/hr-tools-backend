package vacancyreqhandler

import (
	"fmt"
	"hr-tools-backend/db"
	aprovaltaskhandler "hr-tools-backend/lib/aproval-task"
	approvaltaskstore "hr-tools-backend/lib/aproval-task/store"
	citystore "hr-tools-backend/lib/dicts/city/store"
	companyprovider "hr-tools-backend/lib/dicts/company"
	companystructprovider "hr-tools-backend/lib/dicts/company-struct"
	companystore "hr-tools-backend/lib/dicts/company/store"
	departmentprovider "hr-tools-backend/lib/dicts/department"
	jobtitleprovider "hr-tools-backend/lib/dicts/job-title"
	pushhandler "hr-tools-backend/lib/space/push/handler"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	vacancyhandler "hr-tools-backend/lib/vacancy"
	vacancyreqstore "hr-tools-backend/lib/vacancy-req/store"
	"hr-tools-backend/models"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Provider interface {
	Create(spaceID, userID string, data vacancyapimodels.VacancyRequestCreateData) (id, hMsg string, err error)
	GetByID(spaceID, id string) (item vacancyapimodels.VacancyRequestView, err error)
	Update(spaceID, id string, data vacancyapimodels.VacancyRequestEditData) (hMsg string, err error)
	Delete(spaceID, id string) error
	List(spaceID, userID string, filter vacancyapimodels.VrFilter) (list []vacancyapimodels.VacancyRequestView, rowCount int64, err error)
	ChangeStatus(spaceID, id, userID string, status models.VRStatus) (hMsh string, err error)
	CreateVacancy(spaceID, id, userID string) (hMsh string, err error)
	ToPin(id, userID string, isSet bool) error
	ToFavorite(id, userID string, isSet bool) error
	AddComment(spaceID, id string, data vacancyapimodels.Comment) error
	//согласование заявок
	Approve(spaceID, requestID, taskID, userID string) (hMsh string, err error)
	RequestChanges(spaceID, requestID, taskID, userID string, data vacancyapimodels.ApprovalRequestChanges) (hMsh string, err error)
	Reject(spaceID, requestID, taskID, userID string, data vacancyapimodels.ApprovalReject) (hMsh string, err error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store:                 vacancyreqstore.NewInstance(db.DB),
		approvalTaskStore:     approvaltaskstore.NewInstance(db.DB),
		companyProvider:       companyprovider.Instance,
		departmentProvider:    departmentprovider.Instance,
		jobTitleProvider:      jobtitleprovider.Instance,
		cityStore:             citystore.NewInstance(db.DB),
		companyStructProvider: companystructprovider.Instance,
		vacancyHandler:        vacancyhandler.Instance,
		aprovalTaskHandler:    aprovaltaskhandler.Instance,
		spaceUserStore:        spaceusersstore.NewInstance(db.DB),
	}
}

type impl struct {
	store                 vacancyreqstore.Provider
	approvalTaskStore     approvaltaskstore.Provider
	companyProvider       companyprovider.Provider
	departmentProvider    departmentprovider.Provider
	jobTitleProvider      jobtitleprovider.Provider
	cityStore             citystore.Provider
	companyStructProvider companystructprovider.Provider
	vacancyHandler        vacancyhandler.Provider
	aprovalTaskHandler    aprovaltaskhandler.Provider
	spaceUserStore        spaceusersstore.Provider
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

func (i impl) Create(spaceID, userID string, data vacancyapimodels.VacancyRequestCreateData) (id, hMsg string, err error) {
	logger := log.WithField("space_id", spaceID)
	err = i.checkDependency(spaceID, data.VacancyRequestData)
	if err != nil {
		return "", "", err
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
		rec.Status = models.VRStatusDraft
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
		aprovalStagesHandler := aprovaltaskhandler.NewHandlerWithTx(tx)
		if rec.CompanyID == nil && data.CompanyName != "" {
			companyID, err := createCompany(tx, spaceID, data.CompanyName)
			if err != nil {
				return errors.Wrap(err, "ошибка создания компании")
			}
			rec.CompanyID = &companyID
		}
		id, err = store.Create(rec)
		if err != nil {
			return err
		}
		hMsg, err = aprovalStagesHandler.Save(spaceID, id, data.ApprovalTasks.ApprovalTasks)
		return err
	})
	if err != nil {
		return "", "", err
	}
	if hMsg != "" {
		return "", hMsg, nil
	}
	logger.
		WithField("rec_id", id).
		Info("Создана заявка")
	return id, "", nil
}

func (i impl) GetByID(spaceID, id string) (item vacancyapimodels.VacancyRequestView, err error) {
	rec, err := i.getRec(spaceID, id)
	if err != nil {
		return vacancyapimodels.VacancyRequestView{}, err
	}
	return vacancyapimodels.VacancyRequestConvert(*rec), nil
}

func (i impl) Update(spaceID, id string, data vacancyapimodels.VacancyRequestEditData) (hMsg string, err error) {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if data.CompanyID == "" && data.CompanyName != "" {
			companyID, err := createCompany(tx, spaceID, data.CompanyName)
			if err != nil {
				return errors.Wrap(err, "ошибка создания компании")
			}
			data.CompanyID = companyID
		}
		store := vacancyreqstore.NewInstance(tx)
		aprovalStagesHandler := aprovaltaskhandler.NewHandlerWithTx(tx)
		err := i.updateVr(store, spaceID, id, data.VacancyRequestData)
		if err != nil {
			return err
		}
		hMsg, err = aprovalStagesHandler.Save(spaceID, id, data.ApprovalTasks.ApprovalTasks)
		return err
	})
	if err != nil {
		return "", err
	}
	if hMsg != "" {
		return hMsg, nil
	}
	logger.Info("обновлена заявка")
	return "", nil
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
		return nil, 0, err
	}
	result := make([]vacancyapimodels.VacancyRequestView, 0, len(list))
	for _, rec := range recList {
		result = append(result, vacancyapimodels.VacancyRequestConvert(rec))
	}
	return result, rowCount, nil
}

func (i impl) ChangeStatus(spaceID, id, userID string, status models.VRStatus) (hMsh string, err error) {
	logger := log.
		WithField("space_id", spaceID).
		WithField("rec_id", id).
		WithField("new_status", status)
	rec, err := i.getRec(spaceID, id)
	if err != nil {
		return "", err
	}
	if !rec.Status.IsAllowChange(status) {
		return fmt.Sprintf("изменение статуса на %v недопустимо", status), nil
	}
	updMap := map[string]interface{}{
		"status": status,
	}
	err = i.store.Update(spaceID, id, updMap)
	if err != nil {
		return "", err
	}
	logger.Info("статус заявки обновлен")
	if status == models.VRStatusCancelled {
		err = i.cancelVacancies(spaceID, id, userID)
		if err != nil {
			logger.WithError(err).Error("ошибка закрытия вакансии по заявке")
		}
		notification := models.GetPushVRClosed(rec.VacancyName, string(status))
		go i.sendNotification(*rec, notification)
	}
	return "", nil
}

func (i impl) cancelVacancies(spaceID, id, userID string) error {
	filter := vacancyapimodels.VacancyFilter{
		VacancyRequestID: id,
	}
	vacancyList, _, err := i.vacancyHandler.List(spaceID, userID, filter)
	if err != nil {
		return err
	}
	for _, vacancy := range vacancyList {
		err = i.vacancyHandler.StatusChange(spaceID, vacancy.ID, userID, models.VacancyStatusCanceled)
		if err != nil {
			return err
		}
	}
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

func (i impl) Approve(spaceID, requestID, taskID, userID string) (hMsh string, err error) {
	rec, taskRec, hMsh, err := i.approvalPrepare(spaceID, requestID, taskID, userID)
	if hMsh != "" || err != nil {
		return hMsh, err
	}
	if !rec.Status.AllowAccept() {
		return fmt.Sprintf("невозможно согласовать заявку в текущем статусе: %v", rec.Status), nil
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		aprovalStagesHandler := aprovaltaskhandler.NewHandlerWithTx(tx)
		//меняем статус задачи согласования
		updMap := map[string]interface{}{
			"State":     models.AStateApproved,
			"Comment":   "",
			"DecidedAt": nil,
		}
		approvalTaskStore := approvaltaskstore.NewInstance(tx)
		err = approvalTaskStore.Update(spaceID, taskID, updMap)
		if err != nil {
			return err
		}
		if taskRec != nil {
			// для аудита
			taskRec.State = models.AStateApproved
			taskRec.Comment = ""
			taskRec.DecidedAt = nil
			aprovalStagesHandler.Audit(*taskRec)
		}
		taskList, err := approvalTaskStore.List(spaceID, requestID)
		if err != nil {
			return err
		}
		allAprove := true
		for _, task := range taskList {
			if task.State != models.AStateApproved {
				allAprove = false
				break
			}
		}
		if allAprove {
			//все согласовали, меняем статус заявки
			hMsh, err = i.ChangeStatus(spaceID, requestID, userID, models.VRStatusApproved)
			if err != nil {
				return err
			}
			if hMsh != "" {
				return errors.New(hMsh)
			}

			aprovalStagesHandler := aprovaltaskhandler.NewHandlerWithTx(tx)
			auditRec := dbmodels.ApprovalHistory{
				BaseSpaceModel: dbmodels.BaseSpaceModel{SpaceID: spaceID},
				RequestID:      requestID,
				AssigneeUserID: userID,
				Comment:        "Заявка полностью согласована",
				Changes: dbmodels.EntityChanges{
					Description: "Изменен статус заявки",
					Data: []dbmodels.FieldChanges{
						{
							Field:    "Status",
							OldValue: rec.Status,
							NewValue: models.VRStatusApproved},
					},
				},
			}
			aprovalStagesHandler.AuditCommon(auditRec)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	go func(rec dbmodels.VacancyRequest) {
		code := models.PushVRApproved
		logger := log.WithField("space_id", spaceID).
			WithField("rec_id", requestID).
			WithField("event_code", code)
		user, err := i.spaceUserStore.GetByID(userID)
		if err != nil {
			logger.WithError(err).Error("ошибка получения пользователя")
			return
		}
		if user == nil {
			logger.Error("пользователь не найден")
			return
		}
		notification := models.GetPushVRApproved(rec.VacancyName, user.GetFullName())
		i.sendNotification(rec, notification)
	}(*rec)
	return "", nil
}

func (i impl) RequestChanges(spaceID, requestID, taskID, userID string, data vacancyapimodels.ApprovalRequestChanges) (hMsh string, err error) {
	rec, taskRec, hMsh, err := i.approvalPrepare(spaceID, requestID, taskID, userID)
	if hMsh != "" || err != nil {
		return hMsh, err
	}
	if !rec.Status.AllowReject() {
		return fmt.Sprintf("невозможно отправить на доработку заявку в текущем статусе: %v", rec.Status), nil
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		aprovalStagesHandler := aprovaltaskhandler.NewHandlerWithTx(tx)
		//меняем статус задачи согласования
		now := time.Now()
		updMap := map[string]interface{}{
			"State":     models.AStateRequestChanges,
			"Comment":   data.Comment,
			"DecidedAt": now,
		}
		approvalTaskStore := approvaltaskstore.NewInstance(tx)
		err = approvalTaskStore.Update(spaceID, taskID, updMap)
		if err != nil {
			return err
		}

		//меняем статус заявки
		hMsh, err = i.ChangeStatus(spaceID, requestID, userID, models.VRStatusCreated)
		if err != nil {
			return err
		}
		if hMsh != "" {
			return errors.New(hMsh)
		}
		if taskRec != nil {
			// для аудита
			taskRec.State = models.AStateRequestChanges
			taskRec.Comment = data.Comment
			taskRec.DecidedAt = &now
			aprovalStagesHandler.Audit(*taskRec)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return "", nil
}

func (i impl) Reject(spaceID, requestID, taskID, userID string, data vacancyapimodels.ApprovalReject) (hMsh string, err error) {
	rec, taskRec, hMsh, err := i.approvalPrepare(spaceID, requestID, taskID, userID)
	if hMsh != "" || err != nil {
		return hMsh, err
	}
	if !rec.Status.AllowReject() {
		return fmt.Sprintf("невозможно отклонить заявку в текущем статусе: %v", rec.Status), nil
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		aprovalStagesHandler := aprovaltaskhandler.NewHandlerWithTx(tx)
		now := time.Now()
		//меняем статус задачи согласования
		updMap := map[string]interface{}{
			"State":     models.AStateRejected,
			"Comment":   data.Comment,
			"DecidedAt": now,
		}
		approvalTaskStore := approvaltaskstore.NewInstance(tx)
		err = approvalTaskStore.Update(spaceID, taskID, updMap)
		if err != nil {
			return err
		}

		//меняем статус заявки
		hMsh, err = i.ChangeStatus(spaceID, requestID, userID, models.VRStatusRejected)
		if err != nil {
			return err
		}
		if hMsh != "" {
			return errors.New(hMsh)
		}
		if taskRec != nil {
			// для аудита
			taskRec.State = models.AStateRejected
			taskRec.Comment = data.Comment
			taskRec.DecidedAt = &now
			aprovalStagesHandler.Audit(*taskRec)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	go func(rec dbmodels.VacancyRequest) {
		code := models.PushVRRejected
		logger := log.WithField("space_id", spaceID).
			WithField("rec_id", requestID).
			WithField("event_code", code)
		user, err := i.spaceUserStore.GetByID(userID)
		if err != nil {
			logger.WithError(err).Error("ошибка получения пользователя")
			return
		}
		if user == nil {
			logger.Error("пользователь не найден")
			return
		}
		notification := models.GetPushVRRejected(rec.VacancyName, user.GetFullName(), user.Role.ToHuman())
		i.sendNotification(rec, notification)
	}(*rec)
	return "", nil
}

func (i impl) approvalPrepare(spaceID, requestID, taskID, userID string) (vrRec *dbmodels.VacancyRequest, taskRec *dbmodels.ApprovalTask, hMsh string, err error) {
	vacancyRequest, err := i.getRec(spaceID, requestID)
	if err != nil {
		return nil, nil, "", err
	}
	if vacancyRequest == nil {
		return nil, nil, "Заявка не найдена", nil
	}

	if !vacancyRequest.Status.IsAllowChange(models.VRStatusCreated) {
		return nil, nil, fmt.Sprintf("невозможно отклонить заявку в текущем статусе: %v", vacancyRequest.Status), nil
	}

	task, err := i.approvalTaskStore.GetByID(spaceID, taskID)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "ошибка получения задачи на согласование")
	}
	if task == nil || task.RequestID != requestID {
		return nil, nil, "Задача на согласование не найдена", nil
	}

	if task.AssigneeUserID != userID {
		return nil, nil, "На данную задачу назначен другой пользователь", nil
	}
	return vacancyRequest, task, "", nil
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
	if rec.Status != models.VRStatusApproved {
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
	_, hMsg, err := i.vacancyHandler.Create(spaceID, userID, data)
	if err != nil {
		return err
	}
	if hMsg != "" {
		return errors.New(hMsg)
	}
	return nil
}

func (i impl) CreateVacancy(spaceID, id, userID string) (hMsh string, err error) {
	rec, err := i.getRec(spaceID, id)
	if err != nil {
		return "", err
	}
	if rec.Status != models.VRStatusApproved {
		return "для создания вакансии, необходимо согласовать заявку", nil
	}
	exist, err := i.checkVacancyExist(spaceID, id, userID)
	if err != nil {
		return "", err
	}
	if exist {
		return "вакансии уже создана", nil
	}
	return "", i.publish(spaceID, id, userID)
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

func (i impl) AddComment(spaceID, id string, data vacancyapimodels.Comment) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)

	// get vacancy request for sure that it exists
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("заявка на вакансию не найдена")
	}

	spaceUser, err := i.spaceUserStore.GetByID(data.AuthorID)
	if err != nil {
		return err
	}
	if spaceUser == nil || spaceUser.SpaceID != spaceID {
		return errors.New("автор не найден")
	}

	dbComment := dbmodels.VacancyRequestComment{
		ID:               uuid.New().String(),
		VacancyRequestID: id,
		Date:             time.Now(),
		AuthorID:         data.AuthorID,
		Comment:          data.Comment,
	}

	if err := i.store.AddComment(dbComment); err != nil {
		logger.Error("не удалось сохранить комментарий", "error", err)
		return err
	}
	logger.Info("добавлен комментарий к заявке на вакансию")
	return nil
}

func createCompany(tx *gorm.DB, spaceID, name string) (string, error) {
	companyStore := companystore.NewInstance(tx)
	return companyStore.FindOrCreate(spaceID, name)
}

func (i impl) sendNotification(rec dbmodels.VacancyRequest, data models.NotificationData) {
	//отправляем автору
	pushhandler.Instance.SendNotification(rec.AuthorID, data)
	approvalTasks, err := i.approvalTaskStore.List(rec.SpaceID, rec.ID)
	if err != nil {
		log.
			WithError(err).
			WithField("space_id", rec.SpaceID).
			WithField("rec_id", rec.ID).Error("Ошибка получения списка пользователей из цепочки согласования для отправки уведомлений")
		return
	}
	for _, stage := range approvalTasks {
		//отправляем списку пользователей из цепочки согласования
		if rec.AuthorID == stage.AssigneeUserID {
			continue
		}
		pushhandler.Instance.SendNotification(stage.AssigneeUserID, data)
	}
}
