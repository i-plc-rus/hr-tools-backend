package vacancyapimodels

import (
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
)

type Person struct {
	ID       string          `json:"id"`
	FullName string          `json:"full_name"`
	Role     models.UserRole `json:"role"`
	Email    string          `json:"email"` // Email пользователя
}

type TeamPerson struct {
	Person
	Responsible bool `json:"responsible"`
}

func TeamPersonConvert(rec dbmodels.VacancyTeam) TeamPerson {
	result := TeamPerson{
		Person: Person{
			ID:       rec.ID,
		},
		Responsible: rec.Responsible,
	}
	if rec.SpaceUser != nil {
		result.FullName = rec.SpaceUser.GetFullName()
		result.Role = rec.SpaceUser.Role
		result.Email = rec.SpaceUser.Email
	}
	return result
}

func PersonConvert(rec dbmodels.SpaceUser) Person {
	return Person{
		ID:       rec.ID,
		FullName: rec.GetFullName(),
		Role:     rec.Role,
		Email:    rec.Email,
	}
}

type PersonFilter struct {
	Search           string                 `json:"search"`            // Поиск по ФИО
}