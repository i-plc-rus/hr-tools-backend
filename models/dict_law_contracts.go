package models

import (
	"slices"

	"github.com/pkg/errors"
)

// Договор гражданско-правового характера
type CivilLawContract string

const (
	CivilLawContractSelfEmployed           CivilLawContract = "SELF_EMPLOYED"
	CivilLawContractIndividualEntrepreneur CivilLawContract = "INDIVIDUAL_ENTREPRENEUR"
	CivilLawContractIndividualPerson       CivilLawContract = "INDIVIDUAL_PERSON"
)

func CivilLawContractSlice() []CivilLawContract {
	return []CivilLawContract{
		CivilLawContractSelfEmployed,
		CivilLawContractIndividualEntrepreneur,
		CivilLawContractIndividualPerson,
	}
}

func (v CivilLawContract) Code() string {
	return string(v)
}

func (v CivilLawContract) Name() string {
	switch v {
	case CivilLawContractSelfEmployed:
		return "с самозанятым"
	case CivilLawContractIndividualEntrepreneur:
		return "с ИП"
	case CivilLawContractIndividualPerson:
		return "с физлицом"
	default:
		return ""
	}
}

func (v CivilLawContract) ToHh() string {
	return string(v)
}

func (v CivilLawContract) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("договор гражданско-правового характера не указан")
	}
	if !slices.Contains(CivilLawContractSlice(), v) {
		return errors.New("договор гражданско-правового характера указан некорректно")
	}
	return nil
}
