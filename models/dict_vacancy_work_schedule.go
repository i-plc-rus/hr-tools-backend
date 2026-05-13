package models

import (
	"slices"

	"github.com/pkg/errors"
)

// График работы
type WorkScheduleByDays string

const (
	WorkScheduleByDaysSixOnOneOff     WorkScheduleByDays = "6/1"
	WorkScheduleByDaysFiveOnTwoOff    WorkScheduleByDays = "5/2"
	WorkScheduleByDaysFourOnFourOff   WorkScheduleByDays = "4/4"
	WorkScheduleByDaysFourOnThreeOff  WorkScheduleByDays = "4/3"
	WorkScheduleByDaysFourOnTwoOff    WorkScheduleByDays = "4/2"
	WorkScheduleByDaysThreeOnThreeOff WorkScheduleByDays = "3/3"
	WorkScheduleByDaysThreeOnTwoOff   WorkScheduleByDays = "3/2"
	WorkScheduleByDaysTwoOnTwoOff     WorkScheduleByDays = "2/2"
	WorkScheduleByDaysTwoOnOneOff     WorkScheduleByDays = "2/1"
	WorkScheduleByDaysOneOnThreeOff   WorkScheduleByDays = "1/3"
	WorkScheduleByDaysOneOnTwoOff     WorkScheduleByDays = "1/2"
	WorkScheduleByDaysWeekend         WorkScheduleByDays = "По выходным"
	WorkScheduleByDaysFlexible        WorkScheduleByDays = "Свободный"
	WorkScheduleByDaysOther           WorkScheduleByDays = "Другое"
)

func WorkScheduleByDaysSlice() []WorkScheduleByDays {
	return []WorkScheduleByDays{
		WorkScheduleByDaysSixOnOneOff,
		WorkScheduleByDaysFiveOnTwoOff,
		WorkScheduleByDaysFourOnFourOff,
		WorkScheduleByDaysFourOnThreeOff,
		WorkScheduleByDaysFourOnTwoOff,
		WorkScheduleByDaysThreeOnThreeOff,
		WorkScheduleByDaysThreeOnTwoOff,
		WorkScheduleByDaysTwoOnTwoOff,
		WorkScheduleByDaysTwoOnOneOff,
		WorkScheduleByDaysOneOnThreeOff,
		WorkScheduleByDaysOneOnTwoOff,
		WorkScheduleByDaysWeekend,
		WorkScheduleByDaysFlexible,
		WorkScheduleByDaysOther,
	}
}

func (v WorkScheduleByDays) Code() string {
	return string(v)
}

func (v WorkScheduleByDays) Name() string {
	return string(v)
}

func (s WorkScheduleByDays) ToHh() string {
	switch s {
	case WorkScheduleByDaysSixOnOneOff:
		return "SIX_ON_ONE_OFF"
	case WorkScheduleByDaysFiveOnTwoOff:
		return "FIVE_ON_TWO_OFF"
	case WorkScheduleByDaysFourOnFourOff:
		return "FOUR_ON_FOUR_OFF"
	case WorkScheduleByDaysFourOnThreeOff:
		return "FOUR_ON_THREE_OFF"
	case WorkScheduleByDaysFourOnTwoOff:
		return "FOUR_ON_TWO_OFF"
	case WorkScheduleByDaysThreeOnThreeOff:
		return "THREE_ON_THREE_OFF"
	case WorkScheduleByDaysThreeOnTwoOff:
		return "THREE_ON_TWO_OFF"
	case WorkScheduleByDaysTwoOnTwoOff:
		return "TWO_ON_TWO_OFF"
	case WorkScheduleByDaysTwoOnOneOff:
		return "TWO_ON_ONE_OFF"
	case WorkScheduleByDaysOneOnThreeOff:
		return "ONE_ON_THREE_OFF"
	case WorkScheduleByDaysOneOnTwoOff:
		return "ONE_ON_TWO_OFF"
	case WorkScheduleByDaysWeekend:
		return "WEEKEND"
	case WorkScheduleByDaysFlexible:
		return "FLEXIBLE"
	case WorkScheduleByDaysOther:
		return "OTHER"
	default:
		return ""
	}
}

func (v WorkScheduleByDays) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("график работы не указан")
	}
	if !slices.Contains(WorkScheduleByDaysSlice(), v) {
		return errors.New("график работы указан некорректно")
	}
	return nil
}
