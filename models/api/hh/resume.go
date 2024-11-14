package hhapimodels

import (
	"hr-tools-backend/models"
	"time"
)

type ResumeResponse struct {
	Age                       int             `json:"age"`
	AlternateUrl              string          `json:"alternate_url"`
	Area                      DictData        `json:"area"`
	BirthDate                 string          `json:"birth_date"` //ГГГГ-ММ-ДД
	BusinessTripReadiness     DictData        `json:"business_trip_readiness"`
	JobSearchStatusesEmployer DictData        `json:"job_search_statuses_employer"`
	Citizenship               []DictData      `json:"citizenship"`
	Contact                   []Contact       `json:"contact"`
	CreatedAt                 string          `json:"created_at"`
	DriverLicenseTypes        []DictData      `json:"driver_license_types"`
	Education                 Education       `json:"education"`
	Employments               []DictData      `json:"employments"`
	Experience                []interface{}   `json:"experience"`
	Gender                    DictData        `json:"gender"`
	ID                        string          `json:"id"`
	Language                  []Language      `json:"language"`
	FirstName                 string          `json:"first_name"`
	LastName                  string          `json:"last_name"`
	MiddleName                string          `json:"middle_name"`
	Photo                     Photo           `json:"photo"`
	Portfolio                 []interface{}   `json:"portfolio"`
	ProfessionalRoles         []interface{}   `json:"professional_roles"`
	Recommendation            []interface{}   `json:"recommendation"`
	Relocation                Relocation      `json:"relocation"`
	Salary                    ResumeSalary    `json:"salary"`
	Schedules                 []DictData      `json:"schedules"`
	SkillSet                  []string        `json:"skill_set"`
	Title                     string          `json:"title"`
	TotalExperience           TotalExperience `json:"total_experience"`
}

func (r ResumeResponse) GetBirthDate() (time.Time, error) {
	if r.BirthDate == "" {
		return time.Time{}, nil
	}
	date, err := time.Parse("2006-01-02", r.BirthDate)
	if err != nil {
		return time.Time{}, err
	}
	return date, nil
}

type Contact struct {
	Comment   string      `json:"comment"`
	Preferred bool        `json:"preferred"`
	Type      ContactType `json:"type"`
	Value     interface{} `json:"value"`
}

type ContactType struct {
	ID   PreferredContactType `json:"id"`
	Name string               `json:"name"`
}

type PhoneStruct struct {
	Formatted string `json:"formatted"`
}

type ResumeSalary struct {
	Currency string `json:"currency"`
	Amount   int    `json:"amount"`
}

type TotalExperience struct {
	Months int `json:"months"`
}

type Relocation struct {
	Type RelocationType `json:"type"`
}

type RelocationType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (r Relocation) GetRelocationType() models.RelocationType {
	switch r.Type.ID {
	case "no_relocation":
		return models.RelocationTypeNo
	case "relocation_possible":
		return models.RelocationTypeYes
	case "relocation_desirable":
		return models.RelocationTypeWant
	}
	return ""
}

type Language struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Level LanguageLevel `json:"level"`
}

type LanguageLevel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Photo struct {
	Medium string `json:"medium"`
	Small  string `json:"small"`
}

type Education struct {
	Level      DictData              `json:"level"`
	Additional []AdditionalEducation `json:"additional"`
}

type AdditionalEducation struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
