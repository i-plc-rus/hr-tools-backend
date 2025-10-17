package filesdbstorage

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	GetByID(id string) (rec *dbmodels.FileStorage, err error)
	SaveFile(rec dbmodels.FileStorage) (id string, err error)
	DeleteFile(id, spaceID string) (ok bool, err error)
	GetFileIDByType(applicantID string, fileType dbmodels.FileType) (rec *dbmodels.FileStorage, err error)
	GetFileListByType(applicantID string, fileType dbmodels.FileType) (list []dbmodels.FileStorage, err error)
}

func NewInstance(db *gorm.DB) Provider {
	return &impl{db: db}
}

type impl struct {
	db *gorm.DB
}

func (i impl) GetByID(id string) (rec *dbmodels.FileStorage, err error) {
	rec = new(dbmodels.FileStorage)
	err = i.db.
		Model(&dbmodels.FileStorage{}).
		Where("id = ?", id).
		First(&rec).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rec, nil
}

func (i impl) GetFileListByType(applicantID string, fileType dbmodels.FileType) (list []dbmodels.FileStorage, err error) {
	err = i.db.
		Model(&dbmodels.FileStorage{}).
		Where("applicant_id = ? AND type = ?", applicantID, fileType).
		Find(&list).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return list, nil
}

func (i impl) GetFileIDByType(applicantID string, fileType dbmodels.FileType) (rec *dbmodels.FileStorage, err error) {
	rec = new(dbmodels.FileStorage)
	err = i.db.
		Model(&dbmodels.FileStorage{}).
		Where("applicant_id = ? AND type = ?", applicantID, fileType).
		First(&rec).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rec, nil
}

func (i impl) SaveFile(rec dbmodels.FileStorage) (id string, err error) {
	err = i.db.Save(&rec).Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) DeleteFile(id, spaceID string) (ok bool, err error) {
	rec := dbmodels.FileStorage{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			BaseModel: dbmodels.BaseModel{ID: id},
			SpaceID:   spaceID,
		},
	}
	tx := i.db.Delete(&rec)

	err = tx.Error
	ok = tx.RowsAffected > 0

	if err != nil {
		return ok, err
	}
	return ok, nil
}
