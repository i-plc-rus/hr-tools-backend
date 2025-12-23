package dbmodels

import (
	"hr-tools-backend/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	Employment      models.Employment `gorm:"type:varchar(255)"` // Занятость
	Experience      models.Experience `gorm:"type:varchar(255)"` // Опыт работы
	Schedule        models.Schedule   `gorm:"type:varchar(255)"` // Режим работы
	Favorite        bool
	Pinned          bool
	Vacancies       []Vacancy
	Comments        []VacancyRequestComment `gorm:"foreignKey:VacancyRequestID"`
	ApprovalTasks   []ApprovalTask          `gorm:"foreignKey:RequestID"`
}

type VacancyRequestComment struct {
	ID               string
	VacancyRequestID string `gorm:"index"`
	Date             time.Time
	AuthorID         string
	Author           *SpaceUser `gorm:"foreignKey:AuthorID"`
	Comment          string
}

func (v *VacancyRequest) AfterDelete(tx *gorm.DB) (err error) {
	if v.ID == "" {
		return nil
	}
	tx.Clauses(clause.Returning{}).Where("request_id = ?", v.ID).Delete(&ApprovalTask{})
	return
}
