package dbmodels

import (
	"hr-tools-backend/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	Employment      models.Employment `gorm:"type:varchar(255)"` // Занятость
	Experience      models.Experience `gorm:"type:varchar(255)"` // Опыт работы
	Schedule        models.Schedule   `gorm:"type:varchar(255)"` // Режим работы
	SelectionStages []SelectionStage
	VacancyTeam     []VacancyTeam
	HRSurvey        *HRSurvey
	Comments        []VacancyComment `gorm:"foreignKey:VacancyID"`
}

type HhData struct {
	HhID      string                  `gorm:"type:varchar(255)"`
	HhUri     string                  `gorm:"type:varchar(500)"`
	HhStatus  models.VacancyPubStatus `gorm:"type:varchar(255)"` // статус публикации
	HhReasons string                  `gorm:"type:varchar(500)"` // Расширенное описание статуса
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
	Favorite        bool
	Pinned          bool
	ResponsibleID   string
	ResponsibleUser *SpaceUser `gorm:"foreignKey:ResponsibleID"`
}

type VacancyComment struct {
	ID        string
	VacancyID string `gorm:"index"`
	Date      time.Time
	AuthorID  string
	Comment   string
}
