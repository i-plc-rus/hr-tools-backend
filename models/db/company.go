package dbmodels

import (
	"github.com/pkg/errors"
)

type Company struct {
	BaseSpaceModel
	Name string `gorm:"type:varchar(255)"`
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
