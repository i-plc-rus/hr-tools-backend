package models

import (
	"slices"

	"github.com/pkg/errors"
)

type VRSelectionType string

const (
	VRSelectionTypeMass     VRSelectionType = "Массовый"
	VRSelectionTypePersonal VRSelectionType = "Индивидуальный"
)

func VRSelectionTypeSlice() []VRSelectionType {
	return []VRSelectionType{VRSelectionTypeMass, VRSelectionTypePersonal}
}

func (v VRSelectionType) Code() string {
	return string(v)
}

func (v VRSelectionType) Name() string {
	return string(v)
}

func (v VRSelectionType) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("вид подбора не указан")
	}
	if !slices.Contains(VRSelectionTypeSlice(), v) {
		return errors.New("вид подбора указан некорректно")
	}
	return nil
}
