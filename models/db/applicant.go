package dbmodels

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	"time"
)

type Applicant struct {
	BaseSpaceModel
	VacancyID             string                   `gorm:"type:varchar(36)" comment:"Идентификатор вакансии"`
	Vacancy               *Vacancy                 `gorm:"foreignKey:VacancyID"`
	NegotiationID         string                   `gorm:"type:varchar(255);index:idx_negotiation" comment:"Идентификатор отклика во внешней системе"` // ид отклика во внешней системе
	ResumeID              string                   `gorm:"index;type:varchar(255)" comment:"Идентификатор резюме во внешней системе"`                  // ид резюме во внешней системе
	ResumeTitle           string                   `comment:"Заголовок резюме"`
	Source                models.ApplicantSource   `gorm:"index:idx_negotiation" comment:"Источник"`
	NegotiationDate       time.Time                `comment:"Дата отзыва"`
	NegotiationAcceptDate time.Time                `comment:"Дата добавления"` // дата принятия по отклику/дата ручного добавления
	Status                models.ApplicantStatus   `gorm:"index" comment:"Статус"`
	NegotiationStatus     models.NegotiationStatus `comment:"Статус отклика"`
	FirstName             string                   `gorm:"type:varchar(255)" comment:"Имя"`
	LastName              string                   `gorm:"type:varchar(255)" comment:"Фамилия"`
	MiddleName            string                   `gorm:"type:varchar(255)" comment:"Отчество"`
	Phone                 string                   `gorm:"type:varchar(255)" comment:"Телефон"`
	Email                 string                   `gorm:"type:varchar(255)" comment:"Email"`
	Salary                int                      `comment:"Ожидаемая дата ЗП"`
	Address               string                   `comment:"Адрес"`
	BirthDate             time.Time                `comment:"Дата рождения"`
	Citizenship           string                   `gorm:"type:varchar(255)" comment:"Гражданство"`           // Гражданство
	Gender                models.GenderType        `gorm:"type:varchar(50)" comment:"Пол кандидата"`          // Пол кандидата
	Relocation            models.RelocationType    `gorm:"type:varchar(100)" comment:"Готовность к переезду"` // Готовность к переезду
	TotalExperience       int                      `comment:"Опыт работ в месяцах"`
	Comment               string                   `comment:"Комментарий"`
	Params                ApplicantParams          `gorm:"type:jsonb"`
	PhotoUrl              string                   `gorm:"type:varchar(500)"` //todo s3 photo
	SelectionStageID      string                   `gorm:"type:varchar(36)" comment:"Идентификатор этапа подбора"`
	SelectionStage        *SelectionStage          `gorm:"foreignKey:SelectionStageID"`
	Tags                  pq.StringArray           `gorm:"type:text[]" comment:"Тэги"`
	ExtApplicantID        string                   `gorm:"type:varchar(255)" comment:"Идентификатор кандидата во внешней системе"` // Идентификатор кандидата во внешней системе
	NotDuplicates         pq.StringArray           `gorm:"type:text[]"`                                                            // ид кандидатов помеченные как разные кандидаты
	Duplicates            []Applicant              `gorm:"foreignKey:DuplicateID"`                                                 // Список дублей
	DuplicateID           *string                  `gorm:"type:varchar(36)" comment:"Идентификатор кандидата дубликата"`           // текущая запись является дублем кандидата (Идентификатор кандидата)
	StartDate             time.Time                `comment:"Дата выхода"`
}

func (j ApplicantParams) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *ApplicantParams) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

type ApplicantParams struct {
	Education               models.EducationType       `json:"education" comment:"Образование"`                            // Образование
	HaveAdditionalEducation bool                       `json:"have_additional_education" comment:"Повышение квалификации"` // Повышение квалификации, курсы
	Employments             []models.Employment        `json:"employments" comment:"Занятость"`                            // Занятость
	Schedules               []models.Schedule          `json:"schedules" comment:"График работы"`                          // График работы
	Languages               []Language                 `json:"languages" comment:"Знание языков"`                          // Знание языков
	TripReadiness           models.TripReadinessType   `json:"trip_readiness" comment:"Командировки"`                      // Готовность к командировкам
	DriverLicenseTypes      []models.DriverLicenseType `json:"driver_license_types" comment:"Водительсике права"`          // Водительсике права
	SearchStatus            models.SearchStatusType    `json:"search_status" comment:"Статус поиска работы"`               // Статус поиска работы
}

type Language struct {
	Name          string                   `json:"name"`
	LanguageLevel models.LanguageLevelType `json:"language_level"`
}

type ApplicantExt struct {
	Applicant
	AuthorFirstName string
	AuthorLastName  string
}

type ApplicantsStage struct {
	VacancyID        string
	SelectionStageID string
	Total            int
}

func (a Applicant) IsAllowStatusChange(newStatus models.NegotiationStatus) (bool, error) {
	if newStatus != models.NegotiationStatusWait &&
		newStatus != models.NegotiationStatusRejected &&
		newStatus != models.NegotiationStatusAccepted {
		return false, errors.New("неизвестный статус")
	}
	if a.NegotiationStatus == newStatus {
		return false, nil
	}
	if a.Status == models.ApplicantStatusInProcess {
		return false, errors.New("смена статуса отклика недоступна, кандидат в процессе рассмотрения")
	}
	if a.Status == models.ApplicantStatusRejected {
		return false, errors.New("смена статуса отклика недоступна, кандидат уже отклонен")
	}
	if a.Status == models.ApplicantStatusArchive {
		return false, errors.Errorf("смена статуса отклика недоступна, кандидат находится в статусе '%v'", models.ApplicantStatusArchive)
	}
	if a.NegotiationStatus == models.NegotiationStatusAccepted {
		return false, errors.New("смена статуса отклика недоступна, отклик уже принят")
	}
	return true, nil
}

type NegotiationFilter struct {
	VacancyID         string                     `json:"vacancy_id"`          // идентификатор вакансии
	Search            string                     `json:"search"`              // поиск по ФИО/телефон/емайл
	Education         *models.EducationType      `json:"education"`           // Образование
	Experience        *models.ExperienceType     `json:"experience"`          // Опыт
	ResponsePeriod    *models.ResponsePeriodType `json:"response_period"`     // Период отклика на вакансию
	City              string                     `json:"city"`                // Город проживания
	Employment        *models.Employment         `json:"employment"`          // Занятость
	Schedule          *models.Schedule           `json:"schedule"`            // График работы
	Language          string                     `json:"language"`            // Знание языка
	LanguageLevel     *models.LanguageLevelType  `json:"language_level"`      // Уровень знания языка
	Gender            *models.GenderType         `json:"gender"`              // Пол кандидата
	TripReadiness     *models.TripReadinessType  `json:"trip_readiness"`      // Готовность к командировкам
	Citizenship       string                     `json:"citizenship"`         // Гражданство
	SalaryFrom        int                        `json:"salary_from"`         // Уровень дохода от
	SalaryTo          int                        `json:"salary_to"`           // Уровень дохода до
	SalaryProvided    *bool                      `json:"salary_provided"`     // Указан доход
	Source            *models.ApplicantSource    `json:"source"`              // Источник
	DriverLicence     []models.DriverLicenseType `json:"driver_licence"`      // Водительсике права
	JobSearchStatuses *models.SearchStatusType   `json:"job_search_statuses"` // Статус поиска работы
	SearchLabel       *models.SearchLabelType    `json:"search_label"`        // Метка поиска резюме
	AdvancedTraining  *bool                      `json:"advanced_training"`   // Повышение квалификации, курсы
}

func (n NegotiationFilter) Validate() error {
	if n.VacancyID == "" {
		return errors.New("не указан идентификатор вакансии")
	}
	return nil
}

type DuplicateApplicantFilter struct {
	VacancyID      string
	FIO            string
	Phone          string
	Email          string
	ExtApplicantID string
}

func (a Applicant) IsMarkAsNotDuplicate(source Applicant) bool {
	for _, id := range a.NotDuplicates {
		if id == source.ID {
			return true
		}
	}
	for _, id := range source.NotDuplicates {
		if id == a.ID {
			return true
		}
	}
	return false
}
