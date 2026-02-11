package models

type RbacFunc func(spaceID, userID string, role UserRole, path string) bool

type Module string

const (
	UsersModule          Module = "USERS"
	VacancyRequestModule Module = "VACANCY_REQUEST"
	VacancyModule        Module = "VACANCY"
	ApplicantModule      Module = "APPLICANT"
	AnalyticsModule      Module = "ANALYTICS"
	ProfileModule        Module = "PROFILE"
	CompanyProfileModule Module = "COMPANY_PROFILE"
	DictModule           Module = "DICT"
)

type Permission string

const (
	CreatePermission Permission = "CREATE"
	EditPermission   Permission = "EDIT"
	ViewPermission   Permission = "VIEW"
	ManagePermission Permission = "MANAGE"
	FlowPermission   Permission = "FLOW"
	StagesPermission Permission = "STAGES"
	TeamPermission   Permission = "TEAM"
	FilesPermission  Permission = "FILES"
	NotesPermission  Permission = "NOTES"
)
