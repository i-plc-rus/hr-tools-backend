package avitoapimodels

import (
	"hr-tools-backend/models"
	"strings"
)

type Resume struct {
	ID     int          `json:"id"`
	Title  string       `json:"title"`
	Salary int          `json:"salary"`
	Params ResumeParams `json:"params"`
}

type ResumeParams struct {
	Pol                   string     `json:"pol"`
	Age                   int        `json:"age"`
	Moving                string     `json:"moving"`
	Nationality           string     `json:"nationality"`
	Address               string     `json:"address"`
	LanguageList          []Language `json:"language_list"`
	Education             string     `json:"education"`
	AbilityToBusinessTrip string     `json:"ability_to_business_trip"`
	DriverLicence         []string   `json:"driver_licence_category"`
}

type Language struct {
	Language      string `json:"language"`
	LanguageLevel string `json:"language_level"`
}

func (l Language) GetLanguageLevelType() models.LanguageLevelType {
	switch l.LanguageLevel {
	case "Начальный":
		return models.LanguageLevelA1
	case "Средний":
		return models.LanguageLevelB1
	case "Выше среднего":
		return models.LanguageLevelB2
	case "Свободное владение":
		return models.LanguageLevelC2
	}
	return ""
}

func (p ResumeParams) GetTripReadinessType() models.TripReadinessType {
	switch p.AbilityToBusinessTrip {
	case "Готов":
		return models.TripReadinessReady
	case "Иногда":
		return models.TripReadinessSometimes
	case "Не готов":
		return models.TripReadinessNever
	}
	return ""
}

func (p ResumeParams) GetRelocationType() models.RelocationType {
	switch strings.ToUpper(p.Moving) {
	case "НЕВОЗМОЖЕН":
		return models.RelocationTypeNo
	case "ВОЗМОЖЕН":
		return models.RelocationTypeYes
	}
	return ""
}

func (p ResumeParams) GetGender() models.GenderType {
	if p.Pol == "" {
		return ""
	}
	if p.Pol == "Мужской" {
		return models.GenderTypeM
	}
	return models.GenderTypeF
}

func (p ResumeParams) GetEducationType() models.EducationType {
	switch strings.ToUpper(p.Education) {
	case "Высшее":
		return models.EducationTypeHigher
	case "Незаконченное высшее":
		return models.EducationTypeUnfinishedHigher
	case "Среднее":
		return models.EducationTypeSecondary
	case "Среднее специальное":
		return models.EducationTypeSpecialSecondary
	}
	return ""
}

func (p ResumeParams) GetDriverLicence() []models.DriverLicenseType {
	if len(p.DriverLicence) == 0 {
		return []models.DriverLicenseType{}
	}
	result := make([]models.DriverLicenseType, 0, len(p.DriverLicence))
	for _, li := range p.DriverLicence {
		var licence models.DriverLicenseType
		switch li {
		case "a":
			licence = models.DriverLicenseA
		case "b":
			licence = models.DriverLicenseB
		case "be":
			licence = models.DriverLicenseBE
		case "c":
			licence = models.DriverLicenseC
		case "ce":
			licence = models.DriverLicenseCE
		case "d":
			licence = models.DriverLicenseD
		case "de":
			licence = models.DriverLicenseDE
		case "tm":
			licence = models.DriverLicenseTM
		case "tb":
			licence = models.DriverLicenseTB
		}
		if licence != "" {
			result = append(result, licence)
		}
	}
	return result
}
