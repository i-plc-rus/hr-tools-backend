package masaisessionstore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Save(rec dbmodels.MasaiSession) (id string, err error)
	GetAll() ([]dbmodels.MasaiSession, error)
	Delete(id string) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Save(rec dbmodels.MasaiSession) (id string, err error) {
	existedRecs, err := i.GetAll()
	if err != nil {
		return "", err
	}
	if len(existedRecs) > 1 {
		return "", errors.New("найден незавершенный запрос")
	} else if len(existedRecs) == 1 {
		//обновляется существующая сессия
		if existedRecs[0].VkStepID == rec.VkStepID && existedRecs[0].QuestionID == rec.QuestionID {
			rec.ID = existedRecs[0].ID
		} else {
			return "", errors.New("найден незавершенный запрос")
		}
	}
	err = i.db.
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) GetAll() ([]dbmodels.MasaiSession, error) {
	list := []dbmodels.MasaiSession{}
	tx := i.db.
		Model(dbmodels.MasaiSession{})
	err := tx.Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (i impl) Delete(id string) error {
	rec := dbmodels.MasaiSession{
		BaseModel: dbmodels.BaseModel{ID: id},
	}
	err := i.db.
		Delete(&rec).
		Error

	if err != nil {
		return err
	}
	return nil
}
