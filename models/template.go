package models

type TemplateData struct {
	JobTitle            string
	ApplicantFIO        string
	ApplicantName       string
	ApplicantLastName   string
	ApplicantMiddleName string
	VacancyName         string
	CompanyName         string
	ApplicantSource     string
	VacancyLink         string
	SelfName            string
	CompanyDirectorName string
	CompanyAddress      string
	CompanyContact      string
	Files               TemplateFiles
}

type TemplateFiles struct {
	Logo  *File
	Stamp *File
	Sign  *File
}

type File struct {
	FileName    string
	ContentType string
	Body        []byte
}

type SalesTemplateData struct {
	OrganizationName string
	Inn              string
	Kpp              string
	OGRN             string
	DirectorName     string
	UserFIO          string
	UserPhoneNumber  string
	TextRequest      string
}
