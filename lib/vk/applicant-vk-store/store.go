package applicantvkstore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Save(rec dbmodels.ApplicantVkStep) (id string, err error)
	GetByID(id string) (rec *dbmodels.ApplicantVkStep, err error)
	GetByApplicantID(spaceID, applicantID string) (*dbmodels.ApplicantVkStep, error)
	Delete(spaceID, id string) error
	DeleteByApplicantID(spaceID, applicantID string) error
	GetByStatus(status dbmodels.StepStatus) ([]dbmodels.ApplicantVkStep, error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Save(rec dbmodels.ApplicantVkStep) (id string, err error) {
	existedRec, err := i.GetByApplicantID(rec.SpaceID, rec.ApplicantID)
	if err != nil {
		return "", err
	}
	if existedRec != nil {
		rec.ID = existedRec.ID
	}
	err = i.db.
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) GetByID(id string) (*dbmodels.ApplicantVkStep, error) {
	rec := dbmodels.ApplicantVkStep{}
	err := i.db.
		Where("id = ?", id).
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

func (i impl) GetByApplicantID(spaceID, applicantID string) (*dbmodels.ApplicantVkStep, error) {
	rec := dbmodels.ApplicantVkStep{}
	err := i.db.
		Where("space_id = ?", spaceID).
		Where("applicant_id = ?", applicantID).
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

func (i impl) Delete(spaceID, id string) error {
	rec := dbmodels.ApplicantVkStep{
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

func (i impl) DeleteByApplicantID(spaceID, applicantID string) error {
	rec := dbmodels.ApplicantVkStep{}
	err := i.db.Model(&dbmodels.ApplicantVkStep{}).
		Where("space_id = ?", spaceID).
		Where("applicant_id = ?", applicantID).
		Delete(&rec).Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) GetByStatus(status dbmodels.StepStatus) ([]dbmodels.ApplicantVkStep, error) {
	list := []dbmodels.ApplicantVkStep{}
	tx := i.db.
		Model(dbmodels.ApplicantVkStep{}).
		Where("status = ?", status)
	err := tx.Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}