package applicantstore

import (
	"bytes"
	"fmt"
	"hr-tools-backend/models"
	applicantapimodels "hr-tools-backend/models/api/applicant"
	dbmodels "hr-tools-backend/models/db"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Provider interface {
	Create(data dbmodels.Applicant) (id string, err error)
	Update(id string, updMap map[string]interface{}) error
	GetByID(spaceID, id string) (rec *dbmodels.ApplicantExt, err error)
	IsExistNegotiationID(spaceID, negotiationID string, source models.ApplicantSource) (found bool, err error)
	ListOfNegotiation(spaceID string, filter dbmodels.NegotiationFilter) ([]dbmodels.Applicant, error)
	ListCountOfApplicant(spaceID string, filter applicantapimodels.ApplicantFilter) (count int64, err error)
	ListOfApplicant(spaceID string, filter applicantapimodels.ApplicantFilter) ([]dbmodels.Applicant, error)
	ListOfDuplicateApplicant(spaceID string, filter dbmodels.DuplicateApplicantFilter) (list []dbmodels.Applicant, err error)
	ApplicantsByStages(spaceID string, vacancyIDs []string) (list []dbmodels.ApplicantsStage, err error)
	ListOfApplicantByIDs(spaceID string, ids []string, filter *applicantapimodels.ApplicantFilter) ([]dbmodels.ApplicantWithJob, error)
	ListOfApplicantSource(spaceID string, filter applicantapimodels.ApplicantFilter) ([]dbmodels.ApplicantSource, error)
	ListOfActiveApplicants() ([]dbmodels.Applicant, error)
	ListOfActivefNegotiation(withHrSurvy bool) ([]dbmodels.Applicant, error)
	ListForSurveySend() ([]dbmodels.Applicant, error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.Applicant) (id string, err error) {
	err = i.db.Omit(clause.Associations).
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) Update(id string, updMap map[string]interface{}) error {
	if len(updMap) == 0 {
		return nil
	}
	tx := i.db.
		Model(&dbmodels.Applicant{}).
		Where("id = ?", id).
		Updates(updMap)
	if tx.RowsAffected == 0 {
		return errors.New("запись не найдена")
	}
	err := tx.Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) GetByID(spaceID, id string) (*dbmodels.ApplicantExt, error) {
	rec := dbmodels.ApplicantExt{}
	err := i.db.
		Select("applicants.*, v.author_id, s.first_name as author_first_name, s.last_name as author_last_name").
		Model(&dbmodels.Applicant{}).
		Joins("left join vacancies as v on vacancy_id = v.id").
		Joins("left join space_users as s on v.author_id = s.id").
		Where("applicants.id = ?", id).
		Where("applicants.space_id = ?", spaceID).
		Preload(clause.Associations).
		Preload("Vacancy.JobTitle").
		Preload("Vacancy.Space").
		First(&rec).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (i impl) IsExistNegotiationID(spaceID, negotiationID string, source models.ApplicantSource) (found bool, err error) {
	var exists bool
	err = i.db.Model(&dbmodels.Applicant{}).
		Select("count(*) > 0").
		Where("space_id = ?", spaceID).
		Where("negotiation_id = ? and source = ?", negotiationID, source).
		Find(&exists).
		Error
	return exists, err
}

func (i impl) ListOfNegotiation(spaceID string, filter dbmodels.NegotiationFilter) (list []dbmodels.Applicant, err error) {
	list = []dbmodels.Applicant{}
	tx := i.db.
		Model(dbmodels.Applicant{}).
		Where("space_id = ?", spaceID).
		Where("vacancy_id = ?", filter.VacancyID).
		Where("(negotiation_id is not null and negotiation_id <> '')").
		Where("status != ?", models.ApplicantStatusArchive)
	i.addNegotiationFilter(tx, filter)
	err = tx.Preload(clause.Associations).Find(&list).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) ListOfApplicant(spaceID string, filter applicantapimodels.ApplicantFilter) (list []dbmodels.Applicant, err error) {
	list = []dbmodels.Applicant{}
	tx := i.db.
		Select("applicants.*, (last_name || ' ' || first_name|| ' ' || middle_name) as fio").
		Model(dbmodels.Applicant{}).
		Where("applicants.space_id = ?", spaceID).
		Joins("left join vacancies as v on vacancy_id = v.id").
		Joins("left join selection_stages as st on selection_stage_id = st.id")
	i.addApplicantFilter(tx, filter)
	i.addSort(tx, filter.Sort)
	page, limit := filter.GetPage()
	i.setPage(tx, page, limit)
	err = tx.Preload(clause.Associations).
		Preload("Vacancy.JobTitle").
		Preload("Vacancy.Space").
		Find(&list).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) ListCountOfApplicant(spaceID string, filter applicantapimodels.ApplicantFilter) (count int64, err error) {
	var rowCount int64
	tx := i.db.
		Model(dbmodels.Applicant{}).
		Joins("left join vacancies as v on vacancy_id = v.id").
		Joins("left join selection_stages as st on selection_stage_id = st.id").
		Where("applicants.space_id = ?", spaceID)
	i.addApplicantFilter(tx, filter)
	err = tx.Count(&rowCount).Error
	if err != nil {
		log.WithError(err).Error("ошибка получения общего количества кандидатов")
		return 0, errors.New("ошибка получения общего количества кандидатов")
	}
	return rowCount, nil
}

func (i impl) ListOfDuplicateApplicant(spaceID string, filter dbmodels.DuplicateApplicantFilter) (list []dbmodels.Applicant, err error) {
	list = []dbmodels.Applicant{}
	tx := i.db.
		Model(dbmodels.Applicant{}).
		Where("space_id = ?", spaceID).
		Where("status != ?", models.ApplicantStatusArchive).
		Where("vacancy_id = ?", filter.VacancyID).
		Where("LOWER(last_name || ' ' || first_name|| ' ' || middle_name) = ?", strings.ToLower(filter.FIO))
	if filter.ExtApplicantID != "" {
		tx.Or("ext_applicant_id = ?", filter.ExtApplicantID)
	}
	if filter.Phone != "" {
		tx.Or("phone = ?", filter.Phone)
	}
	if filter.Email != "" {
		tx.Or("email = ?", filter.Email)
	}
	err = tx.Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) ApplicantsByStages(spaceID string, vacancyIDs []string) (list []dbmodels.ApplicantsStage, err error) {
	tx := i.db.
		Select("count(id) as total, vacancy_id, selection_stage_id").
		Model(dbmodels.Applicant{}).
		Group("vacancy_id, selection_stage_id").
		Where("vacancy_id in (?)", vacancyIDs)
		// Where("status in (?)", []models.ApplicantStatus{models.ApplicantStatusInProcess, models.ApplicantStatusNegotiation})

	err = tx.Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (i impl) ListOfApplicantByIDs(spaceID string, ids []string, filter *applicantapimodels.ApplicantFilter) (list []dbmodels.ApplicantWithJob, err error) {
	if len(ids) == 0 && filter == nil {
		return nil, nil
	}
	list = []dbmodels.ApplicantWithJob{}
	tx := i.db.
		Select("applicants.*, jt.name as job_title_name").
		Model(dbmodels.Applicant{}).
		Joins("left join vacancies as v on vacancy_id = v.id").
		Joins("left join job_titles as jt on v.job_title_id = jt.id").
		Joins("left join selection_stages as st on selection_stage_id = st.id").
		Where("applicants.space_id = ?", spaceID)
	if len(ids) > 0 {
		tx = tx.Where("applicants.id in (?)", ids)
	} else {
		i.addApplicantFilter(tx, *filter)
		i.addSort(tx, filter.Sort)
	}
	err = tx.Preload(clause.Associations).Find(&list).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) ListOfApplicantSource(spaceID string, filter applicantapimodels.ApplicantFilter) ([]dbmodels.ApplicantSource, error) {
	list := []dbmodels.ApplicantSource{}
	tx := i.db.
		Select("count(*) as total, applicants.source, (negotiation_id is not null and negotiation_id <> '') as is_negotiation").
		Model(dbmodels.Applicant{}).
		Joins("left join vacancies as v on vacancy_id = v.id").
		Joins("left join job_titles as jt on v.job_title_id = jt.id").
		Joins("left join selection_stages as st on selection_stage_id = st.id").
		Where("applicants.space_id = ?", spaceID)

	i.addApplicantFilter(tx, filter)
	tx.Group("applicants.source")
	tx.Group("is_negotiation")

	err := tx.Find(&list).Error

	if err != nil {
		return nil, err
	}
	return list, nil
}

func (i impl) ListOfActiveApplicants() ([]dbmodels.Applicant, error) {
	list := []dbmodels.Applicant{}
	tx := i.db.
		Model(dbmodels.Applicant{}).
		Where("applicants.status in (?)", []models.ApplicantStatus{models.ApplicantStatusNegotiation, models.ApplicantStatusInProcess}).
		Where("applicants.negotiation_status <> ?", models.NegotiationStatusRejected).
		Where("applicants.source in (?)", []models.ApplicantSource{models.ApplicantSourceAvito, models.ApplicantSourceHh})
	err := tx.Preload(clause.Associations).Find(&list).Error

	if err != nil {
		return nil, err
	}
	return list, nil
}

func (i impl) ListOfActivefNegotiation(withHrSurvy bool) ([]dbmodels.Applicant, error) {
	list := []dbmodels.Applicant{}
	tx := i.db.
		Model(dbmodels.Applicant{}).
		Where("applicants.negotiation_id is not null").
		Where("applicants.status = ?", models.ApplicantStatusNegotiation)
	if withHrSurvy {
		tx.Joins("left join vacancies as v on vacancy_id = v.id").
			Joins("join hr_surveys as hr on v.id = hr.vacancy_id").
			Where("hr.is_filled_out = true")
	}
	err := tx.Preload(clause.Associations).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (i impl) ListForSurveySend() ([]dbmodels.Applicant, error) {
	list := []dbmodels.Applicant{}
	tx := i.db.
		Model(dbmodels.Applicant{}).
		Where("applicants.negotiation_id is not null").
		Where("applicants.status = ?", models.ApplicantStatusNegotiation)
	tx.Joins("join applicant_surveys as su on su.applicant_id = applicants.id").
		Where("su.is_filled_out = false").
		Where("su.is_scored = false").
		Where("su.is_sent is null")
	err := tx.Preload(clause.Associations).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (i impl) addApplicantFilter(tx *gorm.DB, filter applicantapimodels.ApplicantFilter) {
	if filter.VacancyID != "" {
		tx.Where("applicants.vacancy_id = ?", filter.VacancyID)
	}
	if filter.Status != nil {
		tx.Where("applicants.status = ?", *filter.Status)
	}
	if filter.VacancyName != "" {
		searchValue := "%" + strings.ToLower(filter.VacancyName) + "%"
		tx.Where("LOWER(v.vacancy_name) like ?", searchValue)
	}
	if filter.Search != "" {
		searchValue := "%" + strings.ToLower(filter.Search) + "%"
		sql := "LOWER(CONCAT(last_name,' ', first_name, ' ' , middle_name)) like ?" +
			" or phone like ? or email like ?" +
			" or LOWER(array_to_string(tags,',', '*')) like ?" +
			" or lower(params::TEXT) like ?" +
			" or lower(comment) like ?" +
			" or lower(citizenship) like ?" +
			" or lower(address) like ?" +
			" or lower(source) like ?"
		tx.Where(sql, searchValue, searchValue, searchValue, searchValue, searchValue,
			searchValue, searchValue, searchValue, searchValue)
	}
	if filter.Relocation != nil {
		tx.Where("relocation = ?", *filter.Relocation)
	}
	if filter.AgeFrom > 0 {
		date := time.Now().AddDate(-filter.AgeFrom, 0, 0)
		tx.Where("birth_date <= ?", date)
	}
	if filter.AgeTo > 0 {
		date := time.Now().AddDate(-(filter.AgeTo + 1), 0, 1) // +1 день, чтобы получить включительно
		tx.Where("birth_date >= ?", date)
	}
	if filter.TotalExperienceFrom > 0 {
		tx.Where("total_experience >= ?", filter.TotalExperienceFrom)
	}
	if filter.TotalExperienceTo > 0 {
		tx.Where("total_experience <= ?", filter.TotalExperienceTo)
	}
	if filter.City != "" {
		searchValue := "%" + strings.ToLower(filter.City) + "%"
		tx.Where("LOWER(address) like ?", searchValue)
	}
	if filter.StageName != "" {
		tx.Where("st.name = ?", filter.StageName)
	}
	if filter.Source != nil {
		tx.Where("source = ?", *filter.Source)
	}
	if filter.Tag != "" {
		tx.Where("? = ANY (tags)", filter.Tag)
	}
	if filter.AddedDay != "" {
		date, _ := filter.GetAddedDay()
		fromDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		toDate := fromDate.AddDate(0, 0, 1)
		tx.Where("negotiation_accept_date between ? and ?", fromDate, toDate)
	}
	if filter.AddedPeriod != nil {
		period := *filter.AddedPeriod
		switch period {
		case models.ApAddedPeriodTypeTDay:
			now := time.Now()
			fromDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			toDate := fromDate.AddDate(0, 0, 1)
			tx.Where("negotiation_accept_date between ? and ?", fromDate, toDate)
		case models.ApAddedPeriodTypeYDay:
			now := time.Now()
			fromDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			fromDate = fromDate.AddDate(0, 0, -1)
			toDate := fromDate.AddDate(0, 0, 1)
			tx.Where("negotiation_accept_date between ? and ?", fromDate, toDate)
		case models.ApAddedPeriodType7days:
			tx.Where("((CURRENT_DATE + INTERVAL '1 day')::date - negotiation_accept_date::date) <= 7")
		case models.ApAddedPeriodTypeMonth:
			tx.Where("((CURRENT_DATE + INTERVAL '1 day')::date - negotiation_accept_date::date) <= 30")
		case models.ApAddedPeriodTypeYear:
			tx.Where("((CURRENT_DATE + INTERVAL '1 day')::date - negotiation_accept_date::date) < 365")
		}
	}
	if filter.AddedType != nil {
		exp := *filter.AddedType
		switch exp {
		case models.AddedTypeAdded:
			tx.Where("(negotiation_id is null or negotiation_id = '')")
		case models.AddedTypeNegotiation:
			tx.Where("(negotiation_id is not null and negotiation_id <> '')")
		}
	}
	if filter.Schedule != "" {
		jWhere := fmt.Sprintf("(params->'schedules')::jsonb ? '%v'", filter.Schedule)
		tx.Where(jWhere)
	}
	if filter.Gender != "" {
		tx.Where("gender = ?", filter.Gender)
	}
	if filter.Language != "" {
		jWhere := fmt.Sprintf("(params->'languages')::jsonb @> '[{\"name\":\"%v\"}]'", filter.Language)
		tx.Where(jWhere)
	}
}

func (i impl) addNegotiationFilter(tx *gorm.DB, filter dbmodels.NegotiationFilter) {
	if filter.Search != "" {
		searchValue := "%" + strings.ToLower(filter.Search) + "%"
		tx.Where("LOWER(CONCAT(last_name,' ', first_name, ' ' , middle_name)) like ? or phone like ? or email like ?", searchValue, searchValue, searchValue)
	}
	if filter.Education != nil {
		jWhere := fmt.Sprintf("params @> '{\"education\":\"%v\"}'", *filter.Education)
		tx.Where(jWhere)
	}
	if filter.Experience != nil {
		exp := *filter.Experience
		switch exp {
		case models.ExperienceTypeNo:
			tx.Where("total_experience is null or total_experience = 0")
		case models.ExperienceTypeBetween1And3:
			tx.Where("total_experience between ? AND ?", 12, 36)
		case models.ExperienceTypeBetween3And6:
			tx.Where("total_experience between ? AND ?", 36, 72)
		case models.ExperienceTypeMoreThan6:
			tx.Where("total_experience >= ?", 72)
		}
	}
	if filter.ResponsePeriod != nil {
		period := *filter.ResponsePeriod
		switch period {
		case models.ResponsePeriodType3days:
			tx.Where("((CURRENT_DATE + INTERVAL '1 day')::date - negotiation_date::date) <= 3")
		case models.ResponsePeriodType7days:
			tx.Where("((CURRENT_DATE + INTERVAL '1 day')::date - negotiation_date::date) <= 7")
		case models.ResponsePeriodType7toMonth:
			tx.Where("((CURRENT_DATE + INTERVAL '1 day')::date - negotiation_date::date) > 7 AND " +
				"((CURRENT_DATE + INTERVAL '1 day')::date - negotiation_date::date) <= 30")
		case models.ResponsePeriodTypeMoreMonth:
			tx.Where("((CURRENT_DATE + INTERVAL '1 day')::date - negotiation_date::date) > 30")
		}
	}
	if filter.City != "" {
		searchValue := "%" + strings.ToLower(filter.City) + "%"
		tx.Where("LOWER(address) like ?", searchValue)
	}
	if filter.Employment != nil {
		jWhere := fmt.Sprintf("(params->'employments')::jsonb ? '%v'", *filter.Employment)
		tx.Where(jWhere)
	}
	if filter.Schedule != nil {
		jWhere := fmt.Sprintf("(params->'schedules')::jsonb ? '%v'", *filter.Schedule)
		tx.Where(jWhere)
	}
	if filter.Language != "" {
		jWhere := fmt.Sprintf("(params->'languages')::jsonb @> '[{\"name\":\"%v\"}]'", filter.Language)
		tx.Where(jWhere)
	}
	if filter.LanguageLevel != nil {
		jWhere := fmt.Sprintf("(params->'languages')::jsonb @> '[{\"language_level\":\"%v\"}]'", *filter.LanguageLevel)
		tx.Where(jWhere)
	}
	if filter.Gender != nil {
		tx.Where("gender = ?", *filter.Gender)
	}
	if filter.TripReadiness != nil {
		jWhere := fmt.Sprintf("params @> '{\"trip_readiness\":\"%v\"}'", *filter.TripReadiness)
		tx.Where(jWhere)
	}
	if filter.Citizenship != "" {
		tx.Where("citizenship = ?", filter.Citizenship)
	}
	if filter.SalaryFrom > 0 {
		tx.Where("salary >= ?", filter.SalaryFrom)
	}
	if filter.SalaryTo > 0 {
		tx.Where("salary <= ?", filter.SalaryTo)
	}
	if filter.SalaryProvided != nil {
		provided := *filter.SalaryProvided
		if provided {
			tx.Where("salary is not null and salary > 0")
		} else {
			tx.Where("salary is null")
		}
	}
	if filter.Source != nil {
		tx.Where("source = ?", filter.Source)
	}
	if len(filter.DriverLicence) != 0 {
		var buffer bytes.Buffer
		licenceLen := len(filter.DriverLicence)
		for k, licence := range filter.DriverLicence {
			if licenceLen > 1 && k != 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString(fmt.Sprintf("'%v'", licence))
		}
		jWhere := fmt.Sprintf("(params->>'driver_license_types')::jsonb ?&array[%v]", buffer.String())
		tx.Where(jWhere)
	}
	if filter.JobSearchStatuses != nil {
		jWhere := fmt.Sprintf("params @> '{\"search_status\":\"%v\"}'", *filter.JobSearchStatuses)
		tx.Where(jWhere)
	}
	if filter.SearchLabel != nil {
		switch *filter.SearchLabel {
		case models.SearchLabelPhoto:
			tx.Where("photo_uri <> ''")
		case models.SearchLabelSalary:
			tx.Where("salary is not null or salary > 0")
		case models.SearchLabelAge:
			tx.Where("birth_date >'1900-01-01'")
		case models.SearchLabelGender:
			tx.Where("gender <> ''")
		}
	}
	if filter.AdvancedTraining != nil {
		if *filter.AdvancedTraining {
			jWhere := fmt.Sprintf("params @> '{\"have_additional_education\":true}'")
			tx.Where(jWhere)
		} else {
			jWhere := fmt.Sprintf("params @> '{\"have_additional_education\":false}'")
			tx.Where(jWhere)
		}
	}
}

func (i impl) setPage(tx *gorm.DB, page, limit int) {
	offset := (page - 1) * limit
	tx.Limit(limit).Offset(offset)
}

func (i impl) addSort(tx *gorm.DB, sort applicantapimodels.ApplicantSort) {
	if sort.FioDesc != nil {
		specifyOrder(tx, "fio", *sort.FioDesc)
	}

	if sort.SalaryDesc != nil {
		specifyOrder(tx, "applicants.salary", *sort.SalaryDesc)
	}

	if sort.AcceptDateDesc != nil {
		specifyOrder(tx, "applicants.negotiation_accept_date", *sort.AcceptDateDesc)
	}

	if sort.NegotiationDateDesc != nil {
		specifyOrder(tx, "applicants.negotiation_date", *sort.NegotiationDateDesc)
	}

	if sort.StartDateDesc != nil {
		specifyOrder(tx, "applicants.start_date", *sort.StartDateDesc)
	}

	if sort.VacancyNameDesc != nil {
		specifyOrder(tx, "v.vacancy_name", *sort.VacancyNameDesc)
	}

	if sort.ResumeTitleDesc != nil {
		specifyOrder(tx, "applicants.resume_title", *sort.ResumeTitleDesc)
	}

	if sort.SourceDesc != nil {
		specifyOrder(tx, "applicants.source", *sort.SourceDesc)
	}

	if sort.StatusDesc != nil {
		specifyOrder(tx, "applicants.status", *sort.StatusDesc)
	}
}

func specifyOrder(tx *gorm.DB, fieldName string, isDesc bool) {
	if isDesc {
		tx = tx.Order(fmt.Sprintf("%v desc", fieldName))
		return
	}
	tx = tx.Order(fmt.Sprintf("%v  asc", fieldName))
}
