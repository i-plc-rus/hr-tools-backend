package applicantstore

import (
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

}
