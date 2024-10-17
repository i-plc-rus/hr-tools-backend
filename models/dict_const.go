package models

import "github.com/pkg/errors"

type VRUrgency string

const (
	VRTypeUrgent    VRUrgency = "Срочно"
	VRTypeNonUrgent VRUrgency = "В плановом порядке"
)

func (v VRUrgency) Validate() error {
	if v == "" {
		return errors.New("параметр срочности не указан")
	}
	if v != VRTypeUrgent && v != VRTypeNonUrgent {
		return errors.New("параметр срочности указан некорректно")
	}
	return nil
}

type VRType string

const (
	VRTypeNew     VRType = "Новая позиция"
	VRTypeReplace VRType = "Замена"
)

func (v VRType) Validate() error {
	if v == "" {
		return errors.New("тип вакансии не указан")
	}
	if v != VRTypeNew && v != VRTypeReplace {
		return errors.New("типа вакансии указан некорректно")
	}
	return nil
}

type VRSelectionType string

const (
	VRSelectionTypeMass     VRSelectionType = "Массовый"
	VRSelectionTypePersonal VRSelectionType = "Индивидуальный"
)

func (v VRSelectionType) Validate() error {
	if v == "" {
		return errors.New("вид подбора не указан")
	}
	if v != VRSelectionTypeMass && v != VRSelectionTypePersonal {
		return errors.New("вид подбора указан некорректно")
	}
	return nil
}

type VacancyStatus string

const (
	VacancyStatusOpened    VacancyStatus = "Открыта"
	VacancyStatusCanceled  VacancyStatus = "Отменена"
	VacancyStatusSuspended VacancyStatus = "Приостановлена"
	VacancyStatusClosed    VacancyStatus = "Закрыта"
)
