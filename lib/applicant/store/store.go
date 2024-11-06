package applicantstore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(data dbmodels.Applicant) (id string, err error)
	GetByID(spaceID, id string) (rec *dbmodels.Applicant, err error)
	IsExistNegotiationID(spaceID, negotiationID string, source models.ApplicantSource) (found bool, err error)
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

func (i impl) GetByID(spaceID, id string) (*dbmodels.Applicant, error) {
	rec := dbmodels.Applicant{}
	err := i.db.
		Model(&dbmodels.Applicant{}).
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
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
