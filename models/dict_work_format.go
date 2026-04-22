package models

import (
	"slices"

	"github.com/pkg/errors"
)

// Формат работы
type WorkFormat string

const (
	WorkFormatOnSite    WorkFormat = "ON_SITE"
	WorkFormatRemote    WorkFormat = "REMOTE"
	WorkFormatHybrid    WorkFormat = "HYBRID"
	WorkFormatFieldWork WorkFormat = "FIELD_WORK"
)

func WorkFormatSlice() []WorkFormat {
	return []WorkFormat{
		WorkFormatOnSite,
		WorkFormatRemote,
		WorkFormatHybrid,
		WorkFormatFieldWork,
	}
}

func WorkFormatNameSlice() []string {
	return []string{"На месте работодателя", "Удалённо", "Гибрид", "Разъездной"}
}

func (v WorkFormat) Code() string {
	return string(v)
}

func (v WorkFormat) Name() string {
	switch v {
	case WorkFormatOnSite:
		return "На месте работодателя"
	case WorkFormatRemote:
		return "Удалённо"
	case WorkFormatHybrid:
		return "Гибрид"
	case WorkFormatFieldWork:
		return "Разъездной"
	default:
		return ""
	}
}

func (s WorkFormat) ToHh() string {
	return string(s)
}

func (v WorkFormat) ToAvito() Schedule {
	switch v {
	case WorkFormatHybrid:
		return ScheduleFlexible
	default:
		return ""
	}
}

func (v WorkFormat) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("формат работы не указан")
	}
	if !slices.Contains(WorkFormatSlice(), v) {
		return errors.New("формат работы указан некорректно")
	}
	return nil
}
