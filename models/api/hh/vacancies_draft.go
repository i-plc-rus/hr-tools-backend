package hhapimodels

type VacancyDraftRequest struct {
	ClosedForApplicants     *bool             `json:"closed_for_applicants"` //false
	VacancyProperties       []VacancyProperty `json:"vacancy_properties"`
	AcceptHandicapped       *bool             `json:"accept_handicapped"`        //true
	AcceptIncompleteResumes *bool             `json:"accept_incomplete_resumes"` //false
	AcceptLaborContract     *bool             `json:"accept_labor_contract"`     //true
	Address                 *AddressDraft     `json:"address"`
	AgeRestriction          *DictItem         `json:"age_restriction"` //AGE_14_PLUS
	AllowMessages           *bool             `json:"allow_messages"`  // Разрешение сообщений true
	Areas                   []DictItem        `json:"areas"`
	AutoResponse            AutoResponse      `json:"auto_response"`
	BrandedTemplate         *DictItem         `json:"branded_template"` //"marketing"
	CivilLawContracts       *[]DictItem       `json:"civil_law_contracts"`
	Code                    *string           `json:"code"` //"код-1234
	Contacts                *Contacts         `json:"contacts,omitempty"`
	Department              *DictItem         `json:"department"`
	Description             string            `json:"description"`
	DriverLicenseTypes      *[]DictItem       `json:"driver_license_types"`
	EmploymentFrom          *DictItem         `json:"employment_form,omitempty"` //Тип занятости
	Experience              *DictItem         `json:"experience,omitempty"`      // Опыт работы
	Internship              *bool             `json:"internship"`                //false
	KeySkills               *[]KeySkill       `json:"key_skills"`
	Languages               *[]Language       `json:"languages"`
	Name                    string            `json:"name"`         // Менеджер по продажам
	NightShifts             *bool             `json:"night_shifts"` //true
	ProfessionalRoles       *[]DictItem       `json:"professional_roles"`
	ResponseLetterRequired  *bool             `json:"response_letter_required"` //true
	ResponseNotifications   *bool             `json:"response_notifications"`   //true
	SalaryRange             *SalaryRange      `json:"salary_range,omitempty"`
	ScheduleAt              *DictItem         `json:"schedule_at,omitempty"` // Время запланированной публикации вакансии
	Test                    *Test             `json:"test"`
	WithZp                  *bool             `json:"with_zp"`
	WorkFormat              *[]DictItem       `json:"work_format"`
	WorkScheduleByDays      *[]DictItem       `json:"work_schedule_by_days"` //WEEKEND
	WorkingHours            *[]DictItem       `json:"working_hours"`         //HOURS_4
	FlyInFlyOutDuration     *[]DictItem       `json:"fly_in_fly_out_duration"`
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
