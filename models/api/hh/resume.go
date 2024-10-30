package hhapimodels

type ResumeResponse struct {
	Age                   int        `json:"age"`
	AlternateUrl          string     `json:"alternate_url"`
	Area                  DictData   `json:"area"`
	BirthDate             string     `json:"birth_date"`
	BusinessTripReadiness DictData   `json:"business_trip_readiness"`
	Citizenship           []DictData `json:"citizenship"`
	Contact               []Contact  `json:"contact"`
	CreatedAt             string     `json:"created_at"`
	DriverLicenseTypes    DictData   `json:"driver_license_types"`
	Education             interface{}
	Employments           []DictData    `json:"employments"`
	Experience            []interface{} `json:"experience"`
	Gender                DictData      `json:"gender"`
	ID                    string        `json:"id"`
	Language              []interface{} `json:"language"`
	FirstName             string        `json:"first_name"`
	LastName              string        `json:"last_name"`
	MiddleName            string        `json:"middle_name"`
	Photo                 string        `json:"photo"`
	Portfolio             []interface{} `json:"portfolio"`
	ProfessionalRoles     []interface{} `json:"professional_roles"`
	Recommendation        []interface{} `json:"recommendation"`
	Relocation            []interface{} `json:"relocation"`
	Salary                ResumeSalary  `json:"salary"`
	Schedules             []DictData    `json:"schedules"`
	SkillSet              []string      `json:"skill_set"`
	Title                 string        `json:"title"`
}

type Contact struct {
	Comment   string      `json:"comment"`
	Preferred bool        `json:"preferred"`
	Type      ContactType `json:"type"`
	Value     interface{} `json:"value"`
}

type ContactType struct {
	ID   PreferredContactType `json:"id"`
	Name string               `json:"name"`
}

type ResumeSalary struct {
	Currency string `json:"currency"`
	Amount   int    `json:"amount"`
}
