package applicant

import (
	"bytes"
	"fmt"
	"hr-tools-backend/db"
	applicanthistoryhandler "hr-tools-backend/lib/applicant-history"
	applicantstore "hr-tools-backend/lib/applicant/store"
	xlsexport "hr-tools-backend/lib/export/xls"
	pushhandler "hr-tools-backend/lib/space/push/handler"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	vacancyhandler "hr-tools-backend/lib/vacancy"
	selectionstagestore "hr-tools-backend/lib/vacancy/selection-stage-store"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	"hr-tools-backend/models"
	applicantapimodels "hr-tools-backend/models/api/applicant"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
	"math"
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
	MultiChangeStage(spaceID, userID string, data applicantapimodels.MultiChangeStageRequest) error
	ResolveDuplicate(spaceID string, mainID, minorID, userID string, isDuplicate bool) error
	ApplicantReject(spaceID string, id, userID string, data applicantapimodels.RejectRequest) error
	ApplicantMultiReject(spaceID string, userID string, data applicantapimodels.MultiRejectRequest) error
	ExportToXls(spaceID string, data applicantapimodels.XlsExportRequest) (*bytes.Buffer, error)
	ListOfSource(spaceID string, filter applicantapimodels.ApplicantFilter) (data applicantapimodels.ApplicantSourceData, err error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store:               applicantstore.NewInstance(db.DB),
		selectionStageStore: selectionstagestore.NewInstance(db.DB),
		vacancyProvider:     vacancyhandler.Instance,
		applicantHistory:    applicanthistoryhandler.Instance,
		userStore:           spaceusersstore.NewInstance(db.DB),
		vacancyStore:        vacancystore.NewInstance(db.DB),
	}
}

type impl struct {
	store               applicantstore.Provider
	selectionStageStore selectionstagestore.Provider
	vacancyProvider     vacancyhandler.Provider
	applicantHistory    applicanthistoryhandler.Provider
	userStore           spaceusersstore.Provider
	vacancyStore        vacancystore.Provider
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
	list, err := i.store.ListOfNegotiation(spaceID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения списка откликов")
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

	go func(rec dbmodels.Applicant, userID string) {
		logger := i.getLogger(rec.SpaceID, rec.ID, userID).
			WithField("event_code", models.PushApplicantNote)
		userRec := i.getUser(userID, logger)
		if userRec == nil {
			return
		}
		vacancyRec := i.getVacancy(rec.SpaceID, rec.VacancyID, logger)
		if vacancyRec == nil {
			return
		}
		notification := models.GetPushApplicantNote(vacancyRec.VacancyName, rec.GetFIO(), userRec.GetFullName())
		i.sendNotification(*vacancyRec, notification)

	}(rec.Applicant, userID)
	return nil
}

func (i impl) UpdateStatus(spaceID, id, userID string, status models.NegotiationStatus) error {
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
			return errors.Wrap(err, "ошибка получения списка этапов подбора")
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
		return errors.Wrap(err, "ошибка обновления кандидата")
	}
	changes := applicanthistoryhandler.GetUpdateChanges(changeMsg, rec.Applicant, updMap)
	i.applicantHistory.Save(rec.SpaceID, id, rec.VacancyID, userID, dbmodels.HistoryTypeUpdate, changes)
	return nil
}

func (i impl) GetByID(spaceID, id string) (negotiationapimodels.NegotiationView, error) {
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return negotiationapimodels.NegotiationView{}, errors.Wrap(err, "ошибка получения отклика")
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
		Source:                data.Source,
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
		return "", errors.Wrap(err, "ошибка получения даты рождения кандидата")
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
		return "", err
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
		return nil, 0, err
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
		return errors.Wrap(err, "Некорректный формат даты рождения кандидата")
	}
	updMap := map[string]interface{}{
		"SpaceID":         spaceID,
		"VacancyID":       data.VacancyID,
		"Source":          data.Source,
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
		return errors.Wrap(err, "ошибка получения кандидата")
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
		return errors.Wrap(err, "ошибка обновления кандидата")
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
		return errors.Wrap(err, "ошибка добавления тега кандидата")
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
		return errors.Wrap(err, "ошибка удаления тега кандидата")
	}
	changes := applicanthistoryhandler.GetUpdateChanges("Удален тег", rec.Applicant, updMap)
	i.applicantHistory.Save(rec.SpaceID, id, rec.VacancyID, userID, dbmodels.HistoryTypeUpdate, changes)
	logger.Info("удален тег у кандидата")
	return nil
}

func (i impl) ChangeStage(spaceID, userID string, applicantID, stageID string) error {
	userName, err := i.getUserName(userID)
	if err != nil {
		return errors.Wrap(err, "ошибка перевода кандидата на этап")
	}
	err = i.сhangeStage(nil, spaceID, userID, userName, applicantID, stageID)
	if err != nil {
		return err
	}
	return nil
}

func (i impl) MultiChangeStage(spaceID, userID string, data applicantapimodels.MultiChangeStageRequest) error {
	userName, err := i.getUserName(userID)
	if err != nil {
		return errors.Wrap(err, "ошибка перевода кандидата на этап")
	}
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		for _, id := range data.IDs {
			err = i.сhangeStage(nil, spaceID, userID, userName, id, data.StageID)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
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

func (i impl) ApplicantReject(spaceID string, id, userID string, data applicantapimodels.RejectRequest) error {
	userName, err := i.getUserName(userID)
	if err != nil {
		return errors.Wrap(err, "ошибка перевода кандидата в отклоненные")
	}
	err = i.applicantReject(nil, spaceID, id, userID, userName, data)
	if err != nil {
		return err
	}
	return nil
}

func (i impl) ApplicantMultiReject(spaceID string, userID string, data applicantapimodels.MultiRejectRequest) error {
	userName, err := i.getUserName(userID)
	if err != nil {
		return errors.Wrap(err, "ошибка перевода кандидатов в отклоненные")
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		for _, id := range data.IDs {
			err := i.applicantReject(tx, spaceID, id, userID, userName, data.Reject)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (i impl) ExportToXls(spaceID string, data applicantapimodels.XlsExportRequest) (*bytes.Buffer, error) {
	list, err := i.store.ListOfApplicantByIDs(spaceID, data.IDs, data.Filter)
	if err != nil {
		return nil, err
	}
	return xlsexport.Instance.ExportApplicantList(list)
}

func (i impl) ListOfSource(spaceID string, filter applicantapimodels.ApplicantFilter) (data applicantapimodels.ApplicantSourceData, err error) {
	logger := i.getLogger(spaceID, "", "")
	recList, err := i.store.ListOfApplicantSource(spaceID, filter)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения аналитики по источникам кандидатов")
		return applicantapimodels.ApplicantSourceData{}, errors.New("ошибка получения аналитики по источникам кандидатов")
	}

	result := applicantapimodels.ApplicantSourceData{
		NegotiationSource: applicantapimodels.SourceData{Data: []applicantapimodels.SourceItem{}},
		AddingSource:      applicantapimodels.SourceData{Data: []applicantapimodels.SourceItem{}},
	}
	for _, rec := range recList {
		if rec.IsNegotiation {
			result.NegotiationSource.Total += rec.Total

			result.NegotiationSource.Data = append(result.NegotiationSource.Data,
				applicantapimodels.SourceItem{
					Name:  string(rec.Source),
					Count: rec.Total,
				})
		} else {
			result.AddingSource.Total += rec.Total
			result.AddingSource.Data = append(result.AddingSource.Data,
				applicantapimodels.SourceItem{
					Name:  string(rec.Source),
					Count: rec.Total,
				})
		}
	}
	result.NegotiationSource.Data = calcPercent(result.NegotiationSource.Data, result.NegotiationSource.Total)
	result.AddingSource.Data = calcPercent(result.AddingSource.Data, result.AddingSource.Total)
	result.TotalSource = applicantapimodels.SourceData{
		Total: result.NegotiationSource.Total + result.AddingSource.Total,
		Data: []applicantapimodels.SourceItem{
			{Name: "Откликнулись", Count: result.NegotiationSource.Total},
			{Name: "Добавлены", Count: result.AddingSource.Total},
		}}
	result.TotalSource.Data = calcPercent(result.TotalSource.Data, result.TotalSource.Total)
	return result, nil
}

func (i impl) markAsDifferentApplicants(spaceID string, mainID, minorID, userID string, logger *log.Entry) error {
	mainRec, err := i.store.GetByID(spaceID, mainID)
	if err != nil {
		return errors.Wrap(err, "ошибка получения данных кандидата")
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
		return errors.Wrap(err, "ошибка получения данных кандидата дубликата")
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
		return errors.Wrap(err, "ошибка установки признака разных кандидатов")
	}
	descr := applicanthistoryhandler.GetNotDuplicateMark(minorRec.Applicant)
	i.applicantHistory.Save(spaceID, mainID, mainRec.VacancyID, userID, dbmodels.HistoryTypeDuplicate, descr)
	logger.Info("установлен признак разных кандидатов")
	return nil
}

func (i impl) joinApplicants(spaceID string, mainID, minorID, userID string, logger *log.Entry) error {
	mainRec, err := i.store.GetByID(spaceID, mainID)
	if err != nil {
		return errors.Wrap(err, "ошибка получения данных основного кандидата")
	}
	if mainRec == nil {
		return errors.New("запись с основным кандидатом не найдена")
	}
	if mainRec.Status == models.ApplicantStatusArchive {
		return errors.Errorf("объединение данных кандидата в статусе '%v' - недоступно", models.ApplicantStatusArchive)
	}
	minorRec, err := i.store.GetByID(spaceID, minorID)
	if err != nil {
		return errors.Wrap(err, "ошибка получения данных кандидата дубликата")
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
			return errors.Wrap(err, "ошибка перевода дубликата в архив")
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

func (i impl) applicantReject(tx *gorm.DB, spaceID string, id, userID, userName string, data applicantapimodels.RejectRequest) error {
	logger := i.getLogger(spaceID, id, userID)
	store := i.store
	applicantHistory := i.applicantHistory
	if tx != nil {
		store = applicantstore.NewInstance(tx)
		applicantHistory = applicanthistoryhandler.NewTxHandler(tx)
	}

	rec, err := store.GetByID(spaceID, id)
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
		"status":           models.ApplicantStatusRejected,
		"reject_reason":    data.Reason,
		"reject_initiator": data.Initiator,
	}
	err = store.Update(id, updMap)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка перевода кандидата в отклоненные")
		return errors.New("ошибка перевода кандидата в отклоненные")
	}

	// История изменений
	changes := applicanthistoryhandler.GetRejectChange(data.Reason, rec.Applicant, updMap)
	applicantHistory.SaveWithUser(spaceID, id, rec.VacancyID, userID, userName, dbmodels.HistoryTypeReject, changes)
	return err
}

func (i impl) getUserName(userID string) (string, error) {
	if userID != "" {
		user, err := i.userStore.GetByID(userID)
		if err != nil {
			return "", errors.Wrap(err, "не удалось получить автора изменений")
		}
		if user == nil {
			return "", errors.New("автор изменений не найден")
		}
		return user.GetFullName(), nil
	}
	return models.SystemUser, nil
}

func (i impl) сhangeStage(tx *gorm.DB, spaceID, userID, userName string, applicantID, stageID string) error {
	logger := i.getLogger(spaceID, applicantID, userID).
		WithField("stage_id", stageID)
	store := i.store
	applicantHistory := i.applicantHistory
	selectionStageStore := i.selectionStageStore
	if tx != nil {
		store = applicantstore.NewInstance(tx)
		applicantHistory = applicanthistoryhandler.NewTxHandler(tx)
		selectionStageStore = selectionstagestore.NewInstance(tx)
	}
	applicantRec, err := store.GetByID(spaceID, applicantID)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения данных кандидата")
		return errors.New("ошибка получения данных кандидата")
	}
	if applicantRec.Status != models.ApplicantStatusInProcess {
		return errors.Errorf("перевода по этапам возможен только на статусе '%v'", models.ApplicantStatusInProcess)
	}
	stageRec, err := selectionStageStore.GetByID(spaceID, applicantRec.VacancyID, stageID)
	if err != nil {
		return err
	}
	if stageRec == nil {
		return errors.New("этап не найден")
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
	err = store.Update(applicantID, updMap)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка изменения этапа подбора для кандидата")
		return errors.New("ошибка изменения этапа подбора для кандидата")
	}
	changes := applicanthistoryhandler.GetStageChange(stageRec.Name)
	applicantHistory.SaveWithUser(spaceID, applicantID, applicantRec.VacancyID, userID, userName, dbmodels.HistoryTypeStageChange, changes)

	go func(rec dbmodels.Applicant, userName string) {
		logger := i.getLogger(rec.SpaceID, rec.ID, userID).
			WithField("event_code", models.PushApplicantNewStage)
		vacancyRec := i.getVacancy(rec.SpaceID, rec.VacancyID, logger)
		if vacancyRec == nil {
			return
		}
		notification := models.GetPushApplicantNewStage(vacancyRec.VacancyName, userName, stageRec.Name)
		i.sendNotification(*vacancyRec, notification)
	}(applicantRec.Applicant, userName)
	return nil
}

func (i *impl) sendNotification(rec dbmodels.Vacancy, data models.NotificationData) {
	//отправляем автору
	pushhandler.Instance.SendNotification(rec.AuthorID, data)
	for _, teamMember := range rec.VacancyTeam {
		//отправляем команде
		if rec.AuthorID == teamMember.ID {
			continue
		}
		pushhandler.Instance.SendNotification(teamMember.ID, data)
	}
}

func (i *impl) getVacancy(spaceID, vacancyID string, logger *log.Entry) *dbmodels.Vacancy {
	vacancyRec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		logger.WithError(err).Error("ошибка получения вакансии")
		return nil
	}
	if vacancyRec == nil {
		logger.Error("вакансия не найдена")
		return nil
	}
	return vacancyRec
}

func (i *impl) getUser(userID string, logger *log.Entry) *dbmodels.SpaceUser {
	userRec, err := i.userStore.GetByID(userID)
	if err != nil {
		logger.WithError(err).Error("ошибка получения пользователя")
		return nil
	}
	if userRec == nil {
		logger.Error("пользователь не найден")
		return nil
	}
	return userRec
}

func calcPercent(sourceData []applicantapimodels.SourceItem, sourceTotal int) []applicantapimodels.SourceItem {
	totalPercent := 0
	lastIdx := len(sourceData) - 1
	for idx, data := range sourceData {
		if idx == lastIdx {
			sourceData[idx].Percent = 100 - totalPercent
		} else {
			sourceData[idx].Percent = int(math.Round(float64(data.Count) / float64(sourceTotal) * 100))
			totalPercent += sourceData[idx].Percent
		}
	}
	return sourceData
}
func percent(count int, all int) int {
	return int(math.Round(float64(count) / float64(all) * 100))
}
