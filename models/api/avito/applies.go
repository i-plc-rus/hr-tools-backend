package avitoapimodels

import "time"

type AppliesIDResponse struct {
	Applies []AppliesIDData `json:"applies"`
}

type AppliesIDData struct {
	ID        string `json:"id"`
	UpdatedAt string `json:"updated_at"` //"2022-03-21T12:37:41Z"
}

func (a AppliesIDData) GetUpdatedAt() (time.Time, bool) {
	if a.UpdatedAt == "" {
		return time.Time{}, false
	}
	date, err := time.Parse(time.RFC3339, a.UpdatedAt)
	return date, err == nil
}

type ApplicationRequest struct {
	IDs []string `json:"ids"`
}

type AppliesResponse struct {
	Applies []Applies `json:"applies"`
}

type Applies struct {
	ID        string              `json:"id"`
	CreatedAt string              `json:"created_at"`
	IsViewed  bool                `json:"is_viewed"`
	Contacts  ApplicationContacts `json:"contacts"`
	Applicant Applicant           `json:"applicant"`
	VacancyID int                 `json:"vacancy_id"`
}

func (a Applies) GetBirthDate() (time.Time, error) {
	if a.Applicant.Data.Birthday == "" {
		return time.Time{}, nil
	}
	date, err := time.Parse("2006-01-02", a.Applicant.Data.Birthday)
	if err != nil {
		return time.Time{}, err
	}
	return date, nil
}

type ApplicationContacts struct {
	Phones []ContactsPhoneValue `json:"phones"`
}

type ContactsPhoneValue struct {
	Value int `json:"value"`
}

type Applicant struct {
	Data     ApplicantData `json:"data"`
	ResumeID int           `json:"resume_id"`
	ID       string        `json:"id"`
}

type ApplicantData struct {
	Birthday    string            `json:"birthday"`
	Citizenship string            `json:"citizenship"`
	FullName    ApplicantFullName `json:"full_name"`
}

type ApplicantFullName struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
}
