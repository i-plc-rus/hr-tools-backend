package vacancysurvaystore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Save(rec dbmodels.HRSurvey) (id string, err error)
	GetByVacancyID(spaceID, vacancyID string) (rec *dbmodels.HRSurvey, err error)
	Update(spaceID, id string, updMap map[string]interface{}) error
	Delete(spaceID, id string) error
	DeleteByVacancyID(spaceID, vacancyID string) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Save(rec dbmodels.HRSurvey) (id string, err error) {
	existedRec, err := i.GetByVacancyID(rec.SpaceID, rec.VacancyID)
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

func (i impl) GetByVacancyID(spaceID, vacancyID string) (*dbmodels.HRSurvey, error) {
	rec := dbmodels.HRSurvey{}
	err := i.db.
		Where("space_id = ?", spaceID).
		Where("vacancy_id = ?", vacancyID).
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
		Model(&dbmodels.HRSurvey{}).
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
	rec := dbmodels.HRSurvey{
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

func (i impl) DeleteByVacancyID(spaceID, vacancyID string) error {
	rec := dbmodels.HRSurvey{}
	err := i.db.Model(&dbmodels.HRSurvey{}).
		Where("space_id = ?", spaceID).
		Where("vacancy_id = ?", vacancyID).
		Delete(&rec).Error
	if err != nil {
		return err
	}
	return nil
}
