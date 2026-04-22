package models

import (
	"slices"

	"github.com/pkg/errors"
)

type EmploymentForm string

const (
	EmploymentFormFull        EmploymentForm = "FULL"
	EmploymentFormPart        EmploymentForm = "PART"
	EmploymentFormProject     EmploymentForm = "PROJECT"
	EmploymentFormFlyInFlyOut EmploymentForm = "FLY_IN_FLY_OUT"
)

func EmploymentFormSlice() []EmploymentForm {
	return []EmploymentForm{
		EmploymentFormFull,
		EmploymentFormPart,
		EmploymentFormProject,
		EmploymentFormFlyInFlyOut,
	}
}

func EmploymentFormNameSlice() []string {
	return []string{"Полная занятость", "Частичная занятость", "Подработка", "Вахта"}
}

func (v EmploymentForm) Code() string {
	return string(v)
}

func (v EmploymentForm) Name() string {
	switch v {
	case EmploymentFormFull:
		return "Полная занятость"
	case EmploymentFormPart:
		return "Частичная занятость"
	case EmploymentFormProject:
		return "Подработка"
	case EmploymentFormFlyInFlyOut:
		return "Вахта"
	default:
		return ""
	}
}

func (v EmploymentForm) ToHh() string {
	return string(v)
}

func (v EmploymentForm) ToAvito() Employment {
	switch v {
	case EmploymentFormFull:
		return EmploymentFull
	case EmploymentFormPart:
		return EmploymentPartial
	case EmploymentFormProject:
		return EmploymentTemporary
	case EmploymentFormFlyInFlyOut:
		return EmploymentFull
	default:
		return ""
	}
}

func (v EmploymentForm) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("тип занятости не указан")
	}
	if !slices.Contains(EmploymentFormSlice(), v) {
		return errors.New("тип занятости указан некорректно")
	}
	return nil
}
