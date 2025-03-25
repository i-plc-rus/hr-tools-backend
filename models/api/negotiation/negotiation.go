package negotiationapimodels

import (
	"fmt"
	"hr-tools-backend/config"
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
	Comment           string                   `json:"comment"`            //коментарий к кандидату
	NegotiationDate   time.Time                `json:"negotiation_date"`   //дата отклика
	Salary            int                      `json:"salary"`             //ожидаемая зп
	Stage             string                   `json:"step"`               //этап
	StageTime         string                   `json:"step_time"`          // время на этапе
	Source            models.ApplicantSource   `json:"source"`             //источник
	NegotiationStatus models.NegotiationStatus `json:"negotiation_status"` // статус отклика
	PhotoUrl          string                   `json:"photo_url"`
	Age               int                      `json:"age"`                      // возраст
	ResumeTitle       string                   `json:"resume_title"`             //заголовок резюме
	Address           string                   `json:"address"`                  //Адрес кандидата
	VacancyAuthor     string                   `json:"vacancy_author,omitempty"` //Автор вакансии
	Citizenship       string                   `json:"citizenship"`              // Гражданство
	Gender            models.GenderType        `json:"gender"`                   // Пол кандидата
	Relocation        models.RelocationType    `json:"relocation"`               // Готовность к переезду
	Params            dbmodels.ApplicantParams `json:"params"`
	SurveyUrl         string                   `json:"survey_url"` // Ссылка на анкету для кандидата
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
		Source:            rec.Source,
		NegotiationStatus: rec.NegotiationStatus,
		ResumeTitle:       rec.ResumeTitle,
		Address:           rec.Address,
		Citizenship:       rec.Citizenship,
		Gender:            rec.Gender,
		Relocation:        rec.Relocation,
		PhotoUrl:          rec.PhotoUrl,
		Params:            rec.Params,
	}
	if rec.SelectionStage != nil {
		result.Stage = rec.SelectionStage.Name
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
	//TODO получение времени на шаге из истории действий
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
	if rec.ApplicantSurvey != nil {
		result.SurveyUrl = config.Conf.UIParams.SurveyPath + rec.ApplicantSurvey.ID
	}

	return result
}

type StatusData struct {
	Status models.NegotiationStatus `json:"status"`
}

type CommentData struct {
	Comment string `json:"comment"`
}
