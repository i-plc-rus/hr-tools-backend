package negotiationapimodels

import (
	"fmt"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"strings"
	"time"
)

type NegotiationView struct {
	ID                string                   `json:"id"`
	FIO               string                   `json:"fio"`
	Phone             string                   `json:"phone"`
	Email             string                   `json:"email"`
	Comment           string                   `json:"comment"`
	NegotiationDate   time.Time                `json:"negotiation_date"`
	Salary            int                      `json:"salary"`
	Stage             string                   `json:"step"`
	StageTime         string                   `json:"step_time"`
	Source            models.ApplicantSource   `json:"source"`
	NegotiationStatus models.NegotiationStatus `json:"negotiation_status"`
	PhotoUrl          string                   `json:"photo_url"`
	Age               int                      `json:"age"`
	ResumeTitle       string                   `json:"resume_title"`
	Address           string                   `json:"address"`
	VacancyAuthor     string                   `json:"vacancy_author"`
}

func NegotiationConvertExt(rec dbmodels.ApplicantExt) NegotiationView {
	result := NegotiationConvert(rec.Applicant)
	result.VacancyAuthor = strings.TrimSpace(fmt.Sprintf("%v %v", rec.AuthorLastName, rec.AuthorFirstName))
	return result
}

func NegotiationConvert(rec dbmodels.Applicant) NegotiationView {
	result := NegotiationView{
		ID:                rec.ID,
		Phone:             rec.Phone,
		Email:             rec.Email,
		Comment:           rec.Comment,
		NegotiationDate:   rec.NegotiationDate,
		Salary:            rec.Salary,
		Stage:             "Откликнулся",
		Source:            rec.Source,
		NegotiationStatus: rec.NegotiationStatus,
		ResumeTitle:       rec.ResumeTitle,
		Address:           rec.Address,
	}
	if !rec.BirthDate.IsZero() {
		difference := time.Now().Sub(rec.BirthDate)
		result.Age = int(difference.Hours() / 24 / 365)
	}

	if result.NegotiationStatus == "" {
		result.NegotiationStatus = "Выберите статус"
	}
	fio := strings.TrimSpace(fmt.Sprintf("%v %v", rec.LastName, rec.FirstName))
	fio = strings.TrimSpace(fmt.Sprintf("%v %v", fio, rec.MiddleName))
	result.FIO = fio
	toTime := time.Now()
	if !rec.NegotiationAcceptDate.IsZero() {
		toTime = rec.NegotiationAcceptDate
	}
	sec := toTime.Unix() - rec.NegotiationDate.Unix()
	minutes := sec / 60
	hours := minutes / 60
	minutes = minutes - hours*60
	if hours > 0 {
		result.StageTime = fmt.Sprintf("%v час", hours)
	}
	if minutes > 0 {
		if result.StageTime != "" {
			result.StageTime += " "
		}
		result.StageTime = result.StageTime + fmt.Sprintf("%v мин", minutes)
	}

	return result
}

type StatusData struct {
	Status models.NegotiationStatus `json:"status"`
}

type CommentData struct {
	Comment string `json:"comment"`
}
