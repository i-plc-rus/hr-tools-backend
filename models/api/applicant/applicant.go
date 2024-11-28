package applicantapimodels

import (
	"fmt"
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	apimodels "hr-tools-backend/models/api"
	dbmodels "hr-tools-backend/models/db"
	"strings"
	"time"
)

type ApplicantView struct {
	ApplicantData
	ID                 string                 `json:"id"`                   // Идентификатор кандидата
	NegotiationID      string                 `json:"negotiation_id"`       // Идентификатор отклика во внешней системе
	NegotiationDate    string                 `json:"negotiation_date"`     // Дата отклика во внешней системе ДД.ММ.ГГГГ
	AcceptDate         string                 `json:"accept_date"`          // Дата добавления
	StartDate          string                 `json:"start_date"`           // Дата выхода
	ResumeID           string                 `json:"resume_id"`            // Идентификатор резюме во внешней системе
	ResumeTitle        string                 `json:"resume_title"`         // Заголовок резюме
	Status             models.ApplicantStatus `json:"status"`               // Статус кандидата
	SelectionStageID   string                 `json:"selection_stage_id"`   // Идентификатор этапа подбора кандидата
	SelectionStageName string                 `json:"selection_stage_name"` // Название этапа
	StageTime          string                 `json:"stage_time"`           // Время на этапе
	VacancyName        string                 `json:"vacancy_name"`         // Название вакансии
	FIO                string                 `json:"fio"`                  // ФИО кандидата
	Age                int                    `json:"age"`                  // возраст
}

type ApplicantViewExt struct {
	ApplicantView
	Tags               []string           `json:"tags"`
	PotentialDuplicate ApplicantDuplicate `json:"potential_duplicate"` // Возможный дубликат
	Duplicates         []string           `json:"duplicates"`          // Идентификатор кандидатов дубликатов
}

type ApplicantDuplicate struct {
	Found         bool                 `json:"found"`          // Найден
	DuplicateID   string               `json:"duplicate_id"`   // Идентификатор кандидата
	DuplicateType models.DuplicateType `json:"duplicate_type"` // Тип дубля (По автору резюме/По контактным данным)
}

type ApplicantData struct {
	VacancyID       string                   `json:"vacancy_id"`       // Идентификатор вакансии
	Source          models.ApplicantSource   `json:"source"`           // Источник кандидата
	FirstName       string                   `json:"first_name"`       // Имя
	LastName        string                   `json:"last_name"`        // Фамилия
	MiddleName      string                   `json:"middle_name"`      // Отчество
	Phone           string                   `json:"phone"`            // Телефон
	Email           string                   `json:"email"`            // Емайл
	Salary          int                      `json:"salary"`           // Желаемая ЗП
	Address         string                   `json:"address"`          // Адрес
	BirthDate       string                   `json:"birth_date"`       // Дата рождения ДД.ММ.ГГГГ
	Citizenship     string                   `json:"citizenship"`      // Гражданство
	Gender          models.GenderType        `json:"gender"`           // Пол кандидата
	Relocation      models.RelocationType    `json:"relocation"`       // Готовность к переезду
	TotalExperience int                      `json:"total_experience"` // Опыт работ в месяцах
	Comment         string                   `json:"comment"`          // Коментарий
	Params          dbmodels.ApplicantParams `json:"params"`           // Доподнительные параметры
	//PhotoUrl        string                `json:"photo_url"` //todo s3 photo
}

type Language struct {
	Name          string                   `json:"name"`           // Название языка
	LanguageLevel models.LanguageLevelType `json:"language_level"` // Уровень владения
}

func (a ApplicantData) Validate() error {
	if a.VacancyID == "" {
		return errors.New("не указана вакансия")
	}
	_, err := a.GetBirthDate()
	if err != nil {
		return errors.New("некоректный формат даты рождения")
	}
	return nil
}

func (a ApplicantData) GetBirthDate() (time.Time, error) {
	if a.BirthDate == "" {
		return time.Time{}, nil
	}
	date, err := time.Parse("02.01.2006", a.BirthDate)
	if err != nil {
		return time.Time{}, err
	}
	return date, nil
}

func ApplicantConvert(rec dbmodels.Applicant) ApplicantView {
	result := ApplicantView{
		ApplicantData: ApplicantData{
			VacancyID:       rec.VacancyID,
			Source:          rec.Source,
			FirstName:       rec.FirstName,
			LastName:        rec.LastName,
			MiddleName:      rec.MiddleName,
			Phone:           rec.Phone,
			Email:           rec.Email,
			Salary:          rec.Salary,
			Address:         rec.Address,
			BirthDate:       "",
			Citizenship:     rec.Citizenship,
			Gender:          rec.Gender,
			Relocation:      rec.Relocation,
			TotalExperience: rec.TotalExperience,
			Comment:         rec.Comment,
			Params:          rec.Params,
		},
		ID:                 rec.ID,
		NegotiationID:      rec.NegotiationID,
		ResumeID:           rec.ResumeID,
		ResumeTitle:        rec.ResumeTitle,
		Status:             rec.Status,
		SelectionStageID:   rec.SelectionStageID,
		SelectionStageName: "",
		StageTime:          "",
		VacancyName:        "",
		FIO:                "",
	}
	if !rec.BirthDate.IsZero() {
		difference := time.Now().Sub(rec.BirthDate)
		result.Age = int(difference.Hours() / 24 / 365)
		result.ApplicantData.BirthDate = rec.BirthDate.Format("02.01.2006")
	}
	if rec.SelectionStage != nil {
		result.SelectionStageName = rec.SelectionStage.Name
	}
	if !rec.NegotiationDate.IsZero() {
		result.NegotiationDate = rec.NegotiationDate.Format("02.01.2006")
	}
	if !rec.NegotiationAcceptDate.IsZero() {
		result.AcceptDate = rec.NegotiationAcceptDate.Format("02.01.2006")
	}
	if !rec.StartDate.IsZero() {
		result.StartDate = rec.StartDate.Format("02.01.2006")
	}
	if rec.Vacancy != nil {
		result.VacancyName = rec.Vacancy.VacancyName
	}
	fio := strings.TrimSpace(fmt.Sprintf("%v %v", rec.LastName, rec.FirstName))
	fio = strings.TrimSpace(fmt.Sprintf("%v %v", fio, rec.MiddleName))
	result.FIO = fio
	return result
}

type ApplicantFilter struct {
	apimodels.Pagination
	VacancyName         string                    `json:"vacancy_name"`          // Название вакансии
	Search              string                    `json:"search"`                // Поиск по ФИО/телефон/емайл/тег
	Relocation          *models.RelocationType    `json:"relocation"`            // Готовность к переезду
	AgeFrom             int                       `json:"age_from"`              // Возраст "от"
	AgeTo               int                       `json:"age_to"`                // Возраст "до"
	TotalExperienceFrom int                       `json:"total_experience_from"` // Опыт работ в месяцах "от"
	TotalExperienceTo   int                       `json:"total_experience_to"`   // Опыт работ в месяцах "до"
	City                string                    `json:"city"`                  // Город проживания
	StageName           string                    `json:"stage_name"`            // Этап
	Status              *models.ApplicantStatus   `json:"status"`                // Статус кандидата
	Source              *models.ApplicantSource   `json:"source"`                // Источник
	Tag                 string                    `json:"tag"`                   // Тэг
	AddedPeriod         *models.ApAddedPeriodType `json:"added_period"`          // Период добавления кандидата
	AddedDay            string                    `json:"added_day"`             // Дата добавления кандидата ДД.ММ.ГГГГ
	AddedType           *models.AddedType         `json:"added_type"`            // Тип добавления
	Sort                ApplicantSort             `json:"sort"`
}

func (a ApplicantFilter) Validate() error {
	_, err := a.GetAddedDay()
	if err != nil {
		return errors.New("некоректный формат даты добавления кандидата")
	}
	return nil
}

func (a ApplicantFilter) GetAddedDay() (time.Time, error) {
	if a.AddedDay == "" {
		return time.Time{}, nil
	}
	date, err := time.Parse("02.01.2006", a.AddedDay)
	if err != nil {
		return time.Time{}, err
	}
	return date, nil
}

type ApplicantSort struct {
	FioDesc        *bool `json:"fio_desc"`         // ФИО, порядок сортировки false = ASC/ true = DESC / nil = нет
	SalaryDesc     *bool `json:"salary_desc"`      // ЗП, порядок сортировки false = ASC/ true = DESC / nil = нет
	AcceptDateDesc *bool `json:"accept_date_desc"` // Дата добавления, порядок сортировки false = ASC/ true = DESC / nil = нет
}

type ApplicantNote struct {
	Note      string `json:"note"`       // Комментарий
	IsPrivate bool   `json:"is_private"` // Приватный
}

type RejectReasons struct {
	HrReasons        []string `json:"hr_reasons"`        //Отказы рекрутера
	HeadReasons      []string `json:"head_reasons"`      //Отказы руководителя
	ApplicantReasons []string `json:"applicant_reasons"` //Отказы кандидата
}
type RejectRequest struct {
	Reason    string                 `json:"reason"`    // Причина отказа
	Initiator models.RejectInitiator `json:"initiator"` // Инициатор отказа
}

func (r RejectRequest) Validate() error {
	if r.Reason == "" {
		return errors.New("не указана причина отказа")
	}
	if r.Initiator == "" {
		return errors.New("не указан инициатор отказа")
	}
	if r.Initiator != models.HrReject &&
		r.Initiator != models.HeadReject &&
		r.Initiator != models.ApplicantReject {
		return errors.New("некорректно указан инициатор отказа")
	}
	return nil
}
