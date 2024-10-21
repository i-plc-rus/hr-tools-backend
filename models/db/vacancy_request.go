package dbmodels

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"hr-tools-backend/models"
	"sort"
)

type VacancyRequest struct {
	BaseSpaceModel
	AuthorID        string
	Author          *SpaceUser `gorm:"foreignKey:AuthorID"`
	Space           *Space
	CompanyID       *string `gorm:"type:varchar(36);index:idx_company"`
	Company         *Company
	DepartmentID    *string `gorm:"type:varchar(36)"`
	Department      *Department
	JobTitleID      *string `gorm:"type:varchar(36)"`
	JobTitle        *JobTitle
	CityID          *string `gorm:"type:varchar(36)"`
	City            *City
	CompanyStructID *string `gorm:"type:varchar(36)"`
	CompanyStruct   *CompanyStruct
	VacancyName     string `gorm:"type:varchar(255)"`
	Confidential    bool
	OpenedPositions int
	Urgency         models.VRUrgency       `gorm:"type:varchar(100)"`
	RequestType     models.VRType          `gorm:"type:varchar(100)"`
	SelectionType   models.VRSelectionType `gorm:"type:varchar(100)"`
	PlaceOfWork     string                 `gorm:"type:varchar(255)"`
	ChiefFio        string                 `gorm:"type:varchar(255)"`
	Interviewer     string                 `gorm:"type:varchar(255)"`
	ShortInfo       string
	Requirements    string
	Description     string
	OutInteraction  string
	InInteraction   string
	Status          models.VRStatus
	ApprovalStages  []*ApprovalStage
}

func (v *VacancyRequest) AfterDelete(tx *gorm.DB) (err error) {
	if v.ID == "" {
		return nil
	}
	tx.Clauses(clause.Returning{}).Where("vacancy_request_id = ?", v.ID).Delete(&ApprovalStage{})
	return
}

func (v *VacancyRequest) GetCurrentApprovalStage() (isLast bool, stage *ApprovalStage) {
	sort.Slice(v.ApprovalStages, func(i, j int) bool {
		return v.ApprovalStages[i].Stage < v.ApprovalStages[j].Stage
	})
	for k, item := range v.ApprovalStages {
		if item.ApprovalStatus == models.AStatusApproved {
			continue
		}
		return k == len(v.ApprovalStages)-1, item
	}
	return true, nil
}
