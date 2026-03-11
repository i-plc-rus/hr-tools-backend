package hhapimodels

import (
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	"strings"
)

type VacancyAttach struct {
	URL string `json:"url"` // ссылка на вакансию в виде: https://izhevsk.hh.ru/vacancy/108984166
}

func (v VacancyAttach) Validate() error {
	id, err := v.GetID()
	if err != nil {
		return err
	}
	if id == "" {
		return errors.New("не указана ссылка на вакансию")
	}
	return nil
}

func (v VacancyAttach) GetID() (string, error) {
	//варианты ссылок
	//https://kazan.hh.ru/vacancy/108984166?from=main&utm_source=headhunter&utm_medium=main_page_bottom&utm_campaign=vacancy_of_the_day_to
	//https://izhevsk.hh.ru/vacancy/108984166
	parts := strings.Split(v.URL, "hh.ru/vacancy/")
	if len(parts) != 2 {
		return "", errors.New("некорректная ссылка на вакансию")
	}
	id, _, _ := strings.Cut(parts[1], "?")
	return id, nil
}

type VacancyPubRequest struct {
	// required
	Area              DictItem   `json:"area"`
	BillingType       DictItem   `json:"billing_type"`
	Description       string     `json:"description"`
	Name              string     `json:"name"`
	Type              DictItem   `json:"type"`
	ProfessionalRoles []DictItem `json:"professional_roles"`
	// optional
	EmploymentFrom *DictItem    `json:"employment_form,omitempty"` //Тип занятости
	Schedule       *DictItem    `json:"schedule,omitempty"`        // График работы
	Experience     *DictItem    `json:"experience,omitempty"`      // Опыт работы
	SalaryRange    *SalaryRange `json:"salary_range,omitempty"`
	Contacts       *Contacts    `json:"contacts,omitempty"`
	AllowMessages  bool         `json:"allow_messages"` // Разрешение сообщений
}

type VacancyResponse struct {
	ID string `json:"id"`
}

type Contacts struct {
	Name   string  `json:"name"`
	Email  string  `json:"email"`
	Phones []Phone `json:"phones"`
}

type Phone struct {
	City    string `json:"city"`
	Country string `json:"country"`
	Number  string `json:"number"`
}

// Deprecated: use SalaryRange
type Salary struct {
	Currency string `json:"currency"`
	From     int    `json:"from,omitempty"`
	To       int    `json:"to,omitempty"`
	Gross    bool   `json:"gross"`
}

type SalaryRange struct {
	// required
	Currency string   `json:"currency"`
	Gross    bool     `json:"gross"`
	Mode     DictItem `json:"mode"`
	// optional
	From *int `json:"from,omitempty"`
	To   *int `json:"to,omitempty"`
}

type DictItem struct {
	ID string `json:"id"`
}

type VacancyInfo struct {
	ID           string     `json:"id"`
	Approved     bool       `json:"approved"`
	Archived     bool       `json:"archived"`
	AlternateUrl string     `json:"alternate_url"`
	Employer     MeEmployer `json:"employer"`
}

func (v VacancyInfo) GetPubStatus() models.VacancyPubStatus {
	if v.Archived {
		return models.VacancyPubStatusClosed
	}
	if v.Approved {
		return models.VacancyPubStatusPublished
	}

	return models.VacancyPubStatusModeration
}

func (v VacancyPubRequest) ToDraft(recID string) VacancyDraftRequest {
	result := VacancyDraftRequest{
		ClosedForApplicants:     false,
		VacancyProperties:       []VacancyProperty{{PropertyType: "HH_STANDARD"}},
		AcceptHandicapped:       true,
		AcceptIncompleteResumes: false,
		AcceptLaborContract:     true,
		AcceptTemporary:         true,
		Address:                 nil,
		AgeRestriction:          &DictItem{ID: "AGE_16_PLUS"},
		AllowMessages:           true,
		Areas:                   []DictItem{v.Area},
		AutoResponse:            AutoResponse{AcceptAutoResponse: false},
		BrandedTemplate:         nil,
		CivilLawContracts:       &[]DictItem{{ID: "INDIVIDUAL_PERSON"}},
		Code:                    nil,
		Contacts:                v.Contacts,
		Department:              nil,
		Description:             v.Description,
		DriverLicenseTypes:      nil,
		Employment:              nil,
		EmploymentFrom:          v.EmploymentFrom,
		Experience:              v.Experience,
		Internship:              false,
		KeySkills:               nil,
		Languages:               nil,
		Name:                    v.Name,
		NightShifts:             false,
		ProfessionalRoles:       &v.ProfessionalRoles,
		ResponseLetterRequired:  false,
		ResponseNotifications:   true,
		SalaryRange:             v.SalaryRange,
		Schedule:                v.Schedule,
		Test:                    nil,
		WithZp:                  false,
		WorkFormat:              &DictItem{ID: "ON_SITE"},
		WorkScheduleByDays:      []DictItem{{ID: "WEEKEND"}},
		WorkingDays:             nil,
		WorkingHours:            []DictItem{{ID: "HOURS_8"}},
		WorkingTimeIntervals:    nil,
		WorkingTimeModes:        nil,
	}

	if recID != "" {
		result.Code = &recID
	}
	return result
}
