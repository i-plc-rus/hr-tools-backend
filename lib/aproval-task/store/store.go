package approvaltaskstore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(rec dbmodels.ApprovalTask) (id string, err error)
	GetByID(spaceID, id string) (rec *dbmodels.ApprovalTask, err error)
	Update(spaceID, id string, updMap map[string]interface{}) error
	Delete(spaceID, id string) error
	DeleteByVacancyRequest(spaceID, requestID string) error
	List(spaceID, requestID string) (list []dbmodels.ApprovalTask, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.ApprovalTask) (id string, err error) {
	err = i.db.
		Omit("AssigneeUser").
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) GetByID(spaceID, id string) (*dbmodels.ApprovalTask, error) {
	rec := dbmodels.ApprovalTask{}
	err := i.db.
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
		Preload("AssigneeUser").
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
		Model(&dbmodels.ApprovalTask{}).
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
	rec := dbmodels.ApprovalTask{
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

func (i impl) DeleteByVacancyRequest(spaceID, requestID string) error {
	rec := dbmodels.ApprovalTask{}
	err := i.db.Model(&dbmodels.ApprovalTask{}).
		Where("space_id = ?", spaceID).
		Where("request_id = ?", requestID).
		Delete(&rec).Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) List(spaceID, requestID string) (list []dbmodels.ApprovalTask, err error) {
	list = []dbmodels.ApprovalTask{}
	tx := i.db.
		Where("space_id = ?", spaceID).
		Where("request_id = ?", requestID).
		Order("created_at ASC").
		Preload("AssigneeUser")
	err = tx.Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}
