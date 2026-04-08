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
	VRUrgency       []CommonDictItem `json:"vacancy_urgency"`
	VRType          []CommonDictItem `json:"vacancy_request_type"`
	VRSelectionType []CommonDictItem `json:"vacancy_selection_type"`
	Employment      []CommonDictItem `json:"vacancy_employment"`
	Experience      []CommonDictItem `json:"vacancy_experience"`
	Schedule        []CommonDictItem `json:"vacancy_schedule"`
	WorkSchedule    []CommonDictItem `json:"vacancy_work_schedule"`
	WorkingHours    []CommonDictItem `json:"vacancy_working_hours"`
}

func GetCommonDicts() CommonDict {
	result := CommonDict{
		VRUrgency:       toCommonDictItems(VRUrgencySlice()),
		VRType:          toCommonDictItems(VRTypeSlice()),
		VRSelectionType: toCommonDictItems(VRSelectionTypeSlice()),
		Employment:      toCommonDictItems(EmploymentSlice()),
		Experience:      toCommonDictItems(ExperienceSlice()),
		Schedule:        toCommonDictItems(ScheduleSlice()),
		WorkSchedule:    toCommonDictItems(WorkScheduleSlice()),
		WorkingHours:    toCommonDictItems(WorkingHoursSlice()),
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