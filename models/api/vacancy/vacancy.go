package vacancyapimodels

import (
	"hr-tools-backend/models"
	apimodels "hr-tools-backend/models/api"
	dbmodels "hr-tools-backend/models/db"
	"time"

	"github.com/pkg/errors"
)

type VacancyData struct {
	VacancyRequestID string                 `json:"vacancy_request_id"` // ид заявки на вакансию
	CompanyID        string                 `json:"company_id"`         // ид компании
	DepartmentID     string                 `json:"department_id"`      // ид подразделения
	JobTitleID       string                 `json:"job_title_id"`       // ид штатной должности
	CityID           string                 `json:"city_id"`            // ид города
	CompanyStructID  string                 `json:"company_struct_id"`  // ид структуры компании
	VacancyName      string                 `json:"vacancy_name"`       // название вакансии
	OpenedPositions  int                    `json:"opened_positions"`   // кол-во открытых позиций
	Urgency          models.VRUrgency       `json:"urgency"`            // срочность
	RequestType      models.VRType          `json:"request_type"`       // тип вакансии
	SelectionType    models.VRSelectionType `json:"selection_type"`     // вид подбора
	PlaceOfWork      string                 `json:"place_of_work"`      // адрес места работы
	ChiefFio         string                 `json:"chief_fio"`          // фио непосредственного руководителя
	Requirements     string                 `json:"requirements"`       // требования/обязанности/условия
	Salary           Salary                 `json:"salary"`             // ожидания по зп
	Employment       models.Employment      `json:"employment"`         // Занятость
	Experience       models.Experience      `json:"experience"`         // Опыт работы
	Schedule         models.Schedule        `json:"schedule"`           // Режим работы
}

func (v VacancyData) Validate(isFromRequest bool) error {
	if v.VacancyName == "" {
		return errors.New("не указано название вакансии")
	}
	if v.JobTitleID == "" {
		return errors.New("отсутсвует ссылка на штатную должность")
	}
	if v.CityID == "" {
		return errors.New("отсутсвует ссылка на город")
	}

	if v.ChiefFio == "" {
		return errors.New("не указано фио непосредственного руководителя")
	}
	if !isFromRequest {
		if v.Salary.InHand == 0 {
			return errors.New("не указана сумма заработной платы 'на руки'")
		}
		if v.Salary.From == 0 {
			return errors.New("не указана сумма заработной платы 'от'")
		}
		if v.Salary.To == 0 {
			return errors.New("не указана сумма заработной платы 'до'")
		}
	}
	if v.OpenedPositions <= 0 {
		return errors.New("не указано количество вакантных позиций")
	}

	if err := v.Urgency.Validate(); err != nil {
		return err
	}
	if err := v.RequestType.Validate(); err != nil {
		return err
	}
	if err := v.SelectionType.Validate(); err != nil {
		return err
	}
	return nil
}

type VacancyInfo struct {
	CompanyName         string               `json:"company_name"`
	DepartmentName      string               `json:"department_name"`
	JobTitleName        string               `json:"job_title_name"`
	City                string               `json:"city"`
	CompanyStructName   string               `json:"company_struct_name"`
	Status              models.VacancyStatus `json:"status"`
	Pinned              bool                 `json:"pinned"`
	Favorite            bool                 `json:"favorite"`
	HH                  ExternalLink         `json:"hh"`
	AuthorID            string               `json:"author_id"`                       // Идентификатор автора вакансии
	AuthorFullName      string               `json:"author_full_name"`                // ФИО автора вакансии
	ResponsibleID       string               `json:"responsible_id,omitempty"`        // Идентификатор ответственного
	ResponsibleFullName string               `json:"responsible_full_name,omitempty"` // ФИО ответственного
}

type VacancyView struct {
	VacancyData
	VacancyInfo
	External        ExternalData         `json:"external"`
	ID              string               `json:"id"`
	CreationDate    time.Time            `json:"creation_date"`
	SelectionStages []SelectionStageView `json:"selection_stages"` // этапы подбора
}

type Salary struct {
	From     int `json:"from"`
	To       int `json:"to"`
	ByResult int `json:"by_result"`
	InHand   int `json:"in_hand"`
}

type ExternalData struct {
	HeadHunter ExternalLink `json:"head_hunter"`
}

type ExternalLink struct {
	ID  string `json:"id"`
	Url string `json:"url"`
}

func VacancyConvert(rec dbmodels.VacancyExt) VacancyView {
	result := VacancyView{
		VacancyData: VacancyData{
			CompanyID:       "",
			DepartmentID:    "",
			JobTitleID:      "",
			CityID:          "",
			CompanyStructID: "",
			VacancyName:     rec.VacancyName,
			OpenedPositions: rec.OpenedPositions,
			Urgency:         rec.Urgency,
			RequestType:     rec.RequestType,
			SelectionType:   rec.SelectionType,
			PlaceOfWork:     rec.PlaceOfWork,
			ChiefFio:        rec.ChiefFio,
			Requirements:    rec.Requirements,
			Salary: Salary{
				From:     rec.From,
				To:       rec.To,
				ByResult: rec.ByResult,
				InHand:   rec.InHand,
			},
			Employment: rec.Employment,
			Experience: rec.Experience,
			Schedule:   rec.Schedule,
		},
		ID:           rec.ID,
		CreationDate: rec.CreatedAt,
		VacancyInfo: VacancyInfo{
			CompanyName:       "",
			DepartmentName:    "",
			JobTitleName:      "",
			City:              "",
			CompanyStructName: "",
			Pinned:            rec.Pinned,
			Favorite:          rec.Favorite,
			Status:            rec.Status,
			AuthorID:          rec.AuthorID,
			ResponsibleID:     rec.ResponsibleID,
		},
	}
	if rec.Author != nil {
		result.AuthorFullName = rec.Author.GetFullName()
	}
	if rec.ResponsibleUser != nil {
		result.ResponsibleFullName = rec.ResponsibleUser.GetFullName()
	}
	if rec.CompanyID != nil {
		result.CompanyID = *rec.CompanyID
	}
	if rec.Company != nil {
		result.CompanyName = rec.Company.Name
	}
	if rec.DepartmentID != nil {
		result.DepartmentID = *rec.DepartmentID
	}
	if rec.Department != nil {
		result.DepartmentName = rec.Department.Name
	}
	if rec.JobTitleID != nil {
		result.JobTitleID = *rec.JobTitleID
	}
	if rec.JobTitle != nil {
		result.JobTitleName = rec.JobTitle.Name
	}
	if rec.CityID != nil {
		result.CityID = *rec.CityID
	}
	if rec.City != nil {
		result.City = rec.City.Address
	}
	if rec.CompanyStructID != nil {
		result.CompanyStructID = *rec.CompanyStructID
	}
	if rec.CompanyStruct != nil {
		result.CompanyStructName = rec.CompanyStruct.Name
	}
	if rec.VacancyRequestID != nil {
		result.VacancyRequestID = *rec.VacancyRequestID
	}

	if rec.HhID != "" {
		result.External.HeadHunter.ID = rec.HhID
		result.External.HeadHunter.Url = rec.HhUri
	}
	result.SelectionStages = make([]SelectionStageView, 0, len(rec.SelectionStages))
	for _, stage := range rec.SelectionStages {
		result.SelectionStages = append(result.SelectionStages, SelectionStageConvert(stage))
	}
	return result
}

type VacancySort struct {
	CreatedAtDesc bool `json:"created_at_desc"` // порядок сортировки false = ASC/ true = DESC
}

type VacancyFilter struct {
	apimodels.Pagination
	VacancyRequestID  string                 `json:"request_id"`         // Идентификатор запроса на вакансию
	Favorite          bool                   `json:"favorite"`           // Отображать избранные
	Search            string                 `json:"search"`             // Поиск
	Statuses          []models.VacancyStatus `json:"statuses"`           // Фильтр по статусам
	CityID            string                 `json:"city_id"`            // Фильтр по идентификатору города
	DepartmentID      string                 `json:"department_id"`      // Фильтр по идентификатору подразделения
	SelectionType     models.VRSelectionType `json:"selection_type"`     // Фильтр по виду подбора
	RequestType       models.VRType          `json:"request_type"`       // Фильтр по тип вакансии
	Urgency           models.VRUrgency       `json:"urgency"`            // Фильтр по срочности
	AuthorID          string                 `json:"author_id"`          // Фильтр по автору вакансии
	RequestAuthorID   string                 `json:"request_author_id"`  // Фильтр по автору запроса на вкансию
	Sort              VacancySort            `json:"sort"`               // Сортировка
	Tab               VacancyTab             `json:"tab"`                // Фильтр по вкладке: 0 - Все, 1 - Мои, 2 - Другие, 3 - Архив
	AuthorSearch      string                 `json:"author_search"`      // Поиск по ФИО автора
	ResponsibleSearch string                 `json:"responsible_search"` // Поиск по ФИО ответственного
}

type VacancyTab int

const (
	VacancyTabAll   VacancyTab = iota
	VacancyTabMy    VacancyTab = iota
	VacancyTabOther VacancyTab = iota
	VacancyTabArch  VacancyTab = iota
)

type StatusChangeRequest struct {
	Status models.VacancyStatus `json:"status"` // новый статус вакансии
}

func (s StatusChangeRequest) Validate() error {
	if s.Status == "" {
		return errors.New("не указан новый статус вакансии")
	}
	if s.Status != models.VacancyStatusOpened && s.Status != models.VacancyStatusCanceled &&
		s.Status != models.VacancyStatusSuspended && s.Status != models.VacancyStatusClosed {
		return errors.New("указан некорректный статус")
	}
	return nil
}
