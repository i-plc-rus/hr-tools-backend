package dictapimodels

import (
	"github.com/pkg/errors"
	dbmodels "hr-tools-backend/models/db"
)

type DepartmentData struct {
	Name      string `json:"name"`
	CompanyID string `json:"company_id"`
	ParentID  string `json:"parent_id"`
}

type DepartmentView struct {
	DepartmentData
	ID string `json:"id"`
}

type DepartmentFind struct {
	Name      string `json:"name"`
	CompanyID string `json:"company_id"`
}

func (c DepartmentData) Validate() error {
	if c.CompanyID == "" {
		return errors.New("отсутсвует ссылка на компанию")
	}
	if c.Name == "" {
		return errors.New("не указано название подразделения")
	}
	return nil
}

func DepartmentConvert(rec dbmodels.Department) DepartmentView {
	return DepartmentView{
		DepartmentData: DepartmentData{
			Name:      rec.Name,
			CompanyID: rec.CompanyID,
			ParentID:  rec.ParentID,
		},
		ID: rec.ID,
	}
}

type DepartmentTreeItem struct {
	DepartmentView
	SubUnits []DepartmentTreeItem `json:"sub_units"`
}
