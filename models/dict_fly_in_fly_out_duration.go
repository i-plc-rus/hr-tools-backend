package models

import (
	"slices"

	"github.com/pkg/errors"
)

// Длительность вахты
type FlyInFlyOutDuration string

const (
	FlyInFlyOutDurationDays15  FlyInFlyOutDuration = "DAYS_15"
	FlyInFlyOutDurationDays20  FlyInFlyOutDuration = "DAYS_20"
	FlyInFlyOutDurationDays30  FlyInFlyOutDuration = "DAYS_30"
	FlyInFlyOutDurationDays40  FlyInFlyOutDuration = "DAYS_40"
	FlyInFlyOutDurationDays45  FlyInFlyOutDuration = "DAYS_45"
	FlyInFlyOutDurationDays60  FlyInFlyOutDuration = "DAYS_60"
	FlyInFlyOutDurationDays90  FlyInFlyOutDuration = "DAYS_90"
	FlyInFlyOutDurationDays120 FlyInFlyOutDuration = "DAYS_120"
	FlyInFlyOutDurationDays180 FlyInFlyOutDuration = "DAYS_180"
	FlyInFlyOutDurationOther   FlyInFlyOutDuration = "OTHER"
)

func FlyInFlyOutDurationSlice() []FlyInFlyOutDuration {
	return []FlyInFlyOutDuration{
		FlyInFlyOutDurationDays15,
		FlyInFlyOutDurationDays20,
		FlyInFlyOutDurationDays30,
		FlyInFlyOutDurationDays40,
		FlyInFlyOutDurationDays45,
		FlyInFlyOutDurationDays60,
		FlyInFlyOutDurationDays90,
		FlyInFlyOutDurationDays120,
		FlyInFlyOutDurationDays180,
		FlyInFlyOutDurationOther,
	}
}

func (v FlyInFlyOutDuration) Code() string {
	return string(v)
}

func (v FlyInFlyOutDuration) Name() string {
	switch v {
	case FlyInFlyOutDurationDays15:
		return "15"
	case FlyInFlyOutDurationDays20:
		return "20"
	case FlyInFlyOutDurationDays30:
		return "30"
	case FlyInFlyOutDurationDays40:
		return "40"
	case FlyInFlyOutDurationDays45:
		return "45"
	case FlyInFlyOutDurationDays60:
		return "60"
	case FlyInFlyOutDurationDays90:
		return "90"
	case FlyInFlyOutDurationDays120:
		return "120"
	case FlyInFlyOutDurationDays180:
		return "180"
	case FlyInFlyOutDurationOther:
		return "Другое"
	default:
		return ""
	}
}


func (v FlyInFlyOutDuration) ToHh() string {
	return string(v)
}

func (v FlyInFlyOutDuration) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("длительность вахты не указана")
	}
	if !slices.Contains(FlyInFlyOutDurationSlice(), v) {
		return errors.New("длительность вахты указана некорректно")
	}
	return nil
}
