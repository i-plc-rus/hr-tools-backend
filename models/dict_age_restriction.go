package models

import (
	"slices"

	"github.com/pkg/errors"
)

// Возрастное ограничение
type AgeRestriction string

const (
	AgeRestriction14Plus AgeRestriction = "AGE_14_PLUS"
	AgeRestriction16Plus AgeRestriction = "AGE_16_PLUS"
)

func AgeRestrictionSlice() []AgeRestriction {
	return []AgeRestriction{
		AgeRestriction14Plus,
		AgeRestriction16Plus,
	}
}

func (v AgeRestriction) Code() string {
	return string(v)
}

func (v AgeRestriction) Name() string {
	switch v {
	case AgeRestriction14Plus:
		return "От 14 лет"
	case AgeRestriction16Plus:
		return "От 16 лет"
	default:
		return ""
	}
}

func (v AgeRestriction) ToHh() string {
	return string(v)
}

func (v AgeRestriction) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("возрастное ограничение не указано")
	}
	if !slices.Contains(AgeRestrictionSlice(), v) {
		return errors.New("возрастное ограничение указано некорректно")
	}
	return nil
}
