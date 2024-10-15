package dictapimodels

import (
	"github.com/pkg/errors"
	dbmodels "hr-tools-backend/models/db"
)

type JobTitleData struct {
	Name         string `json:"name"`
	DepartmentID string `json:"department_id"`
}

type JobTitleView struct {
	JobTitleData
	ID string `json:"id"`
}

func (j JobTitleData) Validate() error {
	if j.DepartmentID == "" {
		return errors.New("отсутсвует ссылка на подразделение")
	}
	if j.Name == "" {
		return errors.New("не указано название штатной должности")
	}
	return nil
}

func JobTitleConvert(rec dbmodels.JobTitle) JobTitleView {
	return JobTitleView{
		JobTitleData: JobTitleData{
			Name:         rec.Name,
			DepartmentID: rec.DepartmentID,
		},
		ID: rec.ID,
	}
}
