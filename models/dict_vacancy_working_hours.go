package models

import (
	"slices"

	"github.com/pkg/errors"
)

// Рабочие часы в день

type WorkingHours string

const (
	WorkingHours8 WorkingHours = "8 часов"
)

func WorkingHoursSlice() []WorkingHours {
	return []WorkingHours{WorkingHours8}
}

func (v WorkingHours) Code() string {
	return string(v)
}

func (v WorkingHours) Name() string {
	return string(v)
}

func (v WorkingHours) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("параметр рабочие часы не указан")
	}
	if !slices.Contains(WorkingHoursSlice(), v) {
		return errors.New("параметр рабочие часы указан некорректно")
	}
	return nil
}

func (s WorkingHours) ToHhCode() string {
	switch s {
	case WorkingHours8:
		return "HOURS_8"
	default:
		return ""
	}
}
