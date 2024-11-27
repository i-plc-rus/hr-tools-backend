package applicant

import (
	"fmt"
	"hr-tools-backend/db"
	applicanthistoryhandler "hr-tools-backend/lib/applicant-history"
	applicantstore "hr-tools-backend/lib/applicant/store"
	vacancyhandler "hr-tools-backend/lib/vacancy"
	selectionstagestore "hr-tools-backend/lib/vacancy/selection-stage-store"
	"hr-tools-backend/models"
	applicantapimodels "hr-tools-backend/models/api/applicant"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Provider interface {
	ListOfNegotiation(spaceID string, filter dbmodels.NegotiationFilter) (list []negotiationapimodels.NegotiationView, err error)
	UpdateComment(spaceID, id, userID string, comment string) error
	UpdateStatus(spaceID, id, userID string, status models.NegotiationStatus) error
	GetByID(spaceID, id string) (negotiationapimodels.NegotiationView, error)
	CreateApplicant(spaceID, userID string, applicant applicantapimodels.ApplicantData) (string, error)
	GetApplicant(spaceID string, id string) (applicantapimodels.ApplicantViewExt, error)
	ListOfApplicant(spaceID string, filter applicantapimodels.ApplicantFilter) (list []applicantapimodels.ApplicantView, rowCount int64, err error)
	UpdateApplicant(spaceID string, id, userID string, applicant applicantapimodels.ApplicantData) error
	ApplicantAddTag(spaceID string, id, userID string, tag string) error
	ApplicantRemoveTag(spaceID string, id, userID string, tag string) error
	ChangeStage(spaceID, userID string, applicantID, stageID string) error
	ResolveDuplicate(spaceID string, mainID, minorID, userID string, isDuplicate bool) error
	ApplicantReject(spaceID string, id, userID string, reason string) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store:               applicantstore.NewInstance(db.DB),
		selectionStageStore: selectionstagestore.NewInstance(db.DB),
		vacancyProvider:     vacancyhandler.Instance,
		applicantHistory:    applicanthistoryhandler.Instance,
	}
}

type impl struct {
	store               applicantstore.Provider
	selectionStageStore selectionstagestore.Provider
	vacancyProvider     vacancyhandler.Provider
	applicantHistory    applicanthistoryhandler.Provider
}

func (i *impl) getLogger(spaceID, applicantID, userID string) *log.Entry {
	logger := log.WithField("space_id", spaceID)
	if applicantID != "" {
		logger = logger.WithField("applicant_id", applicantID)
	}
	if userID != "" {
		logger = logger.WithField("user_id", userID)
	}
	return logger
}

func (i impl) ListOfNegotiation(spaceID string, filter dbmodels.NegotiationFilter) ([]negotiationapimodels.NegotiationView, error) {
	logger := i.getLogger(spaceID, "", "")
	list, err := i.store.ListOfNegotiation(spaceID, filter)
	if err != nil {
		logger.WithField("filter", fmt.Sprintf("%+v", filter)).
			WithError(err).Error("ошибка получения списка откликов")
		return nil, errors.New("ошибка получения списка откликов")
	}
	result := make([]negotiationapimodels.NegotiationView, 0, len(list))
	for _, rec := range list {
		result = append(result, negotiationapimodels.NegotiationConvert(rec))
	}
	return result, nil
}

func (i impl) UpdateComment(spaceID, id, userID string, comment string) error {
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("запись не найдена")
	}
	updMap := map[string]interface{}{
		"comment": comment,
	}
	err = i.store.Update(id, updMap)
	if err != nil {
		return err
	}
	changes := applicanthistoryhandler.GetUpdateChanges("Изменен профиль", rec.Applicant, updMap)
	i.applicantHistory.Save(rec.SpaceID, rec.ID, rec.VacancyID, userID, dbmodels.HistoryTypeUpdate, changes)
	return nil
}

func (i impl) UpdateStatus(spaceID, id, userID string, status models.NegotiationStatus) error {
	logger := i.getLogger(spaceID, id, userID)
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
	changeMsg := fmt.Sprintf("Перевод отклика кандидата на статус %v", status)
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
		changeMsg = "Кандидат из отклика, добавлен на вакансию"
	}
	if status == models.NegotiationStatusRejected {
		updMap["negotiation_accept_date"] = time.Now()
		changeMsg = "Отклик кандидата отклонен"
	}
	err = i.store.Update(id, updMap)
	if err != nil {
		logger.WithError(err).Error("ошибка обновления кандидата")
		return errors.New("ошибка обновления кандидата")
	}
	changes := applicanthistoryhandler.GetUpdateChanges(changeMsg, rec.Applicant, updMap)
	i.applicantHistory.Save(rec.SpaceID, id, rec.VacancyID, userID, dbmodels.HistoryTypeUpdate, changes)
	return nil
}

func (i impl) GetByID(spaceID, id string) (negotiationapimodels.NegotiationView, error) {
	logger := i.getLogger(spaceID, id, "")
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

func (i impl) CreateApplicant(spaceID, userID string, data applicantapimodels.ApplicantData) (id string, err error) {
	logger := i.getLogger(spaceID, "", userID)
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
	changes := applicanthistoryhandler.GetCreateChanges("Кандидат добавлен на вакансию", rec)
	i.applicantHistory.Save(rec.SpaceID, recID, rec.VacancyID, userID, dbmodels.HistoryTypeAdded, changes)
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
	result := applicantapimodels.ApplicantViewExt{
		ApplicantView: applicantapimodels.ApplicantConvert(rec.Applicant),
		Tags:          rec.Tags,
	}
	result.Duplicates = make([]string, 0, len(rec.Duplicates))
	for _, item := range rec.Duplicates {
		result.Duplicates = append(result.Duplicates, item.ID)
	}
	if rec.Status != models.ApplicantStatusArchive {
		result.PotentialDuplicate = i.checkDuplicate(rec)
	}
	return result, nil
}

func (i impl) ListOfApplicant(spaceID string, filter applicantapimodels.ApplicantFilter) (list []applicantapimodels.ApplicantView, rowCount int64, err error) {
	logger := i.getLogger(spaceID, "", "")
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

func (i impl) UpdateApplicant(spaceID string, id, userID string, data applicantapimodels.ApplicantData) error {
	logger := i.getLogger(spaceID, id, userID)
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
	if rec.Status == models.ApplicantStatusArchive {
		return errors.Errorf("обновление данных кандидата в статусе '%v' - недоступно", models.ApplicantStatusArchive)
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
	changes := applicanthistoryhandler.GetUpdateChanges("Изменен профиль", rec.Applicant, updMap)
	if len(changes.Data) != 0 {
		i.applicantHistory.Save(rec.SpaceID, id, rec.VacancyID, userID, dbmodels.HistoryTypeUpdate, changes)
	}
	logger.Info("Обновлен кандидат")
	return nil
}

func (i impl) ApplicantAddTag(spaceID string, id, userID string, tag string) error {
	logger := i.getLogger(spaceID, id, userID).WithField("tag", tag)
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
	changes := applicanthistoryhandler.GetUpdateChanges("Добавлен тег", rec.Applicant, updMap)
	i.applicantHistory.Save(rec.SpaceID, id, rec.VacancyID, userID, dbmodels.HistoryTypeUpdate, changes)
	logger.Info("кандидату добавлен тег")
	return nil
}

func (i impl) ApplicantRemoveTag(spaceID string, id, userID string, tag string) error {
	logger := i.getLogger(spaceID, id, userID).WithField("tag", tag)
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
			WithError(err).
			Error("ошибка удаления тега кандидата")
		return errors.New("ошибка удаления тега кандидата")
	}
	changes := applicanthistoryhandler.GetUpdateChanges("Удален тег", rec.Applicant, updMap)
	i.applicantHistory.Save(rec.SpaceID, id, rec.VacancyID, userID, dbmodels.HistoryTypeUpdate, changes)
	logger.Info("удален тег у кандидата")
	return nil
}

func (i impl) ChangeStage(spaceID, userID string, applicantID, stageID string) error {
	logger := i.getLogger(spaceID, applicantID, userID).
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
	if stageRec == nil {
		return errors.New("этапам не найден")
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
	changes := applicanthistoryhandler.GetStageChange(stageRec.Name)
	i.applicantHistory.Save(spaceID, applicantID, applicantRec.VacancyID, userID, dbmodels.HistoryTypeStageChange, changes)
	return nil
}

func (i impl) ResolveDuplicate(spaceID string, mainID, minorID, userID string, isDuplicate bool) error {
	logger := i.getLogger(spaceID, "", userID).
		WithField("main_id", mainID).
		WithField("minor_ID", mainID)
	if isDuplicate {
		return i.joinApplicants(spaceID, mainID, minorID, userID, logger)
	}
	return i.markAsDifferentApplicants(spaceID, mainID, minorID, userID, logger)
}

func (i impl) ApplicantReject(spaceID string, id, userID string, reason string) error {
	logger := i.getLogger(spaceID, id, userID)
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("кандидат не найден")
	}
	if rec.Status == models.ApplicantStatusRejected {
		return nil
	}
	updMap := map[string]interface{}{
		"status":        models.ApplicantStatusRejected,
		"reject_reason": reason,
	}
	err = i.store.Update(id, updMap)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка перевода кандидата в отклоненные")
		return errors.New("ошибка перевода кандидата в отклоненные")
	}
	changes := applicanthistoryhandler.GetRejectChange(reason, rec.Applicant, updMap)
	i.applicantHistory.Save(spaceID, id, rec.VacancyID, userID, dbmodels.HistoryTypeReject, changes)
	return nil
}

func (i impl) markAsDifferentApplicants(spaceID string, mainID, minorID, userID string, logger *log.Entry) error {
	mainRec, err := i.store.GetByID(spaceID, mainID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения данных кандидата")
		return errors.New("ошибка получения данных кандидата")
	}
	if mainRec == nil {
		return errors.New("запись с кандидатом не найдена")
	}
	for _, notDuplicateID := range mainRec.NotDuplicates {
		if notDuplicateID == minorID {
			logger.Info("признак разных кандидатов уже установлен")
			return nil
		}
	}
	minorRec, err := i.store.GetByID(spaceID, minorID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения данных кандидата дубликата")
		return errors.New("ошибка получения данных кандидата дубликата")
	}
	if minorRec == nil {
		return errors.New("запись с дубликатом кандидата не найдена")
	}

	notDuplicates := append(mainRec.NotDuplicates, minorID)
	updMap := map[string]interface{}{
		"not_duplicates": pq.Array(notDuplicates),
	}
	err = i.store.Update(mainID, updMap)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка установки признака разных кандидатов")
		return errors.New("ошибка установки признака разных кандидатов")
	}
	descr := applicanthistoryhandler.GetNotDuplicateMark(minorRec.Applicant)
	i.applicantHistory.Save(spaceID, mainID, mainRec.VacancyID, userID, dbmodels.HistoryTypeDuplicate, descr)
	logger.Info("установлен признак разных кандидатов")
	return nil
}

func (i impl) joinApplicants(spaceID string, mainID, minorID, userID string, logger *log.Entry) error {
	mainRec, err := i.store.GetByID(spaceID, mainID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения данных основного кандидата")
		return errors.New("ошибка получения данных основного кандидата")
	}
	if mainRec == nil {
		return errors.New("запись с основным кандидатом не найдена")
	}
	if mainRec.Status == models.ApplicantStatusArchive {
		return errors.Errorf("объединение данных кандидата в статусе '%v' - недоступно", models.ApplicantStatusArchive)
	}
	minorRec, err := i.store.GetByID(spaceID, minorID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения данных кандидата дубликата")
		return errors.New("ошибка получения данных кандидата дубликата")
	}
	if minorRec == nil {
		return errors.New("запись с дубликатом кандидата не найдена")
	}
	if minorRec.Status == models.ApplicantStatusArchive {
		return errors.Errorf("объединение данных кандидата в статусе '%v' - недоступно", models.ApplicantStatusArchive)
	}
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		store := applicantstore.NewInstance(tx)
		updMap := map[string]interface{}{
			"status":       models.ApplicantStatusArchive,
			"duplicate_id": mainID,
		}
		err = store.Update(minorID, updMap)
		if err != nil {
			logger.
				WithError(err).
				Error("ошибка перевода дубликата в архив")
			return errors.New("ошибка перевода дубликата в архив")
		}
		return nil
	})
	if err != nil {
		return err
	}
	descr := applicanthistoryhandler.GetDuplicateMark(minorRec.Applicant)
	i.applicantHistory.Save(spaceID, mainID, mainRec.VacancyID, userID, dbmodels.HistoryTypeDuplicate, descr)
	logger.Info("дубликат кандидата перемещем в архив")
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

func (i impl) checkDuplicate(originRec *dbmodels.ApplicantExt) applicantapimodels.ApplicantDuplicate {
	logger := i.getLogger(originRec.SpaceID, originRec.ID, "")
	applicantFIO := getLowerFio(originRec.Applicant)
	filter := dbmodels.DuplicateApplicantFilter{
		VacancyID:      originRec.VacancyID,
		FIO:            applicantFIO,
		Phone:          originRec.Phone,
		Email:          originRec.Email,
		ExtApplicantID: originRec.ExtApplicantID,
	}
	list, err := i.store.ListOfDuplicateApplicant(originRec.SpaceID, filter)
	if err != nil {
		logger.WithError(err).Error("Ошибка получения списка кандидатов для поиска дублей")
		return applicantapimodels.ApplicantDuplicate{}
	}
	if len(list) <= 1 {
		return applicantapimodels.ApplicantDuplicate{}
	}

	for _, rec := range list {
		if rec.ID == originRec.ID {
			continue
		}
		if rec.DuplicateID != nil && *rec.DuplicateID == originRec.ID {
			//уже помечен как дубль
			continue
		}

		if originRec.IsMarkAsNotDuplicate(rec) {
			//уже помечен как не дубль
			continue
		}
		if originRec.ExtApplicantID != "" && originRec.ExtApplicantID == rec.ExtApplicantID {
			//совпадение по автору во внешней системе
			return applicantapimodels.ApplicantDuplicate{
				Found:         true,
				DuplicateID:   rec.ID,
				DuplicateType: models.DuplicateTypeByAuthor,
			}
		}
		if applicantFIO == "" || applicantFIO != getLowerFio(rec) {
			continue
		}
		phoneEquals := false
		if originRec.Phone != "" && originRec.Phone == rec.Phone {
			if originRec.Email == "" || rec.Email == "" {
				//ФИО+телефон если почта не указанна
				return applicantapimodels.ApplicantDuplicate{
					Found:         true,
					DuplicateID:   rec.ID,
					DuplicateType: models.DuplicateTypeByContacts,
				}
			}
			phoneEquals = true
		}
		if originRec.Email != "" && originRec.Email == rec.Email {
			if phoneEquals || (originRec.Phone == "" || rec.Phone == "") {
				//ФИО, почта, телефон или ФИО+почта если телефон не указанна
				return applicantapimodels.ApplicantDuplicate{
					Found:         true,
					DuplicateID:   rec.ID,
					DuplicateType: models.DuplicateTypeByContacts,
				}
			}
		}
	}
	return applicantapimodels.ApplicantDuplicate{}
}

func getLowerFio(rec dbmodels.Applicant) string {
	return fmt.Sprintf("%v %v %v", rec.LastName, rec.FirstName, rec.MiddleName)
}
