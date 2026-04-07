package hhapimodels

type VacancyDraftRequest struct {
	ClosedForApplicants     bool              `json:"closed_for_applicants"` //false
	VacancyProperties       []VacancyProperty `json:"vacancy_properties"`
	AcceptHandicapped       bool              `json:"accept_handicapped"`        //true
	AcceptIncompleteResumes bool              `json:"accept_incomplete_resumes"` //false
	AcceptLaborContract     bool              `json:"accept_labor_contract"`     //true
	AcceptTemporary         bool              `json:"accept_temporary"`          //true
	Address                 *AddressDraft     `json:"address"`
	AgeRestriction          *DictItem         `json:"age_restriction"` //AGE_14_PLUS
	AllowMessages           bool              `json:"allow_messages"`  // Разрешение сообщений true
	Areas                   []DictItem        `json:"areas"`
	AutoResponse            AutoResponse      `json:"auto_response"`
	BrandedTemplate         *DictItem         `json:"branded_template"` //"marketing"
	CivilLawContracts       *[]DictItem       `json:"civil_law_contracts"`
	Code                    *string           `json:"code"` //"код-1234
	Contacts                *Contacts         `json:"contacts,omitempty"`
	Department              *DictItem         `json:"department"`
	Description             string            `json:"description"`
	DriverLicenseTypes      []DictItem        `json:"driver_license_types"`
	Employment              *DictItem         `json:"employment"`
	EmploymentFrom          *DictItem         `json:"employment_form,omitempty"` //Тип занятости
	Experience              *DictItem         `json:"experience,omitempty"`      // Опыт работы
	Internship              bool              `json:"internship"`                //false
	KeySkills               *[]KeySkill       `json:"key_skills"`
	Languages               *[]Language       `json:"languages"`
	Name                    string            `json:"name"`         // Менеджер по продажам
	NightShifts             bool              `json:"night_shifts"` //true
	ProfessionalRoles       *[]DictNameItem   `json:"professional_roles"`
	ResponseLetterRequired  bool              `json:"response_letter_required"` //true
	ResponseNotifications   bool              `json:"response_notifications"`   //true
	SalaryRange             *SalaryRange      `json:"salary_range,omitempty"`
	Schedule                *DictItem         `json:"schedule,omitempty"` // График работы
	Test                    *Test             `json:"test"`
	WithZp                  bool              `json:"with_zp"`
	WorkFormat              []DictNameItem    `json:"work_format"`
	WorkScheduleByDays      []DictNameItem    `json:"work_schedule_by_days"`  //WEEKEND
	WorkingDays             []DictItem        `json:"working_days"`           //only_saturday_and_sunday
	WorkingHours            []DictNameItem    `json:"working_hours"`          //HOURS_4
	WorkingTimeIntervals    []DictItem        `json:"working_time_intervals"` //from_four_to_six_hours_in_a_day
	WorkingTimeModes        []DictItem        `json:"working_time_modes"`     //start_after_sixteen
}

type AddressDraft struct {
	ID            string `json:"id"`
	ShowMetroOnly bool   `json:"show_metro_only"` //true
}

type AutoResponse struct {
	AcceptAutoResponse bool `json:"accept_auto_response"` //false
}

type KeySkill struct {
	Name string `json:"name"`
}

type Test struct {
	ID       string `json:"id"`
	Required bool   `json:"required"`
}

type VacancyProperty struct {
	PropertyType string `json:"property_type"` // HH_STANDARD
}
