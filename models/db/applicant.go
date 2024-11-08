package dbmodels

import (
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	"time"
)

type Applicant struct {
	BaseSpaceModel
	VacancyID             string   `gorm:"type:varchar(36)"`
	Vacancy               *Vacancy `gorm:"foreignKey:VacancyID"`
	NegotiationID         string   `gorm:"type:varchar(255);index:idx_negotiation"` // ид отклика во внешней системе
	ResumeID              string   `gorm:"index;type:varchar(255)"`                 // ид резюме во внешней системе
	ResumeTitle           string
	Source                models.ApplicantSource `gorm:"index:idx_negotiation"`
	NegotiationDate       time.Time
	NegotiationAcceptDate time.Time
	Status                models.ApplicantStatus
	NegotiationStatus     models.NegotiationStatus
	FirstName             string `gorm:"type:varchar(255)"`
	LastName              string `gorm:"type:varchar(255)"`
	MiddleName            string `gorm:"type:varchar(255)"`
	Phone                 string `gorm:"type:varchar(255)"`
	Email                 string `gorm:"type:varchar(255)"`
	Salary                int
	Address               string
	BirthDate             time.Time
	Citizenship           string                `gorm:"type:varchar(255)"`
	Gender                string                `gorm:"type:varchar(50)"`
	LanguageLevel         string                `gorm:"type:varchar(100)"`
	Relocation            models.RelocationType `gorm:"type:varchar(100)"`
	TotalExperience       int                   //опыт работ в месяцах
	Comment               string
	//todo прочие поля резюме
}

type ApplicantExt struct {
	Applicant
	AuthorFirstName string
	AuthorLastName  string
}

func (a Applicant) IsAllowStatusChange(newStatus models.NegotiationStatus) (bool, error) {
	if newStatus != models.NegotiationStatusWait &&
		newStatus != models.NegotiationStatusRejected &&
		newStatus != models.NegotiationStatusAccepted {
		return false, errors.New("неизвестный статус")
	}
	if a.NegotiationStatus == newStatus {
		return false, nil
	}
	if a.Status == models.ApplicantStatusInProcess {
		return false, errors.New("смена статуса отклика недоступна, кандидат в процессе рассмотрения")
	}
	if a.Status == models.ApplicantStatusRejected {
		return false, errors.New("смена статуса отклика недоступна, кандидат уже отклонен")
	}
	if a.NegotiationStatus == models.NegotiationStatusAccepted {
		return false, errors.New("смена статуса отклика недоступна, отклик уже принят")
	}
	return true, nil
}

type NegotiationFilter struct {
	VacancyID string `json:"vacancy_id"`
	Search    string `json:"search"`
}

func (n NegotiationFilter) Validate() error {
	if n.VacancyID == "" {
		return errors.New("не указан идентификатор вакансии")
	}
	return nil
}
