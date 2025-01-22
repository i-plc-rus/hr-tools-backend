package spaceapimodels

import (
	"errors"
	"time"
)

type CreateOrganization struct {
	OrganizationName string     `json:"organization_name"`
	Inn              string     `json:"inn"`
	Kpp              string     `json:"kpp"`
	OGRN             string     `json:"ogrn"`
	FullName         string     `json:"full_name"`
	DirectorName     string     `json:"director_name"`
	AdminData        CreateUser `json:"admin_data"`
}

func (r CreateOrganization) Validate() error {
	//TODO заглушка
	return nil
}

type ProfileData struct {
	OrganizationName string `json:"organization_name"` // Название компании
	Web              string `json:"web"`               // Адрес сайта
	TimeZone         string `json:"time_zone"`         // Часовой пояс
	Description      string `json:"description"`       // Описание компании
	DirectorName     string `json:"director_name"`     // ФИО руководителя
}

func (p ProfileData) Validate() error {
	if p.TimeZone != "" {
		_, err := time.LoadLocation(p.TimeZone)
		if err != nil {
			return errors.New("Часовой пояс не найден")
		}
	}
	if p.OrganizationName == "" {
		return errors.New("Не указано название компании")
	}
	return nil
}
