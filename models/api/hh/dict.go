package hhapimodels

type PreferredContactType string

const (
	ContactTypeHome  PreferredContactType = "home"
	ContactTypeWork  PreferredContactType = "work"
	ContactTypeCell  PreferredContactType = "cell"
	ContactTypeEmail PreferredContactType = "email"
)

type DictData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Area struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	Name     string `json:"name"`
	Areas    []Area `json:"areas"`
}

type EmployerAddresseses struct {
	Items []EmployerAddresses `json:"items"`
}

type EmployerAddresses struct {
	City     string `json:"city"`
	Street   string `json:"street"`
	Building string `json:"building"`
	Raw      string `json:"raw"`
	ID       string `json:"id"`
}
