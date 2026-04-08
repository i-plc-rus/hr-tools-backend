package models

import (
	"slices"

	"github.com/pkg/errors"
)

type VRUrgency string

const (
	VRTypeUrgent    VRUrgency = "Срочно"
	VRTypeNonUrgent VRUrgency = "В плановом порядке"
)

func VRUrgencySlice() []VRUrgency {
	return []VRUrgency{VRTypeUrgent, VRTypeNonUrgent}
}

func (v VRUrgency) Code() string {
	return string(v)
}

func (v VRUrgency) Name() string {
	return string(v)
}

func (v VRUrgency) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("параметр срочности не указан")
	}
	if !slices.Contains(VRUrgencySlice(), v) {
		return errors.New("параметр срочности указан некорректно")
	}
	return nil
}
