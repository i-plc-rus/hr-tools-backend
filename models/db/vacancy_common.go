package dbmodels

import (
	"database/sql/driver"
	"encoding/json"
	"hr-tools-backend/models"
	"regexp"

	"github.com/pkg/errors"
)

func (j VacancyProps) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *VacancyProps) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

func (v *VacancyProps) Validate() error {
	if v.EmploymentForm != nil {
		if err := v.EmploymentForm.Validate(false); err != nil {
			return err
		}
	}

	if v.DriverLicenseTypes != nil {
		for _, value := range *v.DriverLicenseTypes {
			if err := value.Validate(false); err != nil {
				return err
			}
		}
	}

	if v.WorkFormat != nil {
		for _, value := range *v.WorkFormat {
			if err := value.Validate(false); err != nil {
				return err
			}
		}
	}
	if v.WorkScheduleByDays != nil {
		for _, value := range *v.WorkScheduleByDays {
			if err := value.Validate(false); err != nil {
				return err
			}
		}
	}
	if v.WorkingHours != nil {
		for _, value := range *v.WorkingHours {
			if err := value.Validate(false); err != nil {
				return err
			}
		}
	}
	if v.FlyInFlyOutDuration != nil {
		for _, value := range *v.FlyInFlyOutDuration {
			if err := value.Validate(false); err != nil {
				return err
			}
		}
	}
	if v.SalaryRange.From <= 0 && v.SalaryRange.To <= 0 {
		return errors.New("не указана зарплата")
	}
	if v.Contacts != nil && len(v.Contacts.Phones) != 0 {
		re := regexp.MustCompile(`^\d+$`)
		for _, phone := range v.Contacts.Phones {
			if !re.MatchString(phone) {
				return errors.Errorf("Контактный телефон %v указан некорректно", phone)
			}
		}
	}

	if v.KeySkills != nil && len(*v.KeySkills) != 0 {
		for _, keySkill := range *v.KeySkills {
			// для проверки строки надо использовать utf8.RuneCountInString,
			// но перестрахуемся, вдруг на HH проверяют по байтно
			if len(keySkill) > 100 {
				return errors.New("Максимальная длина названия ключевого навыка не должно превышеть 100 символов")
			}
		}
	}

	return nil
}

type VacancyProps struct {
	KeySkills               *[]string                     `json:"key_skills,omitempty"`              // Ключевые навыки
	Languages               *[]VacancyLanguage            `json:"languages,omitempty"`               // Языки вакансии
	DriverLicenseTypes      *[]models.DriverLicenseType   `json:"driver_license_types,omitempty"`    // Права (опционально)
	Contacts                *VacancyContacts              `json:"contacts,omitempty"`                // Контакты
	WithZp                  *bool                         `json:"with_zp,omitempty"`                 // Разместить на Зарплата.ру
	SalaryRange             VacancySalaryRange            `json:"salary_range"`                      // Зарплата
	EmploymentForm          *models.EmploymentForm        `json:"employment_form"`                   // Занятость
	WorkFormat              *[]models.WorkFormat          `json:"work_format,omitempty"`             // Формат работы
	WorkingHours            *[]models.WorkingHours        `json:"working_hours,omitempty"`           // Рабочие часы в день
	WorkScheduleByDays      *[]models.WorkScheduleByDays  `json:"work_schedule_by_days,omitempty"`   // График работы
	NightShifts             *bool                         `json:"night_shifts"`                      // Ночные смены
	FlyInFlyOutDuration     *[]models.FlyInFlyOutDuration `json:"fly_in_fly_out_duration,omitempty"` // Длительность вахты
	Internship              *bool                         `json:"internship,omitempty"`              // Стажировка
	AllowMessages           *bool                         `json:"allow_messages"`                    // Возможность переписки с кандидатами по данной вакансии
	ResponseLetterRequired  *bool                         `json:"response_letter_required"`          // Обязательно ли заполнять сообщение при отклике на вакансию
	ResponseNotifications   *bool                         `json:"response_notifications"`            // Уведомлять ли менеджера о новых откликах
	AutoResponse            *bool                         `json:"auto_response"`                     // Настройки для автооткликов
	AcceptHandicapped       *bool                         `json:"accept_handicapped"`                // Соискатель с инвалидностью
	AcceptIncompleteResumes *bool                         `json:"accept_incomplete_resumes"`         // Разрешен ли отклик на вакансию неполным резюме
	AcceptLaborContract     *bool                         `json:"accept_labor_contract"`             // Указание на возможность приёма кандидата на работу по трудовому договору
	CivilLawContracts       *[]models.CivilLawContract    `json:"civil_law_contracts"`               // Договор гражданско-правового характера
	AgeRestriction          *models.AgeRestriction        `json:"age_restriction"`                   // Указание на возможность приёма кандидата на работу по трудовому договору в соответствии с трудовым законодательством
	ClosedForApplicants     *bool                         `json:"closed_for_applicants"`             // Закрытая или открытая вакансия
}

type VacancyLanguage struct {
	LanguageID string                   `json:"language_id"`
	Level      models.LanguageLevelType `json:"level"`
}

type VacancyContacts struct {
	Name   string   `json:"name"`
	Email  string   `json:"email"`
	Phones []string `json:"phones"`
}

type VacancySalaryRange struct {
	From  int  `json:"from"`
	To    int  `json:"to"`
	Gross bool `json:"gross"`
}
