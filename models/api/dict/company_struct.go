package dictapimodels

import (
	"github.com/pkg/errors"
	dbmodels "hr-tools-backend/models/db"
)

type CompanyStructData struct {
	Name string `json:"name"`
}

type CompanyStructView struct {
	CompanyStructData
	ID string `json:"id"`
}

func (c *CompanyStructData) Validate() error {
	if c.Name == "" {
		return errors.New("не указано название компании")
	}
	return nil
}

func CompanyStructConvert(rec dbmodels.CompanyStruct) CompanyStructView {
	return CompanyStructView{
		CompanyStructData: CompanyStructData{
			Name: rec.Name,
		},
		ID: rec.ID,
	}
}
