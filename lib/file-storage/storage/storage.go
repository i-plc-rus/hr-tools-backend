package filesdbstorage

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	SaveFile(rec dbmodels.FileStorage) (id string, err error)
	GetFileIDByType(applicantID string, fileType dbmodels.FileType) (id string, err error)
	GetFileListByType(applicantID string, fileType dbmodels.FileType) (list []dbmodels.FileStorage, err error)
}

type impl struct {
	db *gorm.DB
}

func (i impl) GetFileListByType(applicantID string, fileType dbmodels.FileType) (list []dbmodels.FileStorage, err error) {
	err = i.db.
		Model(&dbmodels.FileStorage{}).
		Where("applicant_id = ? AND file_type = ?", applicantID, fileType).
		Find(&list).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return list, nil
}

func (i impl) GetFileIDByType(applicantID string, fileType dbmodels.FileType) (id string, err error) {
	rec := dbmodels.FileStorage{}
	err = i.db.
		Model(&dbmodels.FileStorage{}).
		Where("applicant_id = ? AND file_type = ?", applicantID, fileType).
		First(&rec).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return rec.ID, nil
}

func (i impl) SaveFile(rec dbmodels.FileStorage) (id string, err error) {
	err = i.db.Save(&rec).Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func NewInstance(db *gorm.DB) Provider {
	return &impl{db: db}
}
