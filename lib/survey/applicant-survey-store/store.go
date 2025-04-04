package applicantsurveystore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Save(rec dbmodels.ApplicantSurvey) (id string, err error)
	GetByID(id string) (rec *dbmodels.ApplicantSurvey, err error)
	GetSurveyForScore() (list []dbmodels.ApplicantSurvey, err error)
	GetByApplicantID(spaceID, applicantID string) (rec *dbmodels.ApplicantSurvey, err error)
	Delete(spaceID, id string) error
	DeleteByApplicantID(spaceID, applicantID string) error
	SetIsSend(id string, isSend bool) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Save(rec dbmodels.ApplicantSurvey) (id string, err error) {
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

func (i impl) GetByID(id string) (*dbmodels.ApplicantSurvey, error) {
	rec := dbmodels.ApplicantSurvey{}
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

func (i impl) GetSurveyForScore() (list []dbmodels.ApplicantSurvey, err error) {
	list = []dbmodels.ApplicantSurvey{}
	tx := i.db.
		Select("applicant_surveys.*").
		Model(dbmodels.ApplicantSurvey{}).
		Joins("join applicants as a on applicant_id = a.id").
		Where("a.status = 'Отклик'").
		Where("is_filled_out = true").
		Where("is_scored = false")
	err = tx.Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (i impl) GetByApplicantID(spaceID, applicantID string) (*dbmodels.ApplicantSurvey, error) {
	rec := dbmodels.ApplicantSurvey{}
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
	rec := dbmodels.ApplicantSurvey{
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
	rec := dbmodels.ApplicantSurvey{}
	err := i.db.Model(&dbmodels.ApplicantSurvey{}).
		Where("space_id = ?", spaceID).
		Where("applicant_id = ?", applicantID).
		Delete(&rec).Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) SetIsSend(id string, isSend bool) error {
	updMap := map[string]interface{}{
		"is_sent": isSend,
	}
	err := i.db.
		Model(&dbmodels.ApplicantSurvey{}).
		Where("id = ?", id).
		Updates(updMap).
		Error
	if err != nil {
		return err
	}
	return nil
}
