package dbmodels

import "github.com/pkg/errors"

type CompanyStruct struct {
	BaseSpaceModel
	Name string `gorm:"type:varchar(255)"`
}

func (c *CompanyStruct) Validate() error {
	if err := c.BaseSpaceModel.Validate(); err != nil {
		return err
	}
	if c.Name == "" {
		return errors.New("не указано название структуры")
	}
	return nil
}
