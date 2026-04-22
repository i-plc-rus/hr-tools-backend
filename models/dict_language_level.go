package models

import (
	"slices"

	"github.com/pkg/errors"
)

type LanguageLevelType string

const (
	LanguageLevelA1 LanguageLevelType = "a1"
	LanguageLevelA2 LanguageLevelType = "a2"
	LanguageLevelB1 LanguageLevelType = "b1"
	LanguageLevelB2 LanguageLevelType = "b2"
	LanguageLevelC1 LanguageLevelType = "c1"
	LanguageLevelC2 LanguageLevelType = "c2"
	LanguageLevelL1 LanguageLevelType = "l1"
)

func LanguageLevelSlice() []LanguageLevelType {
	return []LanguageLevelType{LanguageLevelA1, LanguageLevelA2, LanguageLevelB1, LanguageLevelB2, LanguageLevelC1, LanguageLevelC2, LanguageLevelL1}
}

func (v LanguageLevelType) Code() string {
	return string(v)
}

func (v LanguageLevelType) Name() string {
	return v.ToString()
}

func (e LanguageLevelType) ToString() string {
	switch e {
	case LanguageLevelA1:
		return "A1 — Начальный"
	case LanguageLevelA2:
		return "A2 — Элементарный"
	case LanguageLevelB1:
		return "B1 — Средний"
	case LanguageLevelB2:
		return "B2 — Средне-продвинутый"
	case LanguageLevelC1:
		return "C1 — Продвинутый"
	case LanguageLevelC2:
		return "C2 — В совершенстве"
	case LanguageLevelL1:
		return "Родной"
	}
	return ""
}

func (e LanguageLevelType) ToHH() string {
	return e.Code()
}

func (v LanguageLevelType) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("уровень владения языком не указан")
	}
	if !slices.Contains(LanguageLevelSlice(), v) {
		return errors.New("уровень владения языком указан некорректно")
	}
	return nil
}
