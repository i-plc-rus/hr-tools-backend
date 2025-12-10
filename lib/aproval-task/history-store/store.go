package approvaltaskhistorystore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(rec dbmodels.ApprovalHistory) (id string, err error)
	Delete(id string) error
	DeleteByVacancyRequest(spaceID, requestID string) error
	List(spaceID, requestID string) (list []dbmodels.ApprovalHistory, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.ApprovalHistory) (id string, err error) {
	err = i.db.
		Omit("AssigneeUser").
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) Delete(id string) error {
	rec := dbmodels.ApprovalHistory{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			BaseModel: dbmodels.BaseModel{ID: id},
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
	rec := dbmodels.ApprovalHistory{}
	err := i.db.Model(&dbmodels.ApprovalHistory{}).
		Where("space_id = ?", spaceID).
		Where("request_id = ?", requestID).
		Delete(&rec).Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) List(spaceID, requestID string) (list []dbmodels.ApprovalHistory, err error) {
	list = []dbmodels.ApprovalHistory{}
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
