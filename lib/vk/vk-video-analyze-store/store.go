package vkvideoanalyzestore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Save(rec dbmodels.ApplicantVkVideoSurvey) (id string, err error)
	GetByStepQuestion(applicantVkStepID, questionID string) (*dbmodels.ApplicantVkVideoSurvey, error)
	GetByApplicantVkStep(applicantVkStepID string) ([]dbmodels.ApplicantVkVideoSurvey, error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Save(rec dbmodels.ApplicantVkVideoSurvey) (id string, err error) {
	existedRec, err := i.GetByStepQuestion(rec.ApplicantVkStepID, rec.QuestionID)
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



func (i impl) GetByStepQuestion(applicantVkStepID, questionID string) (*dbmodels.ApplicantVkVideoSurvey, error) {
	rec := dbmodels.ApplicantVkVideoSurvey{}
	err := i.db.
		Where("applicant_vk_step_id = ?", applicantVkStepID).
		Where("question_id = ?", questionID).
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

func (i impl) GetByApplicantVkStep(applicantVkStepID string) ([]dbmodels.ApplicantVkVideoSurvey, error) {
	list := []dbmodels.ApplicantVkVideoSurvey{}
	tx := i.db.
		Model(dbmodels.ApplicantVkVideoSurvey{}).
		Where("applicant_vk_step_id = ?", applicantVkStepID)
	err := tx.Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}