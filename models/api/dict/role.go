package dictapimodels

import "hr-tools-backend/models"

func GetRoles() []RoleView {
	return []RoleView{
		GetRole(models.AdminRole),
		GetRole(models.HRRole),
		GetRole(models.ManagerRole),
		GetRole(models.SpecialistRole),
	}
}

func GetRole(role models.UserRole) RoleView {
	return RoleView{
		Code: string(role),
		Name: role.ToHuman(),
	}
}

type RoleView struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
