package dictapimodels

import (
	"github.com/pkg/errors"
	dbmodels "hr-tools-backend/models/db"
)

type CompanyData struct {
	Name string `json:"name"`
}

type CompanyView struct {
	CompanyData
	ID string `json:"id"`
}

func (c *CompanyData) Validate() error {
	if c.Name == "" {
		return errors.New("не указано название компании")
	}
	return nil
}

func CompanyConvert(rec dbmodels.Company) CompanyView {
	return CompanyView{
		CompanyData: CompanyData{
			Name: rec.Name,
		},
		ID: rec.ID,
	}
}
