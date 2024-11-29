package models

type UserRole string

const (
	SpaceAdminRole     UserRole = "SPACE_ADMIN_ROLE"
	SpaceUserRole      UserRole = "SPACE_USER_ROLE"
	UserRoleSuperAdmin UserRole = "SUPER_ADMIN"
)

var roleHumanName = map[UserRole]string{
	SpaceAdminRole:     "Администратор",
	SpaceUserRole:      "Пользователь",
	UserRoleSuperAdmin: "Суперадмин системы",
}

func (r UserRole) ToHuman() string {
	if human, exist := roleHumanName[r]; exist {
		return human
	}
	return string(r)

}

func (r UserRole) IsSpaceAdmin() bool {
	return r == SpaceAdminRole
}

const SystemUser = "Система"