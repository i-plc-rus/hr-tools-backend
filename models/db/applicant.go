package dbmodels

import (
	"hr-tools-backend/models"
	"time"
)

type Applicant struct {
	BaseSpaceModel
	VacancyID       string   `gorm:"type:varchar(36)"`
	Vacancy         *Vacancy `gorm:"foreignKey:VacancyID"`
	NegotiationID   string   `gorm:"type:varchar(255);index:idx_negotiation"` // ид отклика во внешней системе
	ResumeID        string   `gorm:"index;type:varchar(255)"`                 // ид резюме во внешней системе
	ResumeTitle     string
	Source          models.ApplicantSource `gorm:"index:idx_negotiation"`
	NegotiationDate time.Time
	Status          models.ApplicantStatus
	FirstName       string `gorm:"type:varchar(255)"`
	LastName        string `gorm:"type:varchar(255)"`
	MiddleName      string `gorm:"type:varchar(255)"`
	Phone           string `gorm:"type:varchar(255)"`
	Email           string `gorm:"type:varchar(255)"`
	Salary          int
	Address         string
	BirthDate       time.Time
	Citizenship     string                `gorm:"type:varchar(255)"`
	Gender          string                `gorm:"type:varchar(50)"`
	LanguageLevel   string                `gorm:"type:varchar(100)"`
	Relocation      models.RelocationType `gorm:"type:varchar(100)"`
	TotalExperience int                   //опыт работ в месяцах
	Comment         string
	//todo прочие поля резюме
}
