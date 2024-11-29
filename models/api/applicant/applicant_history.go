package applicantapimodels

import (
	apimodels "hr-tools-backend/models/api"
	dbmodels "hr-tools-backend/models/db"
)

type ApplicantHistoryFilter struct {
	apimodels.Pagination
	CommentsOnly bool `json:"comments_only"` // Только комментарии
}

type ApplicantHistoryView struct {
	VacancyID   string                    `json:"vacancy_id"`   // Идентификатор вакансии
	VacancyName string                    `json:"vacancy_name"` // Название вакансии
	UserID      string                    `json:"user_id"`      // Идентификатор сотрудника
	UserName    string                    `json:"user_name"`    // Имя сотрудника
	ActionType  dbmodels.ActionType       `json:"action_type"`  // Тип действия
	Changes     dbmodels.ApplicantChanges `json:"changes"`      // Изменения
}

func Convert(rec dbmodels.ApplicantHistory) ApplicantHistoryView {
	result := ApplicantHistoryView{
		VacancyID:   rec.VacancyID,
		VacancyName: "",
		UserID:      "",
		UserName:    rec.UserName,
		ActionType:  rec.ActionType,
		Changes:     rec.Changes,
	}
	if rec.Vacancy != nil {
		result.VacancyName = rec.Vacancy.VacancyName
	}
	if rec.UserID != nil {
		result.UserID = *rec.UserID
	}
	return result
}
