package teamstore

import (
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Provider interface {
	Create(rec dbmodels.VacancyTeam) (id string, err error)
	Update(spaceID, vacancyID, userID string, updMap map[string]interface{}) error
	GetByID(spaceID, vacancyID, userID string) (*dbmodels.VacancyTeam, error)
	List(spaceID, vacancyID string) (list []dbmodels.VacancyTeam, err error)
	Delete(spaceID, vacancyID, userID string) (err error)
	SetAsResponsible(spaceID, vacancyID, userID string) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.VacancyTeam) (id string, err error) {
	err = i.db.
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) GetByID(spaceID, vacancyID, userID string) (*dbmodels.VacancyTeam, error) {
	rec := dbmodels.VacancyTeam{}
	err := i.db.
		Model(&dbmodels.VacancyTeam{}).
		Where("user_id = ?", userID).
		Where("space_id = ?", spaceID).
		Where("vacancy_id = ?", vacancyID).
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

func (i impl) Update(spaceID, vacancyID, userID string, updMap map[string]interface{}) error {
	if len(updMap) == 0 {
		return nil
	}
	err := i.db.
		Model(&dbmodels.VacancyTeam{}).
		Where("user_id = ?", userID).
		Where("space_id = ?", spaceID).
		Where("vacancy_id = ?", vacancyID).
		Updates(updMap).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) List(spaceID, vacancyID string) (list []dbmodels.VacancyTeam, err error) {
	list = []dbmodels.VacancyTeam{}
	tx := i.db.
		Where("space_id = ?", spaceID).
		Where("vacancy_id = ?", vacancyID).
		Preload(clause.Associations)
	err = tx.Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) Delete(spaceID, vacancyID, userID string) (err error) {
	err = i.db.
		Model(&dbmodels.VacancyTeam{}).
		Where("space_id = ?", spaceID).
		Where("vacancy_id = ?", vacancyID).
		Where("user_id = ?", userID).
		Delete(&dbmodels.VacancyTeam{}).Error

	if err != nil {
		return err
	}
	return nil
}

func (i impl) SetAsResponsible(spaceID, vacancyID, userID string) error {
	tx := i.db.Model(&dbmodels.VacancyTeam{}).
		Where("space_id = ?", spaceID).
		Where("vacancy_id = ?", vacancyID).
		UpdateColumn("responsible", gorm.Expr("user_id = ?", userID))
	err := tx.Error
	if err != nil {
		return err
	}
	if tx.RowsAffected == 0 {
		return errors.New("нет записей для обновления")
	}
	return nil
}
