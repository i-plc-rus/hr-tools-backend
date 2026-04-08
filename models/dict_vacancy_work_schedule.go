package models

import (
	"slices"

	"github.com/pkg/errors"
)

// График работы
type WorkSchedule string

const (
	WorkScheduleSixOnOneOff     WorkSchedule = "6/1"
	WorkScheduleFiveOnTwoOff    WorkSchedule = "5/2"
	WorkScheduleFourOnFourOff   WorkSchedule = "4/4"
	WorkScheduleFourOnThreeOff  WorkSchedule = "4/3"
	WorkScheduleFourOnTwoOff    WorkSchedule = "4/2"
	WorkScheduleThreeOnThreeOff WorkSchedule = "3/3"
	WorkScheduleThreeOnTwoOff   WorkSchedule = "3/2"
	WorkScheduleTwoOnTwoOff     WorkSchedule = "2/2"
	WorkScheduleTwoOnOneOff     WorkSchedule = "2/1"
	WorkScheduleOneOnThreeOff   WorkSchedule = "1/3"
	WorkScheduleOneOnTwoOff     WorkSchedule = "1/2"
	WorkScheduleWeekend         WorkSchedule = "По выходным"
	WorkScheduleFlexible        WorkSchedule = "Свободный"
	WorkScheduleOther           WorkSchedule = "Другое"
)

func WorkScheduleSlice() []WorkSchedule {
	return []WorkSchedule{
		WorkScheduleSixOnOneOff,
		WorkScheduleFiveOnTwoOff,
		WorkScheduleFourOnFourOff,
		WorkScheduleFourOnThreeOff,
		WorkScheduleFourOnTwoOff,
		WorkScheduleThreeOnThreeOff,
		WorkScheduleThreeOnTwoOff,
		WorkScheduleTwoOnTwoOff,
		WorkScheduleTwoOnOneOff,
		WorkScheduleOneOnThreeOff,
		WorkScheduleOneOnTwoOff,
		WorkScheduleWeekend,
		WorkScheduleFlexible,
		WorkScheduleOther,
	}
}

func (v WorkSchedule) Code() string {
	return string(v)
}

func (v WorkSchedule) Name() string {
	return string(v)
}

func (s WorkSchedule) ToHhCode() string {
	switch s {
	case WorkScheduleSixOnOneOff:
		return "SIX_ON_ONE_OFF"
	case WorkScheduleFiveOnTwoOff:
		return "FIVE_ON_TWO_OFF"
	case WorkScheduleFourOnFourOff:
		return "FOUR_ON_FOUR_OFF"
	case WorkScheduleFourOnThreeOff:
		return "FOUR_ON_THREE_OFF"
	case WorkScheduleFourOnTwoOff:
		return "FOUR_ON_TWO_OFF"
	case WorkScheduleThreeOnThreeOff:
		return "THREE_ON_THREE_OFF"
	case WorkScheduleThreeOnTwoOff:
		return "THREE_ON_TWO_OFF"
	case WorkScheduleTwoOnTwoOff:
		return "TWO_ON_TWO_OFF"
	case WorkScheduleTwoOnOneOff:
		return "TWO_ON_ONE_OFF"
	case WorkScheduleOneOnThreeOff:
		return "ONE_ON_THREE_OFF"
	case WorkScheduleOneOnTwoOff:
		return "ONE_ON_TWO_OFF"
	case WorkScheduleWeekend:
		return "WEEKEND"
	case WorkScheduleFlexible:
		return "FLEXIBLE"
	case WorkScheduleOther:
		return "OTHER"
	default:
		return ""
	}
}

func (v WorkSchedule) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("график работы не указан")
	}
	if !slices.Contains(WorkScheduleSlice(), v) {
		return errors.New("график работы указан некорректно")
	}
	return nil
}
