package models

import (
	"slices"

	"github.com/pkg/errors"
)

type VRType string

const (
	VRTypeNew     VRType = "Новая позиция"
	VRTypeReplace VRType = "Замена"
)

func VRTypeSlice() []VRType {
	return []VRType{VRTypeNew, VRTypeReplace}
}

func (v VRType) Code() string {
	return string(v)
}

func (v VRType) Name() string {
	return string(v)
}

func (v VRType) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("тип вакансии не указан")
	}
	if !slices.Contains(VRTypeSlice(), v) {
		return errors.New("типа вакансии указан некорректно")
	}
	return nil
}
