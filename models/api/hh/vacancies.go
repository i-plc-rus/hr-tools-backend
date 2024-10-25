package hhapimodels

type VacancyPubRequest struct {
	Description       string     `json:"description"`
	Name              string     `json:"name"`
	Area              DictItem   `json:"area"`
	Employment        *DictItem  `json:"employment,omitempty"` //Тип занятости
	Schedule          *DictItem  `json:"schedule,omitempty"`   // График работы
	Experience        *DictItem  `json:"experience,omitempty"` // Опыт работы
	Salary            *Salary    `json:"salary,omitempty"`
	Contacts          *Contacts  `json:"contacts,omitempty"`
	ProfessionalRoles []DictItem `json:"professional_roles"`
	BillingType       DictItem   `json:"billing_type"`
	Type              DictItem   `json:"type"`
}

type VacancyResponse struct {
	ID string `json:"id"`
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

type Salary struct {
	Currency string `json:"currency"`
	From     int    `json:"from,omitempty"`
	To       int    `json:"to,omitempty"`
	Gross    bool   `json:"gross"`
}

type DictItem struct {
	ID string `json:"id"`
}
