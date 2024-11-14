package applicantstore

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"strings"
)

type Provider interface {
	Create(data dbmodels.Applicant) (id string, err error)
	Update(id string, updMap map[string]interface{}) error
	GetByID(spaceID, id string) (rec *dbmodels.ApplicantExt, err error)
	IsExistNegotiationID(spaceID, negotiationID string, source models.ApplicantSource) (found bool, err error)
	ListOfNegotiation(spaceID string, filter dbmodels.NegotiationFilter) ([]dbmodels.Applicant, error)
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
		Where("negotiation_id is not null")
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
		if exp == models.ExperienceTypeNo {
			fmt.Println("no")
		}
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
