package rejectreasondictstore

import (
	"hr-tools-backend/models"
	dictapimodels "hr-tools-backend/models/api/dict"
	dbmodels "hr-tools-backend/models/db"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Provider interface {
	IsUnique(spaceID string, selfID, name string, initiator models.RejectInitiator) (bool, error)
	Create(rec dbmodels.RejectReason) (id string, err error)
	GetByID(spaceID, id string) (rec *dbmodels.RejectReason, err error)
	List(spaceID string, filter dictapimodels.RejectReasonFind) (list []dbmodels.RejectReason, err error)
	Update(spaceID, id string, updMap map[string]interface{}) error
	Delete(spaceID, id string) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.RejectReason) (id string, err error) {
	err = rec.Validate()
	if err != nil {
		return "", err
	}

	err = i.db.
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) GetByID(spaceID, id string) (*dbmodels.RejectReason, error) {
	rec := dbmodels.RejectReason{}
	err := i.db.
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
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

func (i impl) List(spaceID string, filter dictapimodels.RejectReasonFind) (list []dbmodels.RejectReason, err error) {
	list = []dbmodels.RejectReason{}
	tx := i.db.Where("space_id = ?", spaceID)
	if filter.Initiator != "" {
		tx = tx.Where("initiator = ?", filter.Initiator)
	}
	if filter.Search != "" {
		tx.Where("LOWER(name) like ?", "%"+strings.ToLower(filter.Search)+"%")
	}
	err = tx.Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (i impl) Update(spaceID, id string, updMap map[string]interface{}) error {
	if len(updMap) == 0 {
		return nil
	}
	err := i.db.
		Model(&dbmodels.RejectReason{}).
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
	rec := dbmodels.RejectReason{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			BaseModel: dbmodels.BaseModel{
				ID: id,
			},
			SpaceID: spaceID,
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

func (i impl) IsUnique(spaceID string, selfID, name string, initiator models.RejectInitiator) (bool, error) {
	var rowCount int64
	tx := i.db.Model(dbmodels.RejectReason{})
	tx.Where("space_id = ?", spaceID)
	tx.Where("name = ?", name)
	tx.Where("initiator = ?", initiator)
	if selfID != "" {
		tx.Where("id <> ?", selfID)
	}
	err := tx.Count(&rowCount).Error
	if err != nil {
		return false, errors.Wrap(err, "ошибка проверки уникальности причины отказа")
	}

	return rowCount != 0, nil
}
