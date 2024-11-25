package applicant

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	vacancyhandler "hr-tools-backend/lib/vacancy"
	selectionstagestore "hr-tools-backend/lib/vacancy/selection-stage-store"
	"hr-tools-backend/models"
	applicantapimodels "hr-tools-backend/models/api/applicant"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

type Provider interface {
	ListOfNegotiation(spaceID string, filter dbmodels.NegotiationFilter) (list []negotiationapimodels.NegotiationView, err error)
	UpdateComment(id string, comment string) error
	UpdateStatus(spaceID, id string, status models.NegotiationStatus) error
	GetByID(spaceID, id string) (negotiationapimodels.NegotiationView, error)
	CreateApplicant(spaceID string, applicant applicantapimodels.ApplicantData) (string, error)
	GetApplicant(spaceID string, id string) (applicantapimodels.ApplicantViewExt, error)
	ListOfApplicant(spaceID string, filter applicantapimodels.ApplicantFilter) (list []applicantapimodels.ApplicantView, rowCount int64, err error)
	UpdateApplicant(spaceID string, id string, applicant applicantapimodels.ApplicantData) error
	ApplicantAddTag(spaceID string, id string, tag string) error
	ApplicantRemoveTag(spaceID string, id string, tag string) error
	ChangeStage(spaceID, userID string, applicantID, stageID string) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store:               applicantstore.NewInstance(db.DB),
		selectionStageStore: selectionstagestore.NewInstance(db.DB),
		vacancyProvider:     vacancyhandler.Instance,
	}
}

type impl struct {
	store               applicantstore.Provider
	selectionStageStore selectionstagestore.Provider
	vacancyProvider     vacancyhandler.Provider
}

func (i impl) ListOfNegotiation(spaceID string, filter dbmodels.NegotiationFilter) ([]negotiationapimodels.NegotiationView, error) {
	list, err := i.store.ListOfNegotiation(spaceID, filter)
	if err != nil {
		log.WithField("filter", fmt.Sprintf("%+v", filter)).
			WithError(err).Error("ошибка получения списка откликов")
		return nil, errors.New("ошибка получения списка откликов")
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
	logger := log.
		WithField("space_id", spaceID).
		WithField("rec_id", id)
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		logger.WithError(err).Error("ошибка получения отклика")
		return negotiationapimodels.NegotiationView{}, errors.New("ошибка получения отклика")
	}
	if rec == nil {
		return negotiationapimodels.NegotiationView{}, errors.New("отклик не найден")
	}
	return negotiationapimodels.NegotiationConvertExt(*rec), nil
}

func (i impl) CreateApplicant(spaceID string, data applicantapimodels.ApplicantData) (id string, err error) {
	logger := log.WithField("space_id", spaceID)
	vacancy, err := i.checkDependency(spaceID, data)
	if err != nil {
		return "", err
	}
	rec := dbmodels.Applicant{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		VacancyID:             data.VacancyID,
		NegotiationID:         "",
		ResumeID:              "",
		ResumeTitle:           "",
		Source:                models.ApplicantSourceManual,
		NegotiationDate:       time.Time{},
		NegotiationAcceptDate: time.Now(),
		Status:                models.ApplicantStatusInProcess,
		FirstName:             data.FirstName,
		LastName:              data.LastName,
		MiddleName:            data.MiddleName,
		Salary:                data.Salary,
		Address:               data.Address,
		Citizenship:           data.Citizenship,
		Gender:                data.Gender,
		Relocation:            data.Relocation,
		Phone:                 data.Phone,
		Email:                 data.Email,
		TotalExperience:       data.TotalExperience,
		Params:                data.Params,
		Comment:               data.Comment,
	}
	birthDate, err := data.GetBirthDate()
	if err != nil {
		logger.WithError(err).Error("ошибка получения даты рождения кандидата")
		return "", errors.New("ошибка получения даты рождения кандидата")
	}
	rec.BirthDate = birthDate
	for _, stage := range vacancy.SelectionStages {
		if stage.Name == dbmodels.AddedStage {
			rec.SelectionStageID = stage.ID
			break
		}
	}
	recID, err := i.store.Create(rec)
	if err != nil {
		logger.
			WithField("request", fmt.Sprintf("%+v", data)).
			WithError(err).
			Error("ошибка создания кандидата")
		return "", errors.New("Ошибка создания кандидата")
	}
	logger.
		WithField("rec_id", recID).
		Info("Создан кандидат")
	return recID, nil

}

func (i impl) GetApplicant(spaceID string, id string) (applicantapimodels.ApplicantViewExt, error) {
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return applicantapimodels.ApplicantViewExt{}, err
	}
	if rec == nil {
		return applicantapimodels.ApplicantViewExt{}, errors.New("кандидат не найден")
	}
	//todo дубликаты
	result := applicantapimodels.ApplicantViewExt{
		ApplicantView: applicantapimodels.ApplicantConvert(rec.Applicant),
		Tags:          rec.Tags,
	}
	return result, nil
}

func (i impl) ListOfApplicant(spaceID string, filter applicantapimodels.ApplicantFilter) (list []applicantapimodels.ApplicantView, rowCount int64, err error) {
	logger := log.WithField("space_id", spaceID)
	rowCount, err = i.store.ListCountOfApplicant(spaceID, filter)
	if err != nil {
		return nil, 0, err
	}

	page, limit := filter.GetPage()
	offset := (page - 1) * limit
	if int64(offset) > rowCount {
		return []applicantapimodels.ApplicantView{}, rowCount, nil
	}

	recList, err := i.store.ListOfApplicant(spaceID, filter)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения списка кандидатов")
		return nil, 0, errors.New("ошибка получения списка кандидатов")
	}
	result := make([]applicantapimodels.ApplicantView, 0, len(recList))
	for _, rec := range recList {
		result = append(result, applicantapimodels.ApplicantConvert(rec))
	}
	return result, rowCount, nil
}

func (i impl) UpdateApplicant(spaceID string, id string, data applicantapimodels.ApplicantData) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	vacancy, err := i.checkDependency(spaceID, data)
	if err != nil {
		return err
	}

	birthDate, err := data.GetBirthDate()
	if err != nil {
		logger.WithError(err).Error("некорректный формат даты рождения кандидата")
		return errors.New("Некорректный формат даты рождения кандидата")
	}
	updMap := map[string]interface{}{
		"SpaceID":         spaceID,
		"VacancyID":       data.VacancyID,
		"Source":          models.ApplicantSourceManual,
		"Status":          models.ApplicantStatusInProcess,
		"FirstName":       data.FirstName,
		"LastName":        data.LastName,
		"MiddleName":      data.MiddleName,
		"Salary":          data.Salary,
		"Address":         data.Address,
		"BirthDate":       birthDate,
		"Citizenship":     data.Citizenship,
		"Gender":          data.Gender,
		"Relocation":      data.Relocation,
		"Phone":           data.Phone,
		"Email":           data.Email,
		"TotalExperience": data.TotalExperience,
		"Params":          data.Params,
		"Comment":         data.Comment,
	}
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		logger.WithError(err).Error("ошибка получения кандидата")
		return errors.New("ошибка получения кандидата")
	}
	if rec == nil {
		return errors.New("кандидат не найден")
	}
	//сменили вакансию, ищем такой же шаг
	if rec.VacancyID != data.VacancyID {
		currentStageName := ""
		if rec.SelectionStage != nil {
			currentStageName = rec.SelectionStage.Name
		}
		newSelectionStageID := ""
		for _, stage := range vacancy.SelectionStages {
			if stage.Name == currentStageName {
				newSelectionStageID = stage.ID
				break
			}
		}
		if newSelectionStageID == "" {
			return errors.New("смена вакансии невозможна, не найден этап подбора")
		}
		updMap["SelectionStageID"] = newSelectionStageID
	}

	err = i.store.Update(id, updMap)
	if err != nil {
		logger.
			WithField("request", fmt.Sprintf("%+v", data)).
			WithError(err).
			Error("ошибка обновления кандидата")
		return errors.New("ошибка обновления кандидата")
	}
	logger.Info("Обновлен кандидат")
	return nil
}

func (i impl) ApplicantAddTag(spaceID string, id string, tag string) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id).
		WithField("tag", tag)
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("кандидат не найден")
	}
	for _, recTag := range rec.Tags {
		if recTag == tag {
			//уже существует
			return nil
		}
	}
	tags := append(rec.Tags, tag)
	updMap := map[string]interface{}{
		"tags": pq.Array(tags),
	}
	err = i.store.Update(id, updMap)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка добавления тега кандидата")
		return errors.New("ошибка добавления тега кандидата")
	}
	logger.Info("кандидату добавлен тег")
	return nil
}

func (i impl) ApplicantRemoveTag(spaceID string, id string, tag string) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("кандидат не найден")
	}
	tags := make([]string, 0, len(rec.Tags)-1)
	for _, recTag := range rec.Tags {
		if recTag == tag {
			continue
		}
		tags = append(tags, recTag)
	}
	if len(tags) == len(rec.Tags) {
		// тэг не найден
		return nil
	}
	updMap := map[string]interface{}{
		"tags": pq.Array(tags),
	}
	err = i.store.Update(id, updMap)
	if err != nil {
		logger.
			WithField("tag", tag).
			WithError(err).
			Error("ошибка удаления тега кандидата")
		return errors.New("ошибка удаления тега кандидата")
	}
	logger.Info("удаленин тег у кандидата")
	return nil
}

func (i impl) ChangeStage(spaceID, userID string, applicantID, stageID string) error {
	logger := log.WithField("space_id", spaceID).
		WithField("applicant_id", applicantID).
		WithField("stage_id", stageID)
	applicantRec, err := i.store.GetByID(spaceID, applicantID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения данных кандидата")
		return errors.New("ошибка получения данных кандидата")
	}
	if applicantRec.Status != models.ApplicantStatusInProcess {
		return errors.Errorf("перевода по этапам возможен только на статусе '%v'", models.ApplicantStatusInProcess)
	}
	stageRec, err := i.selectionStageStore.GetByID(spaceID, applicantRec.VacancyID, stageID)
	if err != nil {
		return err
	}

	updMap := map[string]interface{}{
		"selection_stage_id": stageRec.ID,
	}

	switch stageRec.Name {
	case dbmodels.NegotiationStage:
		return errors.Errorf("перевод кандидата с текущего этапа на этап '%v' невозможен", stageRec.Name)
	case dbmodels.AddedStage:
		return errors.Errorf("перевод кандидата с текущего этапа на этап '%v' невозможен", stageRec.Name)
	case dbmodels.ScreenStage:
	case dbmodels.ManagerInterviewStage:
	case dbmodels.ClientInterviewStage:
	case dbmodels.OfferStage:
		break
	case dbmodels.HiredStage:
		updMap["start_date"] = time.Now()
	}
	err = i.store.Update(applicantID, updMap)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка изменения этапа подбора для кандидата")
		return errors.New("ошибка изменения этапа подбора для кандидата")
	}
	// TODO обновление истории
	return nil
}

func (i impl) checkDependency(spaceID string, data applicantapimodels.ApplicantData) (vacancy vacancyapimodels.VacancyView, err error) {
	if data.VacancyID == "" {
		return vacancyapimodels.VacancyView{}, errors.New("необходима указать вакансию")
	}
	vacancy, err = i.vacancyProvider.GetByID(spaceID, data.VacancyID)
	if err != nil {
		return vacancyapimodels.VacancyView{}, err
	}

	return vacancy, nil
}
