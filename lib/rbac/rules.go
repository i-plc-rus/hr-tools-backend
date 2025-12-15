package rbac

import (
	vacancyreqhandler "hr-tools-backend/lib/vacancy-req"
	"hr-tools-backend/models"
)

var (
	AdminHrManagerRoleSet         = []models.UserRole{models.AdminRole, models.HRRole, models.ManagerRole}
	AdminHrRoleSet                = []models.UserRole{models.AdminRole, models.HRRole}
	AdminManagerRoleSet           = []models.UserRole{models.AdminRole, models.ManagerRole}
	AdminManagerSpecialistRoleSet = []models.UserRole{models.AdminRole, models.ManagerRole, models.SpecialistRole}
	AllRoles                      = []models.UserRole{models.AdminRole, models.HRRole, models.ManagerRole, models.SpecialistRole}
)

func (i *impl) initRules() {
	i.addUsersRbac()
	i.addVacancyReqestRbac()
	i.addVacancyRbac()
	i.applicant()
	i.analytics()
	i.profile()
	i.companyProfile()
}

func (i *impl) addUsersRbac() {
	//VIEW
	i.RegisterRule(models.UsersModule, models.ViewPermission, AllRoles, "/api/v1/users/list [post]", nil)
	i.RegisterRule(models.UsersModule, models.ViewPermission, AllRoles, "/api/v1/users/{id} [get]", nil)
	//MANAGE
	i.RegisterRule(models.UsersModule, models.ManagePermission, AdminHrManagerRoleSet, "/api/v1/users [post]", nil)
	i.RegisterRule(models.UsersModule, models.ManagePermission, AdminHrManagerRoleSet, "/api/v1/users/{id} [delete]", nil)
	i.RegisterRule(models.UsersModule, models.ManagePermission, AdminHrManagerRoleSet, "/api/v1/users/{id} [put]", nil)
}

func (i *impl) addVacancyReqestRbac() {
	//VIEW
	i.RegisterRule(models.VacancyRequestModule, models.ViewPermission, AllRoles, "/api/v1/space/vacancy_request/list [post]", nil)
	i.RegisterRule(models.VacancyRequestModule, models.ViewPermission, AllRoles, "/api/v1/space/vacancy_request/{id} [get]", nil)
	// CREATE
	i.RegisterRule(models.VacancyRequestModule, models.CreatePermission, AdminManagerSpecialistRoleSet, "/api/v1/space/vacancy_request [post]", nil)
	// CREATE/EDIT + только к своей заявке
	selfAllow := vacancyreqhandler.Instance.GetRbacSelfAllow()
	i.RegisterRule(models.VacancyRequestModule, models.EditPermission, AdminManagerSpecialistRoleSet, "/api/v1/space/vacancy_request/{id} [put]", selfAllow)
	i.RegisterRule(models.VacancyRequestModule, models.EditPermission, AdminManagerSpecialistRoleSet, "/api/v1/space/vacancy_request/{id} [delete]", selfAllow)
	i.RegisterRule(models.VacancyRequestModule, models.EditPermission, AdminManagerSpecialistRoleSet, "/api/v1/space/vacancy_request/{id}/on_create [put]", selfAllow)
	i.RegisterRule(models.VacancyRequestModule, models.EditPermission, AdminManagerSpecialistRoleSet, "/api/v1/space/vacancy_request/{id}/on_approval [put]", selfAllow)
	i.RegisterRule(models.VacancyRequestModule, models.EditPermission, AdminManagerSpecialistRoleSet, "/api/v1/space/vacancy_request/{id}/cancel [put]", selfAllow)
	i.RegisterRule(models.VacancyRequestModule, models.EditPermission, AdminHrRoleSet, "/api/v1/space/vacancy_request/{id}/publish [put]", nil)
	i.RegisterRule(models.VacancyRequestModule, models.EditPermission, AdminManagerSpecialistRoleSet, "/api/v1/space/vacancy_request/{id}/approvals [put]", selfAllow)
	//FLOW
	flowAllow := vacancyreqhandler.Instance.GetRbacFlowAllow()
	i.RegisterRule(models.VacancyRequestModule, models.FlowPermission, AdminHrRoleSet, "/api/v1/space/vacancy_request/{id}/approvals/{taskId}/approve [post]", flowAllow)
	i.RegisterRule(models.VacancyRequestModule, models.FlowPermission, AdminHrRoleSet, "/api/v1/space/vacancy_request/{id}/approvals/{taskId}/request_changes [post]", flowAllow)
	i.RegisterRule(models.VacancyRequestModule, models.FlowPermission, AdminHrRoleSet, "/api/v1/space/vacancy_request/{id}/approvals/{taskId}/reject [post]", flowAllow)
}

func (i *impl) addVacancyRbac() {
	// VIEW
	i.RegisterRule(models.VacancyModule, models.ViewPermission, AllRoles, "/api/v1/space/vacancy/list [post]", nil)
	i.RegisterRule(models.VacancyModule, models.ViewPermission, AllRoles, "/api/v1/space/vacancy/{id} [get]", nil)
	//CREATE/EDIT
	i.RegisterRule(models.VacancyModule, models.CreatePermission, AdminHrRoleSet, "/api/v1/space/vacancy [post]", nil)
	i.RegisterRule(models.VacancyModule, models.EditPermission, AdminHrRoleSet, "/api/v1/space/vacancy/{id} [put]", nil)
	i.RegisterRule(models.VacancyModule, models.EditPermission, AdminHrRoleSet, "/api/v1/space/vacancy/{id} [delete]", nil)
	i.RegisterRule(models.VacancyModule, models.EditPermission, AdminHrRoleSet, "/api/v1/space/vacancy/{id}/change_status [put]", nil)
	//STAGES
	i.RegisterRule(models.VacancyModule, models.StagesPermission, AdminHrRoleSet, "/api/v1/space/vacancy/{id}/stage/change_order [put]", nil)
	i.RegisterRule(models.VacancyModule, models.StagesPermission, AdminHrRoleSet, "/api/v1/space/vacancy/{id}/stage [post]", nil)
	i.RegisterRule(models.VacancyModule, models.StagesPermission, AdminHrRoleSet, "/api/v1/space/vacancy/{id}/stage [delete]", nil)
	//TEAM
	i.RegisterRule(models.VacancyModule, models.TeamPermission, AdminHrRoleSet, "/api/v1/space/vacancy/{id}/team/{user_id}/invite [put]", nil)
	i.RegisterRule(models.VacancyModule, models.TeamPermission, AdminHrRoleSet, "/api/v1/space/vacancy/{id}/team/{user_id}/exclude [put]", nil)
	i.RegisterRule(models.VacancyModule, models.TeamPermission, AdminHrRoleSet, "/api/v1/space/vacancy/{id}/team/{user_id}/set_as_responsible [put]", nil)
}

func (i *impl) applicant() {
	// VIEW
	i.RegisterRule(models.ApplicantModule, models.ViewPermission, AdminHrManagerRoleSet, "/api/v1/space/applicant/list [post]", nil)
	i.RegisterRule(models.ApplicantModule, models.ViewPermission, AdminHrManagerRoleSet, "/api/v1/space/applicant/{id} [get]", nil)

	//EDIT
	i.RegisterRule(models.ApplicantModule, models.EditPermission, AdminHrRoleSet, "/api/v1/space/applicant [post]", nil)
	i.RegisterRule(models.ApplicantModule, models.EditPermission, AdminHrRoleSet, "/api/v1/space/applicant/{id} [put]", nil)
	i.RegisterRule(models.ApplicantModule, models.EditPermission, AdminHrRoleSet, "/api/v1/space/applicant/{id}/change_stage [put]", nil)
	i.RegisterRule(models.ApplicantModule, models.EditPermission, AdminHrRoleSet, "/api/v1/space/applicant/{id}/reject [put]", nil)
	i.RegisterRule(models.ApplicantModule, models.EditPermission, AdminHrRoleSet, "/api/v1/space/applicant/multi-actions/reject [put]", nil)
	i.RegisterRule(models.ApplicantModule, models.EditPermission, AdminHrRoleSet, "/api/v1/space/applicant/multi-actions/change_stage [put]", nil)
	//FILES/NOTES
	i.RegisterRule(models.ApplicantModule, models.FilesPermission, AdminHrRoleSet, "/api/v1/space/applicant/{id}/upload-resume [post]", nil)
	i.RegisterRule(models.ApplicantModule, models.FilesPermission, AdminHrRoleSet, "/api/v1/space/applicant/{id}/upload-doc [post]", nil)
	i.RegisterRule(models.ApplicantModule, models.FilesPermission, AdminHrRoleSet, "/api/v1/space/applicant/{id}/upload-photo [post]", nil)
	i.RegisterRule(models.ApplicantModule, models.FilesPermission, AdminHrRoleSet, "/api/v1/space/applicant/{id}/photo [delete]", nil)
	i.RegisterRule(models.ApplicantModule, models.FilesPermission, AdminHrRoleSet, "/api/v1/space/applicant/{id}/resume [delete]", nil)
	i.RegisterRule(models.ApplicantModule, models.FilesPermission, AdminHrRoleSet, "/api/v1/space/applicant/doc/{id} [delete]", nil)
	i.RegisterRule(models.ApplicantModule, models.NotesPermission, AdminHrRoleSet, "/api/v1/space/applicant/{id}/note [put]", nil)
}

func (i *impl) analytics() {
	// VIEW
	i.RegisterRule(models.AnalyticsModule, models.ViewPermission, AdminHrManagerRoleSet, "/api/v1/space/analytics/source [put]", nil)
	i.RegisterRule(models.AnalyticsModule, models.ViewPermission, AdminHrManagerRoleSet, "/api/v1/space/analytics/source_export [put]", nil)
}

func (i *impl) profile() {
	// EDIT
	i.RegisterRule(models.ProfileModule, models.EditPermission, AllRoles, "/api/v1/user_profile [get]", nil)
	i.RegisterRule(models.ProfileModule, models.EditPermission, AllRoles, "/api/v1/user_profile [put]", nil)
	i.RegisterRule(models.ProfileModule, models.EditPermission, AllRoles, "/api/v1/user_profile/change_password [put]", nil)
	i.RegisterRule(models.ProfileModule, models.EditPermission, AllRoles, "/api/v1/user_profile/photo [post]", nil)
	i.RegisterRule(models.ProfileModule, models.EditPermission, AllRoles, "/api/v1/user_profile/photo [get]", nil)
}

func (i *impl) companyProfile() {
	// VIEW
	i.RegisterRule(models.CompanyProfileModule, models.ViewPermission, AllRoles, "/api/v1/space/profile [get]", nil)
	i.RegisterRule(models.CompanyProfileModule, models.ViewPermission, AllRoles, "/api/v1/space/profile/photo [get]", nil)
	// EDIT
	i.RegisterRule(models.CompanyProfileModule, models.EditPermission, AdminManagerRoleSet, "/api/v1/space/profile [put]", AllowByRoleFunc(AdminManagerRoleSet))
	i.RegisterRule(models.CompanyProfileModule, models.EditPermission, AdminManagerRoleSet, "/api/v1/space/profile/photo [post]", AllowByRoleFunc(AdminManagerRoleSet))
	i.RegisterRule(models.CompanyProfileModule, models.EditPermission, AdminManagerRoleSet, "/api/v1/space/profile/send_license_request [put]", AllowByRoleFunc(AdminManagerRoleSet))
}
