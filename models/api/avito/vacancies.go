package avitoapimodels

import (
	"github.com/pkg/errors"
	"hr-tools-backend/models"
)

type VacancyStatusInfo struct {
	Status string `json:"url"`
	Url    string `json:"url"`
}

type VacancyAttach struct {
	ID int `json:"id"` // идентификатор вакансии в виде: 3364561973
}

func (v VacancyAttach) Validate() error {
	if v.ID == 0 {
		return errors.New("не указана ссылка на вакансию")
	}
	return nil
}

type VacancyPubRequest struct {
	ApplyProcessing ApplyProcessing   `json:"apply_processing"`
	BillingType     string            `json:"billing_type"` //Enum: "package" "single" "packageOrSingle"
	BusinessArea    int               `json:"business_area"`
	Description     string            `json:"description"`
	Employment      models.Employment `json:"employment"`
	Experience      models.Experience `json:"experience"`
	Location        Location          `json:"location"`
	Schedule        models.Schedule   `json:"schedule"`
	Title           string            `json:"title"`
	SalaryRange     *SalaryRange      `json:"salary_range,omitempty"`
}

type Location struct {
	Address LocationAddress `json:"address"`
}

type LocationAddress struct {
	Locality string `json:"locality"`
}

type ApplyType string

const (
	ApplyTypeWithResume    ApplyType = "only_with_resume"
	ApplyTypeWithAssistant ApplyType = "with_assistant"
)

type ApplyProcessing struct {
	ApplyType ApplyType `json:"apply_type"`
}

type VacancyPubResponse struct {
	ID string `json:"id"`
}

type StatusRequest struct {
	IDs []string `json:"ids"`
}

type StatusResponse []StatusItem

type StatusItem struct {
	ID         string        `json:"id"`
	LastAction interface{}   `json:"last_action"`
	Vacancy    VacancyStatus `json:"vacancy"`
}

type VacancyStatus struct {
	ID               string      `json:"id"`
	ModerationStatus string      `json:"moderation_status"`
	Status           string      `json:"status"`
	Url              string      `json:"url"`
	Reasons          interface{} `json:"reasons"`
}

func (v VacancyStatus) GetPubStatus() models.VacancyPubStatus {
	switch v.Status {
	case "closed", "expired", "archived":
		return models.VacancyPubStatusClosed
	case "created", "new":
		return models.VacancyPubStatusModeration
	case "blocked", "rejected":
		return models.VacancyPubStatusRejected
	default:
		return models.VacancyPubStatusPublished
	}
}

type Contacts struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phones Phone  `json:"phones"`
}

type Phone struct {
	City    string `json:"city"`
	Country string `json:"country"`
	Number  string `json:"number"`
}

type SalaryRange struct {
	From int `json:"from,omitempty"`
	To   int `json:"to,omitempty"`
}

type DictItem struct {
	ID string `json:"id"`
}

type VacancyInfo struct {
	ID       string `json:"id"`
	Url      string `json:"url"`
	IsActive bool   `json:"is_active"`
}
