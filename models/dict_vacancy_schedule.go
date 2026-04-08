package models

import (
	"slices"

	"github.com/pkg/errors"
)

// График работы
// Deprecated: используются WorkSchedule
type Schedule string

const (
	ScheduleFlyInFlyOut Schedule = "flyInFlyOut" // Вахта
	SchedulePartTime    Schedule = "partTime"    // Неполный день
	ScheduleFullDay     Schedule = "fullDay"     // Полный день
	ScheduleFlexible    Schedule = "flexible"    // Гибкий
	ScheduleShift       Schedule = "shift"       // Сменный
)

func ScheduleNameSlice() []string {
	return []string{"Вахта", "Неполный день", "Полный день", "Гибкий", "Сменный"}
}

func ScheduleSlice() []Schedule {
	return []Schedule{
		ScheduleFlyInFlyOut,
		SchedulePartTime,
		ScheduleFullDay,
		ScheduleFlexible,
		ScheduleShift,
	}
}

func (v Schedule) Code() string {
	return string(v)
}

func (v Schedule) Name() string {
	return v.ToString()
}

func (s Schedule) ToString() string {
	switch s {
	case ScheduleFlyInFlyOut:
		return "Вахта"
	case SchedulePartTime:
		return "Неполный день"
	case ScheduleFullDay:
		return "Полный день"
	case ScheduleFlexible:
		return "Гибкий"
	case ScheduleShift:
		return "Сменный"
	}
	return ""
}

func (v Schedule) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("занятость не указана")
	}
	if !slices.Contains(ScheduleSlice(), v) {
		return errors.New("занятость указана некорректно")
	}
	return nil
}