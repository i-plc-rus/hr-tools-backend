package models

import (
	"slices"

	"github.com/pkg/errors"
)

// Рабочие часы в день
type WorkingHours string

const (
	WorkingHours2        WorkingHours = "HOURS_2"
	WorkingHours3        WorkingHours = "HOURS_3"
	WorkingHours4        WorkingHours = "HOURS_4"
	WorkingHours5        WorkingHours = "HOURS_5"
	WorkingHours6        WorkingHours = "HOURS_6"
	WorkingHours7        WorkingHours = "HOURS_7"
	WorkingHours8        WorkingHours = "HOURS_8"
	WorkingHours9        WorkingHours = "HOURS_9"
	WorkingHours10       WorkingHours = "HOURS_10"
	WorkingHours11       WorkingHours = "HOURS_11"
	WorkingHours12       WorkingHours = "HOURS_12"
	WorkingHours24       WorkingHours = "HOURS_24"
	WorkingHoursFlexible WorkingHours = "FLEXIBLE"
	WorkingHoursOther    WorkingHours = "OTHER"
)

// WorkingHoursSlice возвращает список всех возможных значений.
func WorkingHoursSlice() []WorkingHours {
	return []WorkingHours{
		WorkingHours2,
		WorkingHours3,
		WorkingHours4,
		WorkingHours5,
		WorkingHours6,
		WorkingHours7,
		WorkingHours8,
		WorkingHours9,
		WorkingHours10,
		WorkingHours11,
		WorkingHours12,
		WorkingHours24,
		WorkingHoursFlexible,
		WorkingHoursOther,
	}
}

// Code возвращает идентификатор значения (то же самое, что строковое представление).
func (v WorkingHours) Code() string {
	return string(v)
}

// Name возвращает человекочитаемое название на русском языке.
func (v WorkingHours) Name() string {
	switch v {
	case WorkingHours2:
		return "2 часа"
	case WorkingHours3:
		return "3 часа"
	case WorkingHours4:
		return "4 часа"
	case WorkingHours5:
		return "5 часов"
	case WorkingHours6:
		return "6 часов"
	case WorkingHours7:
		return "7 часов"
	case WorkingHours8:
		return "8 часов"
	case WorkingHours9:
		return "9 часов"
	case WorkingHours10:
		return "10 часов"
	case WorkingHours11:
		return "11 часов"
	case WorkingHours12:
		return "12 часов"
	case WorkingHours24:
		return "24 часа"
	case WorkingHoursFlexible:
		return "По договорённости"
	case WorkingHoursOther:
		return "Другое"
	default:
		return ""
	}
}

func (v WorkingHours) ToHh() string {
	return string(v)
}

func (v WorkingHours) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("рабочие часы в день не указаны")
	}
	if !slices.Contains(WorkingHoursSlice(), v) {
		return errors.New("рабочие часы в день указаны некорректно")
	}
	return nil
}