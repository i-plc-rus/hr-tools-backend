package models

import (
	"slices"

	"github.com/pkg/errors"
)

type Employment string

const (
	EmploymentTemporary  Employment = "temporary"  //Временная
	EmploymentFull       Employment = "full"       //Полная
	EmploymentInternship Employment = "internship" //Стажировка
	EmploymentPartial    Employment = "partial"    //Частичная
	EmploymentVolunteer  Employment = "volunteer"  //Волонтерство
	EmploymentProbation  Employment = "probation"  //Стажировка
)

func EmploymentNamesSlice() []string {
	return []string{"Временная", "Полная", "Стажировка", "Частичная", "Волонтерство"}
}

func EmploymentSlice() []Employment {
	return []Employment{EmploymentTemporary, EmploymentFull, EmploymentInternship, EmploymentPartial, EmploymentVolunteer, EmploymentProbation}
}

func (v Employment) Code() string {
	return string(v)
}

func (v Employment) Name() string {
	return v.ToString()
}

func (e Employment) ToString() string {
	switch e {
	case EmploymentTemporary:
		return "Временная"
	case EmploymentFull:
		return "Полная"
	case EmploymentInternship:
		return "Стажировка"
	case EmploymentPartial:
		return "Частичная"
	case EmploymentVolunteer:
		return "Волонтерство"
	case EmploymentProbation:
		return "Стажировка"
	}
	return ""
}

func (e Employment) ToHHEmploymentForm() string {
	switch e {
	case EmploymentTemporary:
		return "PROJECT"
	case EmploymentFull:
		return "FULL"
	case EmploymentPartial:
		return "PART"
	}
	return ""
}

func (v Employment) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("занятость не указана")
	}
	if !slices.Contains(EmploymentSlice(), v) {
		return errors.New("занятость указана некорректно")
	}
	return nil
}
