package dbmodels

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"hr-tools-backend/models"
)

func (v *Vacancy) AfterDelete(tx *gorm.DB) (err error) {
	if v.ID == "" {
		return nil
	}
	tx.Clauses(clause.Returning{}).Where("vacancy_id = ?", v.ID).Delete(&Pinned{})
	tx.Clauses(clause.Returning{}).Where("vacancy_id = ?", v.ID).Delete(&Favorite{})
	return
}

type Vacancy struct {
	BaseSpaceModel
	Salary
	VacancyRequestID *string `gorm:"type:varchar(36)"`
	VacancyRequest   *VacancyRequest
	AuthorID         string
	Author           *SpaceUser `gorm:"foreignKey:AuthorID"`
	Space            *Space
	CompanyID        *string `gorm:"type:varchar(36);index:idx_company"`
	Company          *Company
	DepartmentID     *string `gorm:"type:varchar(36)"`
	Department       *Department
	JobTitleID       *string `gorm:"type:varchar(36)"`
	JobTitle         *JobTitle
	CityID           *string `gorm:"type:varchar(36)"`
	City             *City
	CompanyStructID  *string `gorm:"type:varchar(36)"`
	CompanyStruct    *CompanyStruct
	VacancyName      string `gorm:"type:varchar(255)"`
	OpenedPositions  int
	Urgency          models.VRUrgency       `gorm:"type:varchar(100)"`
	RequestType      models.VRType          `gorm:"type:varchar(100)"`
	SelectionType    models.VRSelectionType `gorm:"type:varchar(100)"`
	PlaceOfWork      string                 `gorm:"type:varchar(255)"`
	ChiefFio         string                 `gorm:"type:varchar(255)"`
	Requirements     string
	Status           models.VacancyStatus
	HhData
	AvitoData
	Employment models.Employment `gorm:"type:varchar(255)"` // Занятость
	Experience models.Experience `gorm:"type:varchar(255)"` // Опыт работы
	Schedule   models.Schedule   `gorm:"type:varchar(255)"` // Режим работы
}

type HhData struct {
	HhID     string                  `gorm:"type:varchar(255)"`
	HhUri    string                  `gorm:"type:varchar(500)"`
	HhStatus models.VacancyPubStatus `gorm:"type:varchar(255)"` // статус публикации
}

type AvitoData struct {
	AvitoPublishID string                  `gorm:"type:varchar(255)"` // ид публикации
	AvitoID        int                     // ид вакансии
	AvitoUri       string                  `gorm:"type:varchar(500)"` // урл вакансии на сайте авито
	AvitoStatus    models.VacancyPubStatus `gorm:"type:varchar(255)"` // статус публикации
	AvitoReasons   string                  `gorm:"type:varchar(500)"` // Расширенное описание статуса
}

type Salary struct {
	From     int `gorm:"column:salary_from"`
	To       int `gorm:"column:salary_to"`
	ByResult int `gorm:"column:salary_result"`
	InHand   int `gorm:"column:salary_in_hand"`
}

type VacancyExt struct {
	Vacancy
	Favorite bool
	Pinned   bool
}

type VacancySort struct {
	CreatedAtDesc bool `json:"created_at_desc"` // порядок сортировки false = ASC/ true = DESC
}

type VacancyFilter struct {
	VacancyRequestID string                 `json:"request_id"`
	Favorite         bool                   `json:"favorite"`
	Search           string                 `json:"search"`
	Statuses         []models.VacancyStatus `json:"statuses"`
	CityID           string                 `json:"city_id"`
	DepartmentID     string                 `json:"department_id"`
	SelectionType    models.VRSelectionType `json:"selection_type"`
	RequestType      models.VRType          `json:"request_type"`
	Urgency          models.VRUrgency       `json:"urgency"`
	AuthorID         string                 `json:"author_id"`
	RequestAuthorID  string                 `json:"request_author_id"`
	Sort             VacancySort            `json:"sort"`
}
