package models

type UserRole string

var AllAvailableRoles = []UserRole{AdminRole, HRRole, ManagerRole, SpecialistRole}
const (
	AdminRole          UserRole = "ADMIN"
	HRRole             UserRole = "HR"
	ManagerRole        UserRole = "MANAGER"
	SpecialistRole     UserRole = "SPECIALIST"
	UserRoleSuperAdmin UserRole = "SUPER_ADMIN"
	AllRoles           UserRole = "ALL"
)

var roleHumanName = map[UserRole]string{
	UserRoleSuperAdmin: "Суперадмин системы",
	AdminRole:          "Администратор",
	HRRole:             "HR",
	ManagerRole:        "Руководитель",
	SpecialistRole:     "Специалист",
}

func (r UserRole) ToHuman() string {
	if human, exist := roleHumanName[r]; exist {
		return human
	}
	return string(r)

}

func (r UserRole) IsSpaceAdmin() bool {
	return r == AdminRole
}

const SystemUser = "Система"

type UserStatus string

const (
	SpaceWorkingStatus    UserStatus = "WORKING"
	SpaceOnVacationStatus UserStatus = "VACATION"
	SpaceDismissedStatus  UserStatus = "DISMISSED"
)

var userStatusHumanName = map[UserStatus]string{
	SpaceWorkingStatus:    "Работает",
	SpaceOnVacationStatus: "В отпуске",
	SpaceDismissedStatus:  "Уволен",
}

func (r UserStatus) ToHuman() string {
	if human, exist := userStatusHumanName[r]; exist {
		return human
	}
	return string(r)
}
