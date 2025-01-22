package vacancyhandler

import (
	"context"
	"fmt"
	"hr-tools-backend/db"
	applicanthistoryhandler "hr-tools-backend/lib/applicant-history"
	applicantstore "hr-tools-backend/lib/applicant/store"
	citystore "hr-tools-backend/lib/dicts/city/store"
	companyprovider "hr-tools-backend/lib/dicts/company"
	companystructprovider "hr-tools-backend/lib/dicts/company-struct"
	companystore "hr-tools-backend/lib/dicts/company/store"
	departmentprovider "hr-tools-backend/lib/dicts/department"
	jobtitleprovider "hr-tools-backend/lib/dicts/job-title"
	avitohandler "hr-tools-backend/lib/external-services/avito"
	hhhandler "hr-tools-backend/lib/external-services/hh"
	pushhandler "hr-tools-backend/lib/space/push/handler"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	selectionstagestore "hr-tools-backend/lib/vacancy/selection-stage-store"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	teamstore "hr-tools-backend/lib/vacancy/team-store"
	"hr-tools-backend/models"
	apimodels "hr-tools-backend/models/api"
	applicantapimodels "hr-tools-backend/models/api/applicant"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
	"sort"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Provider interface {
	Create(spaceID, userID string, data vacancyapimodels.VacancyData) (id string, err error)
	GetByID(spaceID, id string) (item vacancyapimodels.VacancyView, err error)
	Update(spaceID, id string, data vacancyapimodels.VacancyData) error
	Delete(spaceID, id string) error
	List(spaceID, userID string, filter vacancyapimodels.VacancyFilter) (list []vacancyapimodels.VacancyView, rowCount int64, err error)
	ToPin(id, userID string, isSet bool) error
	ToFavorite(id, userID string, isSet bool) error
	StageList(spaceID, id string) (list []vacancyapimodels.SelectionStageView, err error)
	StageCreate(spaceID, id string, data vacancyapimodels.SelectionStageAdd) error
	StageDelete(spaceID, id, stageID string) (hMsh string, err error)
	StageChangeOrder(spaceID, id, stageID string, newOrder int) error
	StatusChange(spaceID, id, userID string, status models.VacancyStatus) error
	GetTeam(spaceID, vacancyID string) (result []vacancyapimodels.TeamPerson, err error)
	InviteToTeam(tx *gorm.DB, spaceID, vacancyID, userID string, responsible bool) (id string, err error)
	UsersList(spaceID, vacancyID string, filter vacancyapimodels.PersonFilter) (result []vacancyapimodels.Person, err error)
	ExecuteFromTeam(spaceID, vacancyID, userID string) (hMsg string, err error)
	SetAsResponsible(spaceID, vacancyID, userID string) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store:                 vacancystore.NewInstance(db.DB),
		selectionStageStore:   selectionstagestore.NewInstance(db.DB),
		companyProvider:       companyprovider.Instance,
		departmentProvider:    departmentprovider.Instance,
		jobTitleProvider:      jobtitleprovider.Instance,
		cityStore:             citystore.NewInstance(db.DB),
		companyStructProvider: companystructprovider.Instance,
		applicantHistory:      applicanthistoryhandler.Instance,
		applicantStore:        applicantstore.NewInstance(db.DB),
		teamStore:             teamstore.NewInstance(db.DB),
		spaceUserStore:        spaceusersstore.NewInstance(db.DB),
	}
}

type impl struct {
	store                 vacancystore.Provider
	selectionStageStore   selectionstagestore.Provider
	companyProvider       companyprovider.Provider
	departmentProvider    departmentprovider.Provider
	jobTitleProvider      jobtitleprovider.Provider
	cityStore             citystore.Provider
	companyStructProvider companystructprovider.Provider
	applicantHistory      applicanthistoryhandler.Provider
	applicantStore        applicantstore.Provider
	teamStore             teamstore.Provider
	spaceUserStore        spaceusersstore.Provider
}

func (i impl) checkDependency(spaceID string, data vacancyapimodels.VacancyData) (err error) {
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

func (i impl) Create(spaceID, userID string, data vacancyapimodels.VacancyData) (id string, err error) {
	logger := i.getLogger(spaceID, "", userID)
	err = i.checkDependency(spaceID, data)
	if err != nil {
		return "", err
	}
	recID := ""
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		rec := dbmodels.Vacancy{
			BaseSpaceModel: dbmodels.BaseSpaceModel{
				SpaceID: spaceID,
			},
			VacancyName:     data.VacancyName,
			OpenedPositions: data.OpenedPositions,
			Urgency:         data.Urgency,
			RequestType:     data.RequestType,
			SelectionType:   data.SelectionType,
			PlaceOfWork:     data.PlaceOfWork,
			ChiefFio:        data.ChiefFio,
			Requirements:    data.Requirements,
			Salary: dbmodels.Salary{
				From:     data.Salary.From,
				To:       data.Salary.To,
				ByResult: data.Salary.ByResult,
				InHand:   data.Salary.InHand,
			},
			AuthorID:   userID,
			Status:     models.VacancyStatusOpened,
			Employment: data.Employment,
			Experience: data.Experience,
			Schedule:   data.Schedule,
		}
		if data.VacancyRequestID != "" {
			rec.VacancyRequestID = &data.VacancyRequestID
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
		if rec.CompanyID == nil && data.CompanyName != "" {
			companyID, err := createCompany(tx, spaceID, data.CompanyName)
			if err != nil {
				return errors.Wrap(err, "ошибка создания компании")
			}
			rec.CompanyID = &companyID
		}
		store := vacancystore.NewInstance(tx)
		recID, err = store.Create(rec)
		if err != nil {
			return err
		}
		err = i.initSelectionStages(tx, spaceID, recID)
		if err != nil {
			return errors.Wrap(err, "ошибка инициализации этапов подбора")
		}
		_, err = i.InviteToTeam(tx, spaceID, recID, userID, true)
		if err != nil {
			return errors.Wrap(err, "Ошибка приглашения участника в команду")
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	logger.
		WithField("rec_id", recID).
		Info("Создана вакансия")
	return recID, nil
}

func (i impl) GetByID(spaceID, id string) (item vacancyapimodels.VacancyView, err error) {
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return vacancyapimodels.VacancyView{}, err
	}
	if rec == nil {
		return vacancyapimodels.VacancyView{}, errors.New("вакансия не найдена")
	}
	recExt := dbmodels.VacancyExt{
		Vacancy:  *rec,
		Favorite: false,
		Pinned:   false,
	}
	return vacancyapimodels.VacancyConvert(recExt), nil
}

func (i impl) Update(spaceID, id string, data vacancyapimodels.VacancyData) error {
	logger := i.getLogger(spaceID, id, "")
	err := i.checkDependency(spaceID, data)
	if err != nil {
		return err
	}
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if data.CompanyID == "" && data.CompanyName != "" {
			companyID, err := createCompany(tx, spaceID, data.CompanyName)
			if err != nil {
				return errors.Wrap(err, "ошибка создания компании")
			}
			data.CompanyID = companyID
		}

		updMap := map[string]interface{}{
			"SpaceID":         spaceID,
			"CompanyID":       data.CompanyID,
			"DepartmentID":    data.DepartmentID,
			"JobTitleID":      data.JobTitleID,
			"CityID":          data.CityID,
			"CompanyStructID": data.CompanyStructID,
			"VacancyName":     data.VacancyName,
			"OpenedPositions": data.OpenedPositions,
			"Urgency":         data.Urgency,
			"RequestType":     data.RequestType,
			"SelectionType":   data.SelectionType,
			"PlaceOfWork":     data.PlaceOfWork,
			"ChiefFio":        data.ChiefFio,
			"Requirements":    data.Requirements,
			"salary_from":     data.Salary.From,
			"salary_to":       data.Salary.To,
			"salary_result":   data.Salary.ByResult,
			"salary_in_hand":  data.Salary.InHand,
			"Employment":      data.Employment,
			"Experience":      data.Experience,
			"Schedule":        data.Schedule,
		}
		store := vacancystore.NewInstance(tx)
		err = store.Update(spaceID, id, updMap)
		if err != nil {
			return errors.Wrap(err, "ошибка обновления вакансии")
		}
		return nil
	})
	if err != nil {
		return err
	}
	logger.Info("обновлена вакансия")
	return nil
}

func (i impl) Delete(spaceID, id string) error {
	logger := i.getLogger(spaceID, id, "")
	err := i.store.Delete(spaceID, id)
	if err != nil {
		return err
	}
	logger.Info("удалена вакансия")
	return nil
}

func (i impl) List(spaceID, userID string, filter vacancyapimodels.VacancyFilter) (list []vacancyapimodels.VacancyView, rowCount int64, err error) {
	logger := i.getLogger(spaceID, "", userID)
	rowCount, err = i.store.ListCount(spaceID, userID, filter)
	if err != nil {
		return nil, 0, err
	}

	page, limit := filter.GetPage()
	offset := (page - 1) * limit
	if int64(offset) > rowCount {
		return []vacancyapimodels.VacancyView{}, rowCount, nil
	}

	recList, err := i.store.List(spaceID, userID, filter)
	if err != nil {
		return nil, 0, err
	}
	if len(recList) == 0 {
		return nil, 0, nil
	}

	ids := make([]string, 0, len(list))
	for _, rec := range recList {
		ids = append(ids, rec.ID)
	}
	stagesMap := map[string][]dbmodels.ApplicantsStage{}
	stages, err := i.applicantStore.ApplicantsByStages(spaceID, ids)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения списка активных кандидатов по этапам")
	} else {
		for _, stage := range stages {
			list, ok := stagesMap[stage.VacancyID]
			if !ok {
				list = make([]dbmodels.ApplicantsStage, 0, 10)
			}
			list = append(list, stage)
			stagesMap[stage.VacancyID] = list
		}
	}
	result := make([]vacancyapimodels.VacancyView, 0, len(list))
	for _, rec := range recList {
		item := vacancyapimodels.VacancyConvert(rec)
		if stages, ok := stagesMap[rec.ID]; ok {
			for k, selectionStage := range item.SelectionStages {
				for _, stage := range stages {
					if stage.SelectionStageID == selectionStage.ID {
						item.SelectionStages[k].Total = stage.Total
					}
				}
			}
		}
		result = append(result, item)
	}
	return result, rowCount, nil
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

func (i impl) StageList(spaceID, id string) (list []vacancyapimodels.SelectionStageView, err error) {
	recList, err := i.selectionStageStore.List(spaceID, id)
	if err != nil {
		return nil, err
	}
	result := make([]vacancyapimodels.SelectionStageView, 0, len(list))
	for _, rec := range recList {
		result = append(result, vacancyapimodels.SelectionStageConvert(rec))
	}
	return result, nil
}

func (i impl) StageCreate(spaceID, id string, data vacancyapimodels.SelectionStageAdd) error {
	rec := dbmodels.SelectionStage{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		VacancyID:  id,
		Name:       data.Name,
		StageType:  data.StageType,
		CanDelete:  true,
		LimitValue: data.LimitValue,
		LimitType:  data.LimitType,
	}
	id, err := i.selectionStageStore.Create(rec)
	if err != nil {
		return err
	}
	return nil
}

func (i impl) StageDelete(spaceID, id, stageID string) (hMsh string, err error) {
	rec, err := i.selectionStageStore.GetByID(spaceID, id, stageID)
	if err != nil || rec == nil {
		return "", err
	}
	if !rec.CanDelete {
		return "этап нельзя удалить", nil
	}
	err = i.selectionStageStore.Delete(spaceID, id, stageID)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (i impl) StageChangeOrder(spaceID, id, stageID string, newOrder int) error {
	logger := i.getLogger(spaceID, id, "").
		WithField("stage_id", stageID)
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		selectionStageStore := selectionstagestore.NewInstance(tx)
		list, err := selectionStageStore.List(spaceID, id)
		if err != nil {
			return errors.Wrap(err, "ошибка получения списка этапов подбора")
		}

		var changed *dbmodels.SelectionStage
		cnangedList := make([]dbmodels.SelectionStage, 0, len(list))
		for k, rec := range list {
			if rec.ID == stageID {
				if rec.StageOrder != newOrder {
					changed = &rec
					cnangedList = list[:k]
					cnangedList = append(cnangedList, list[k+1:]...)
					break
				}
			}
		}
		if changed == nil {
			return nil
		}
		sort.Slice(cnangedList, func(i, j int) bool {
			return cnangedList[i].StageOrder < cnangedList[j].StageOrder
		})
		newSet := make([]dbmodels.SelectionStage, 0, len(list))
		order := 1
		for k, rec := range cnangedList {
			if k+1 == newOrder {
				changed.StageOrder = order
				newSet = append(newSet, *changed)
				order++
			}
			rec.StageOrder = order
			newSet = append(newSet, rec)
			order++
		}
		if len(cnangedList) < newOrder {
			changed.StageOrder = order
			newSet = append(newSet, *changed)
		}

		for _, rec := range newSet {
			updMap := map[string]interface{}{
				"stage_order": rec.StageOrder,
			}
			if err = selectionStageStore.Update(spaceID, id, rec.ID, updMap); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	logger.Info("изменен порядок списка этапов подбора")
	return nil
}

func (i impl) StatusChange(spaceID, vacancyID, userID string, status models.VacancyStatus) error {
	logger := i.getLogger(spaceID, vacancyID, userID).
		WithField("status", status)
	rec, err := i.store.GetByID(spaceID, vacancyID)
	if err != nil {
		return errors.Wrap(err, "ошибка получения вакансии")
	}
	if rec == nil {
		return errors.New("вакансия не найдена")
	}
	if rec.Status == status {
		return nil
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// смена статуса вакансии
		updMap := map[string]interface{}{
			"status": status,
		}
		store := vacancystore.NewInstance(tx)
		err = store.Update(spaceID, vacancyID, updMap)
		if err != nil {
			return errors.Wrap(err, "ошибка обновления статуса вакансии")
		}
		if !status.IsClosed() {
			return nil
		}

		//получение списка кандидатов по вакансии
		applicantStore := applicantstore.NewInstance(tx)
		filter := applicantapimodels.ApplicantFilter{
			VacancyID: vacancyID,
			Pagination: apimodels.Pagination{
				Limit: 100,
			},
		}
		list, err := applicantStore.ListOfApplicant(spaceID, filter)
		if err != nil {
			return errors.Wrap(err, "ошибка получения списка кандидатов по вакансии")
		}
		reason := fmt.Sprintf("Вакансия %v", status)
		applicantHistory := applicanthistoryhandler.NewTxHandler(tx)
		for _, applicantRec := range list {
			if applicantRec.Status == models.ApplicantStatusArchive {
				continue
			}
			//перевод кандидата в архив
			updMap = map[string]interface{}{
				"status": models.ApplicantStatusArchive,
			}
			err = applicantStore.Update(applicantRec.ID, updMap)
			if err != nil {
				return errors.Wrapf(err, "ошибка перевода кандидата (%v) в архив", applicantRec.ID)
			}
			//добавление в историю по кандидату
			changes := applicanthistoryhandler.GetArchiveChange(reason)
			applicantHistory.Save(spaceID, applicantRec.ID, vacancyID, userID, dbmodels.HistoryTypeArchive, changes)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if status != models.VacancyStatusOpened {
		err = i.cancelJobSite(*rec, *logger)
		if err != nil {
			return err
		}
	}

	logger.Info("обновлен статус вакансии")
	rec.Status = status
	go func(rec dbmodels.Vacancy) {
		notification := models.GetPushVacancyNewStatus(rec.VacancyName, string(status))
		i.sendNotification(rec, notification)
	}(*rec)
	return nil
}

func (i impl) GetTeam(spaceID, vacancyID string) (result []vacancyapimodels.TeamPerson, err error) {
	recList, err := i.teamStore.List(spaceID, vacancyID)
	if err != nil {
		return nil, err
	}
	result = make([]vacancyapimodels.TeamPerson, 0, len(recList))

	for _, rec := range recList {
		result = append(result, vacancyapimodels.TeamPersonConvert(rec))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].FullName < result[j].FullName
	})
	return result, nil
}

func (i impl) InviteToTeam(tx *gorm.DB, spaceID, vacancyID, userID string, responsible bool) (id string, err error) {
	vt, err := i.teamStore.GetByID(spaceID, vacancyID, userID)
	if err != nil || vt != nil {
		return "", err
	}

	user, err := i.spaceUserStore.GetByID(userID)
	if err != nil {
		return "", err
	}
	if user == nil || user.SpaceID != spaceID {
		return "", errors.New("Участник не найден")
	}
	rec := dbmodels.VacancyTeam{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		UserID:      userID,
		VacancyID:   vacancyID,
		Responsible: responsible,
	}
	teamStore := i.teamStore
	if tx != nil {
		teamStore = teamstore.NewInstance(tx)
	}
	id, err = teamStore.Create(rec)
	if err != nil {
		return "", errors.Wrap(err, "Ошибка добавления участника в команду")
	}
	return id, nil
}

func (i impl) UsersList(spaceID, vacancyID string, filter vacancyapimodels.PersonFilter) (result []vacancyapimodels.Person, err error) {
	recList, err := i.spaceUserStore.GetListForVacancy(spaceID, vacancyID, filter)
	if err != nil {
		return nil, err
	}
	result = make([]vacancyapimodels.Person, 0, len(recList))
	for _, rec := range recList {
		result = append(result, vacancyapimodels.PersonConvert(rec))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].FullName < result[j].FullName
	})
	return result, nil
}

func (i impl) ExecuteFromTeam(spaceID, vacancyID, userID string) (hMsg string, err error) {
	vt, err := i.teamStore.GetByID(spaceID, vacancyID, userID)
	if err != nil || vt == nil {
		return "", err
	}
	if vt.Responsible {
		return "Перед исключением из команды, необходимо назначить другого ответственного", nil
	}
	return "", i.teamStore.Delete(spaceID, vacancyID, userID)
}

func (i impl) SetAsResponsible(spaceID, vacancyID, userID string) error {
	vt, err := i.teamStore.GetByID(spaceID, vacancyID, userID)
	if err != nil {
		return err
	}
	if vt == nil {
		return errors.New("пользователь не в команде")
	}
	err = i.teamStore.SetAsResponsible(spaceID, vacancyID, userID)
	if err != nil {
		return err
	}
	go func() {
		logger := log.
			WithField("space_id", spaceID).
			WithField("user_id", userID).
			WithField("event_code", models.PushVacancyResponsible).
			WithField("vacancy_id", vacancyID)
		rec, err := i.store.GetByID(spaceID, vacancyID)
		if err != nil {
			logger.WithError(err).Error("ошибка получения вакансии")
			return
		}
		if rec == nil {
			logger.Error("вакансия не найдена")
			return
		}
		notification := models.GetPushVacancyResponsible(rec.VacancyName, vt.SpaceUser.GetFullName())
		i.sendNotification(*rec, notification)
	}()
	return nil
}

func (i impl) initSelectionStages(tx *gorm.DB, spaceID, vacancyID string) error {
	selectionStageStore := selectionstagestore.NewInstance(tx)
	for k, name := range dbmodels.DefaultSelectionStages {
		rec := dbmodels.SelectionStage{
			BaseSpaceModel: dbmodels.BaseSpaceModel{
				SpaceID: spaceID,
			},
			VacancyID:  vacancyID,
			StageOrder: k,
			Name:       name,
			StageType:  "",
			CanDelete:  false,
		}
		_, err := selectionStageStore.Create(rec)
		if err != nil {
			return errors.Wrapf(err, "ошибка добавления этапа подбора: %v", name)
		}
	}
	return nil
}

func (i *impl) getLogger(spaceID, vacancyID, userID string) *log.Entry {
	logger := log.WithField("space_id", spaceID)
	if vacancyID != "" {
		logger = logger.WithField("vacancy_id", vacancyID)
	}
	if userID != "" {
		logger = logger.WithField("user_id", userID)
	}
	return logger
}

func (i impl) cancelJobSite(rec dbmodels.Vacancy, logger log.Entry) error {
	errorList := []string{}
	if rec.AvitoID != 0 && (rec.AvitoStatus == models.VacancyPubStatusModeration || rec.AvitoStatus != models.VacancyPubStatusPublished) {
		err := avitohandler.Instance.VacancyClose(context.TODO(), rec.SpaceID, rec.ID)
		if err != nil {
			logger.
				WithError(err).
				Error("не удалось снять вакансию с публикации на Avito")
			errorList = append(errorList, "не удалось снять вакансию с публикации на Avito")
		}
		logger.Info("вакансия снята с публикации на Avito")
	}
	if rec.HhID != "" && (rec.HhStatus == models.VacancyPubStatusModeration || rec.HhStatus != models.VacancyPubStatusPublished) {
		err := hhhandler.Instance.VacancyClose(context.TODO(), rec.SpaceID, rec.ID)
		if err != nil {
			logger.
				WithError(err).
				Error("не удалось снять вакансию с публикации на HeadHunter")
			errorList = append(errorList, "не удалось снять вакансию с публикации на HeadHunter")
		}
		logger.Info("вакансия снята с публикации на HeadHunter")
	}
	if len(errorList) != 0 {
		return errors.Errorf("%v", errorList)
	}
	return nil
}

func createCompany(tx *gorm.DB, spaceID, name string) (string, error) {
	companyStore := companystore.NewInstance(tx)
	return companyStore.FindOrCreate(spaceID, name)
}

func (i impl) sendNotification(rec dbmodels.Vacancy, data models.NotificationData) {
	pushhandler.Instance.SendNotification(rec.AuthorID, data)
	for _, teamMember := range rec.VacancyTeam {
		//отправляем команде
		if rec.AuthorID == teamMember.ID {
			continue
		}
		pushhandler.Instance.SendNotification(teamMember.ID, data)
	}
}
