package promptcheckstore

import (
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Provider interface {
	Save(rec dbmodels.PromptExecution) (string, error)
	Update(id string, updMap map[string]any) error
	GetByID(id string) (*dbmodels.PromptExecution, error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Save(rec dbmodels.PromptExecution) (string, error) {
	err := i.db.
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) Update(id string, updMap map[string]any) error {
	if len(updMap) == 0 {
		return nil
	}
	err := i.db.
		Model(&dbmodels.PromptExecution{}).
		Where("id = ?", id).
		Updates(updMap).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) GetByID(id string) (*dbmodels.PromptExecution, error) {
	rec := dbmodels.PromptExecution{BaseModel: dbmodels.BaseModel{ID: id}}
	err := i.db.First(&rec).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}
