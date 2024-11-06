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
	Pol          string     `json:"pol"`
	Age          int        `json:"age"`
	Moving       string     `json:"moving"`
	Nationality  string     `json:"nationality"`
	Address      string     `json:"address"`
	LanguageList []Language `json:"language_list"`
}

type Language struct {
	Language      string `json:"language"`
	LanguageLevel string `json:"language_level"`
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

func (p ResumeParams) GetEngLevel() string {
	for _, lng := range p.LanguageList {
		if strings.ToUpper(lng.Language) == "АНГЛИЙСКИЙ" {
			return lng.LanguageLevel
		}
	}
	return ""
}
