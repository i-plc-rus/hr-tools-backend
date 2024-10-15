package approvalstagestore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(rec dbmodels.ApprovalStage) (id string, err error)
	GetByID(spaceID, id string) (rec *dbmodels.ApprovalStage, err error)
	Update(spaceID, id string, updMap map[string]interface{}) error
	Delete(spaceID, id string) error
	DeleteByVacancyRequest(spaceID, vacancyRequestID string) error
	List(spaceID, vacancyRequestID string) (list []dbmodels.ApprovalStage, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.ApprovalStage) (id string, err error) {
	err = i.db.
		Omit("VacancyRequest").
		Omit("SpaceUser").
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) GetByID(spaceID, id string) (*dbmodels.ApprovalStage, error) {
	rec := dbmodels.ApprovalStage{}
	err := i.db.
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
		Preload("SpaceUser").
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
	err := i.db.
		Model(&dbmodels.ApprovalStage{}).
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
		Updates(updMap).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) Delete(spaceID, id string) error {
	rec := dbmodels.ApprovalStage{
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

func (i impl) DeleteByVacancyRequest(spaceID, vacancyRequestID string) error {
	rec := dbmodels.ApprovalStage{}
	err := i.db.Model(&dbmodels.ApprovalStage{}).
		Where("space_id = ?", spaceID).
		Where("vacancy_request_id = ?", vacancyRequestID).
		Delete(&rec).Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) List(spaceID, vacancyRequestID string) (list []dbmodels.ApprovalStage, err error) {
	list = []dbmodels.ApprovalStage{}
	tx := i.db.
		Where("space_id = ?", spaceID).
		Where("vacancy_request_id = ?", vacancyRequestID).
		Preload("SpaceUser")
	err = tx.Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}
