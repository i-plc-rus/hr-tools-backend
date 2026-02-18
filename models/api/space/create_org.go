package spaceapimodels

import (
	"errors"
	"regexp"
	"time"
)

type CreateOrganization struct {
	OrganizationName string           `json:"organization_name"`
	Inn              string           `json:"inn"`
	Kpp              string           `json:"kpp"`
	OGRN             string           `json:"ogrn"`
	FullName         string           `json:"full_name"`
	DirectorName     string           `json:"director_name"`
	AdminData        CreateSpaceAdmin `json:"admin_data"`
}

func (r CreateOrganization) Validate() error {
	if r.OrganizationName == "" {
		return errors.New("Укажите organization_name")
	}
	if r.Inn == "" {
		return errors.New("Укажите inn")
	}
	re := regexp.MustCompile(`^(([0-9]{12})|([0-9]{10}))?$`)
	if !re.MatchString(r.Inn) {
		return errors.New("Указан некорректный inn")
	}

	return r.AdminData.Validate()
}

type ProfileData struct {
	OrganizationName string `json:"organization_name"` // Название компании
	Web              string `json:"web"`               // Адрес сайта
	TimeZone         string `json:"time_zone"`         // Часовой пояс
	Description      string `json:"description"`       // Описание компании
	DirectorName     string `json:"director_name"`     // ФИО руководителя
	CompanyAddress   string `json:"company_address"`   //адрес организации
	CompanyContact   string `json:"company_contact"`   //контакт организации
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
