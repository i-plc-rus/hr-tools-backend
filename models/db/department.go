package dbmodels

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Department struct {
	BaseSpaceModel
	CompanyStructID string `gorm:"type:varchar(36);index:idx_company_struct"`
	ParentID        string `gorm:"type:varchar(36);index:idx_company_struct"`
	Name            string `gorm:"type:varchar(255)"`
	BusinessAreaID  int
}

func (d *Department) AfterDelete(tx *gorm.DB) (err error) {
	if d.ID == "" {
		return nil
	}
	tx.Clauses(clause.Returning{}).Where("department_id = ?", d.ID).Delete(&JobTitle{})
	tx.Clauses(clause.Returning{}).Where("parent_id = ?", d.ID).Delete(&Department{})
	return
}

func (d *Department) Validate() error {
	if err := d.BaseSpaceModel.Validate(); err != nil {
		return err
	}
	if d.CompanyStructID == "" {
		return errors.New("отсутсвует ссылка на структуру компании")
	}
	if d.Name == "" {
		return errors.New("не указано название подразделения")
	}
	return nil
}
