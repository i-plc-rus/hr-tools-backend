package vacancyhandler

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"hr-tools-backend/db"
	citystore "hr-tools-backend/lib/dicts/city/store"
	companyprovider "hr-tools-backend/lib/dicts/company"
	companystructprovider "hr-tools-backend/lib/dicts/company-struct"
	departmentprovider "hr-tools-backend/lib/dicts/department"
	jobtitleprovider "hr-tools-backend/lib/dicts/job-title"
	selectionstagestore "hr-tools-backend/lib/vacancy/selection-stage-store"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	"hr-tools-backend/models"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
	"sort"
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
	StageDelete(spaceID, id, stageID string) error
	StageChangeOrder(spaceID, id, stageID string, newOrder int) error
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
	logger := log.WithField("space_id", spaceID)
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
		store := vacancystore.NewInstance(tx)
		recID, err = store.Create(rec)
		if err != nil {
			logger.
				WithField("request", fmt.Sprintf("%+v", data)).
				WithError(err).
				Error("ошибка создания вакансии")
			return errors.New("ошибка создания вакансии")
		}
		err = i.initSelectionStages(tx, spaceID, recID)
		if err != nil {
			logger.WithError(err).Error("ошибка инициализации этапов подбора")
			return errors.New("ошибка инициализации этапов подбора")
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
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения вакансии")
		return vacancyapimodels.VacancyView{}, errors.New("ошибка получения вакансии")
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
	err = i.store.Update(spaceID, id, updMap)
	if err != nil {
		logger.
			WithField("request", fmt.Sprintf("%+v", data)).
			WithError(err).
			Error("ошибка обновления вакансии")
		return err
	}
	logger.Info("обновлена вакансия")
	return nil
}

func (i impl) Delete(spaceID, id string) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	err := i.store.Delete(spaceID, id)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка удаления вакансии")
		return err
	}
	logger.Info("удалена вакансия")
	return nil
}

func (i impl) List(spaceID, userID string, filter vacancyapimodels.VacancyFilter) (list []vacancyapimodels.VacancyView, rowCount int64, err error) {
	logger := log.WithField("space_id", spaceID)
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
		logger.
			WithError(err).
			Error("ошибка получения списка заявок")
		return nil, 0, err
	}
	result := make([]vacancyapimodels.VacancyView, 0, len(list))
	for _, rec := range recList {
		result = append(result, vacancyapimodels.VacancyConvert(rec))
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
	logger := log.
		WithField("space_id", spaceID).
		WithField("vacancy_id", id)
	recList, err := i.selectionStageStore.List(spaceID, id)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения списка этапов подбора")
		return nil, errors.New("ошибка получения списка этапов подбора")
	}
	result := make([]vacancyapimodels.SelectionStageView, 0, len(list))
	for _, rec := range recList {
		result = append(result, vacancyapimodels.SelectionStageConvert(rec))
	}
	return result, nil
}

func (i impl) StageCreate(spaceID, id string, data vacancyapimodels.SelectionStageAdd) error {
	logger := log.
		WithField("space_id", spaceID).
		WithField("vacancy_id", id)
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
		logger.
			WithError(err).
			Error("ошибка добавления этапа подбора")
		return errors.New("ошибка добавления этапа подбора")
	}
	return nil
}

func (i impl) StageDelete(spaceID, id, stageID string) error {
	logger := log.
		WithField("space_id", spaceID).
		WithField("vacancy_id", id).
		WithField("stage_id", stageID)
	rec, err := i.selectionStageStore.GetByID(spaceID, id, stageID)
	if err != nil || rec == nil {
		return err
	}
	if !rec.CanDelete {
		return errors.New("этап нельзя удалить")
	}
	err = i.selectionStageStore.Delete(spaceID, id, stageID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка удаления этапа подбора")
		return errors.New("ошибка удаления этапа подбора")
	}
	return nil
}

func (i impl) StageChangeOrder(spaceID, id, stageID string, newOrder int) error {
	logger := log.
		WithField("space_id", spaceID).
		WithField("vacancy_id", id)
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		selectionStageStore := selectionstagestore.NewInstance(tx)
		list, err := selectionStageStore.List(spaceID, id)
		if err != nil {
			logger.
				WithError(err).
				Error("ошибка получения списка этапов подбора")
			return errors.New("ошибка получения списка этапов подбора")
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
		logger.
			WithError(err).
			Error("ошибка изменения порядка в списке этапов подбора")
		return errors.New("ошибка изменения порядка в списке этапов подбора")
	}
	logger.Info("изменен порядок списка этапов подбора")
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
