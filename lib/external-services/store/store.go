package extservicestore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Set(spaceID, code string, value []byte) error
	Get(spaceID, code string) (value []byte, ok bool, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Set(spaceID, code string, value []byte) error {
	rec, err := i.get(spaceID, code)
	if err != nil {
		return err
	}
	if rec == nil {
		rec = &dbmodels.ExtData{
			SpaceID: spaceID,
			Code:    code,
			Value:   value,
		}
		return i.db.
			Save(rec).
			Error
	}
	updMap := map[string]interface{}{
		"Value": value,
	}
	tx := i.db.
		Model(&dbmodels.ExtData{}).
		Where("id = ?", rec.ID).
		Updates(updMap)
	err = tx.Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) Get(spaceID, code string) (value []byte, ok bool, err error) {
	rec, err := i.get(spaceID, code)
	if err != nil || rec == nil {
		return nil, false, err
	}
	return rec.Value, true, nil
}

func (i impl) get(spaceID, code string) (rec *dbmodels.ExtData, err error) {
	err = i.db.Model(dbmodels.ExtData{}).
		Where("space_id = ? AND code = ?", spaceID, code).
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
