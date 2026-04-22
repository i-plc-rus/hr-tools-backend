package models

type IDict interface {
	Code() string
	Name() string
	Validate(optional bool) error
}

type CommonDictItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CommonDict struct {
	VRUrgency           []CommonDictItem `json:"vacancy_urgency"`
	VRType              []CommonDictItem `json:"vacancy_request_type"`
	VRSelectionType     []CommonDictItem `json:"vacancy_selection_type"`
	EmploymentForm      []CommonDictItem `json:"vacancy_employment_from"`
	Experience          []CommonDictItem `json:"vacancy_experience"`
	Schedule            []CommonDictItem `json:"vacancy_schedule"`
	WorkSchedule        []CommonDictItem `json:"vacancy_work_schedule"`
	WorkingHours        []CommonDictItem `json:"vacancy_working_hours"`
	LanguageLevelTypes  []CommonDictItem `json:"language_level_types"`
	DriverLicenseTypes  []CommonDictItem `json:"driver_license_types"`
	WorkFormat          []CommonDictItem `json:"work_format"`
	WorkHours           []CommonDictItem `json:"work_hours"`
	FlyInFlyOutDuration []CommonDictItem `json:"fly_in_fly_out_duration"`
	CivilLawContract    []CommonDictItem `json:"civil_law_contract"`
	AgeRestriction      []CommonDictItem `json:"age_restriction"`
}

func GetCommonDicts() CommonDict {
	result := CommonDict{
		VRUrgency:           toCommonDictItems(VRUrgencySlice()),
		VRType:              toCommonDictItems(VRTypeSlice()),
		VRSelectionType:     toCommonDictItems(VRSelectionTypeSlice()),
		EmploymentForm:      toCommonDictItems(EmploymentFormSlice()),
		Experience:          toCommonDictItems(ExperienceSlice()),
		Schedule:            toCommonDictItems(ScheduleSlice()),
		WorkSchedule:        toCommonDictItems(WorkScheduleByDaysSlice()),
		LanguageLevelTypes:  toCommonDictItems(LanguageLevelSlice()),
		DriverLicenseTypes:  toCommonDictItems(DriverLicensesSlice()),
		WorkFormat:          toCommonDictItems(WorkFormatSlice()),
		WorkingHours:        toCommonDictItems(WorkingHoursSlice()),
		FlyInFlyOutDuration: toCommonDictItems(FlyInFlyOutDurationSlice()),
		CivilLawContract:    toCommonDictItems(CivilLawContractSlice()),
		AgeRestriction:      toCommonDictItems(AgeRestrictionSlice()),
	}
	return result
}

func toCommonDictItems[T IDict](items []T) []CommonDictItem {
	result := make([]CommonDictItem, len(items))
	for i, item := range items {
		result[i] = CommonDictItem{
			ID:   item.Code(),
			Name: item.Name(),
		}
	}
	return result
}
