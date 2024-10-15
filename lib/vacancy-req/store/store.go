package vacancyreqstore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(rec dbmodels.VacancyRequest) (id string, err error)
	GetByID(spaceID, id string) (rec *dbmodels.VacancyRequest, err error)
	Update(spaceID, id string, updMap map[string]interface{}) error
	Delete(spaceID, id string) error
	List(spaceID string) (list []dbmodels.VacancyRequest, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.VacancyRequest) (id string, err error) {
	err = i.db.Omit(clause.Associations).
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) GetByID(spaceID, id string) (*dbmodels.VacancyRequest, error) {
	rec := dbmodels.VacancyRequest{}
	err := i.db.
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
		Preload(clause.Associations).
		Preload("ApprovalStages.SpaceUser").
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

func (i impl) Update(spaceID, id string, updMap map[string]interface{}) error {
	if len(updMap) == 0 {
		return nil
	}
	tx := i.db.
		Model(&dbmodels.VacancyRequest{}).
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
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

func (i impl) Delete(spaceID, id string) error {
	rec := dbmodels.VacancyRequest{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			BaseModel: dbmodels.BaseModel{ID: id},
			SpaceID:   spaceID,
		},
	}
	err := i.db.
		Delete(&rec).
		Error

	if err != nil {
		return err
	}
	return nil
}

func (i impl) List(spaceID string) (list []dbmodels.VacancyRequest, err error) {
	list = []dbmodels.VacancyRequest{}
	tx := i.db.Where("space_id = ?", spaceID).
		Preload(clause.Associations).
		Preload("ApprovalStages.SpaceUser")
	err = tx.Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}
