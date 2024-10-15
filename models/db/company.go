package dbmodels

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Company struct {
	BaseSpaceModel
	Name string `gorm:"type:varchar(255)"`
}

func (c *Company) AfterDelete(tx *gorm.DB) (err error) {
	if c.ID == "" {
		return nil
	}
	tx.Clauses(clause.Returning{}).Where("company_id = ?", c.ID).Delete(&Department{})
	return
}

func (c *Company) Validate() error {
	if err := c.BaseSpaceModel.Validate(); err != nil {
		return err
	}
	if c.Name == "" {
		return errors.New("не указано название компании")
	}
	return nil
}
