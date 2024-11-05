package hhapimodels

type NegotiationCollections struct {
	Collections []NegotiationCollection `json:"collections"`
}
type NegotiationCollection struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Url         string `json:"url"`
}

type NegotiationResponse struct {
	Found   int           `json:"found"`
	Page    int           `json:"page"`
	Pages   int           `json:"pages"`
	PerPage int           `json:"per_page"`
	Items   []Negotiation `json:"items"`
}

type Negotiation struct {
	ID          string            `json:"id"`
	CreatedAt   string            `json:"created_at"`
	Url         string            `json:"url"`
	ChatID      int               `json:"chat_id"`
	MessagesUrl string            `json:"messages_url"`
	Source      string            `json:"source"`
	Resume      NegotiationResume `json:"resume"`
}

type NegotiationResume struct {
	ID           string       `json:"id"`
	AlternateUrl string       `json:"alternate_url"`
	Title        string       `json:"title"`
	FirstName    string       `json:"first_name"`
	LastName     string       `json:"last_name"`
	MiddleName   string       `json:"middle_name"`
	Photo        string       `json:"photo"`
	ResumeUrl    string       `json:"url"`
	Salary       ResumeSalary `json:"salary"`
}

type NegotiationReadRequest struct {
	TopicID string `json:"topic_id"`
}
