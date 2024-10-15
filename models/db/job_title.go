package dbmodels

import "github.com/pkg/errors"

type JobTitle struct {
	BaseSpaceModel
	DepartmentID string `gorm:"type:varchar(36);index"`
	Name         string `gorm:"type:varchar(255)"`
}

func (j JobTitle) Validate() error {
	if err := j.BaseSpaceModel.Validate(); err != nil {
		return err
	}
	if j.DepartmentID == "" {
		return errors.New("отсутсвует ссылка на подразделение")
	}
	if j.Name == "" {
		return errors.New("не указано название штатной должности")
	}
	return nil
}
