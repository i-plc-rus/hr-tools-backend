package models

import (
	"slices"

	"github.com/pkg/errors"
)

type DriverLicenseType string

const (
	DriverLicenseA  DriverLicenseType = "A"
	DriverLicenseB  DriverLicenseType = "B"
	DriverLicenseC  DriverLicenseType = "C"
	DriverLicenseD  DriverLicenseType = "D"
	DriverLicenseE  DriverLicenseType = "E"
	DriverLicenseBE DriverLicenseType = "BE"
	DriverLicenseCE DriverLicenseType = "CE"
	DriverLicenseDE DriverLicenseType = "DE"
	DriverLicenseTM DriverLicenseType = "TM"
	DriverLicenseTB DriverLicenseType = "TB"
)

func DriverLicensesSlice() []DriverLicenseType {
	return []DriverLicenseType{DriverLicenseA, DriverLicenseB, DriverLicenseC, DriverLicenseD, DriverLicenseE, DriverLicenseBE,
		DriverLicenseCE, DriverLicenseDE, DriverLicenseTM, DriverLicenseTB,
	}
}

func (v DriverLicenseType) Code() string {
	return string(v)
}

func (v DriverLicenseType) Name() string {
	return string(v)
}

func (v DriverLicenseType) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("параметр категория водительских прав не указан")
	}
	if !slices.Contains(DriverLicensesSlice(), v) {
		return errors.New("параметр категория водительских прав указан некорректно")
	}
	return nil
}

func (s DriverLicenseType) ToHh() string {
	return s.Code()
}
