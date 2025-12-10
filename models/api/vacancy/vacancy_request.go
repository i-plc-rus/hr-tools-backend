package vacancyapimodels

import (
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	apimodels "hr-tools-backend/models/api"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

type VacancyRequestData struct {
	CompanyID       string                 `json:"company_id"`        // ид компании
	CompanyName     string                 `json:"company_name"`      // название компании
	DepartmentID    string                 `json:"department_id"`     // ид подразделения
	JobTitleID      string                 `json:"job_title_id"`      // ид штатной должности
	CityID          string                 `json:"city_id"`           // ид города
	CompanyStructID string                 `json:"company_struct_id"` // ид структуры компании
	VacancyName     string                 `json:"vacancy_name"`      // название вакансии
	Confidential    bool                   `json:"confidential"`      // конфиденциальная вакансия
	OpenedPositions int                    `json:"opened_positions"`  // кол-во открытых позиций
	Urgency         models.VRUrgency       `json:"urgency"`           // срочность
	RequestType     models.VRType          `json:"request_type"`      // тип вакансии
	SelectionType   models.VRSelectionType `json:"selection_type"`    // вид подбора
	PlaceOfWork     string                 `json:"place_of_work"`     // адрес места работы
	ChiefFio        string                 `json:"chief_fio"`         // фио непосредственного руководителя
	Requirements    string                 `json:"requirements"`      // требования/обязанности/условия
	Interviewer     string                 `json:"interviewer"`       // сотрудник проводящий интервью
	ShortInfo       string                 `json:"short_info"`        // краткая информация о комманде отдела
	Description     string                 `json:"description"`       // Коментарий к заявке
	OutInteraction  string                 `json:"out_interaction"`   // внешнее взаимодействие
	InInteraction   string                 `json:"in_interaction"`    // внутреннее взаимодействие
	Employment      models.Employment      `json:"employment"`        // Занятость
	Experience      models.Experience      `json:"experience"`        // Опыт работы
	Schedule        models.Schedule        `json:"schedule"`          // Режим работы
}

func (v VacancyRequestData) Validate() error {
	if v.VacancyName == "" {
		return errors.New("отсутсвует название вакансии")
	}
	if v.CityID == "" {
		return errors.New("отсутсвует ссылка на город")
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

type VacancyRequestCreateData struct {
	VacancyRequestEditData
	AsTemplate bool `json:"as_template"` // сохранить как шаблон
}

type VacancyRequestEditData struct {
	VacancyRequestData
	ApprovalTasks
}

func (v VacancyRequestEditData) Validate() error {
	err := v.VacancyRequestData.Validate()
	if err != nil {
		return err
	}
	return v.ApprovalTasks.Validate()
}

type VacancyRequestPreView struct {
	ID           string          `json:"id"`
	CreationDate time.Time       `json:"creation_date"`
	Status       models.VRStatus `json:"status"`
}

type VacancyRequestView struct {
	VacancyRequestData
	ID                string          `json:"id"`
	CreationDate      time.Time       `json:"creation_date"`
	Status            models.VRStatus `json:"status"`
	DepartmentName    string          `json:"department_name"`
	JobTitleName      string          `json:"job_title_name"`
	City              string          `json:"city"`
	CompanyStructName string          `json:"company_struct_name"`
	Pinned            bool            `json:"pinned"`
	Favorite          bool            `json:"favorite"`
	OpenVacancies     int             `json:"open_vacancies"` // кол-во вакансий открытых по заявке
	Comments          []CommentView   `json:"comments"`
}

func VacancyRequestConvert(rec dbmodels.VacancyRequest) VacancyRequestView {
	result := VacancyRequestView{
		VacancyRequestData: VacancyRequestData{
			VacancyName:     rec.VacancyName,
			Confidential:    rec.Confidential,
			OpenedPositions: rec.OpenedPositions,
			Urgency:         rec.Urgency,
			RequestType:     rec.RequestType,
			SelectionType:   rec.SelectionType,
			PlaceOfWork:     rec.PlaceOfWork,
			ChiefFio:        rec.ChiefFio,
			Interviewer:     rec.Interviewer,
			ShortInfo:       rec.ShortInfo,
			Requirements:    rec.Requirements,
			Description:     rec.Description,
			OutInteraction:  rec.OutInteraction,
			InInteraction:   rec.InInteraction,
			Employment:      rec.Employment,
			Experience:      rec.Experience,
			Schedule:        rec.Schedule,
		},
		ID:           rec.ID,
		CreationDate: rec.CreatedAt,
		Status:       rec.Status,
		Pinned:       rec.Pinned,
		Favorite:     rec.Favorite,
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
	result.OpenVacancies = len(rec.Vacancies)
	for _, comment := range rec.Comments {
		commentView := CommentView{
			Comment: Comment{
				Date:     comment.Date,
				AuthorID: comment.AuthorID,
				Comment:  comment.Comment,
			},
		}
		if comment.Author != nil {
			commentView.AuthorFIO = comment.Author.GetFullName()
		}
		result.Comments = append(result.Comments, commentView)
	}
	return result
}

type ExtVacancyInfo struct {
	Url    string                  `json:"url"`    //урл вакансии
	Status models.VacancyPubStatus `json:"status"` //статус публикации
	Reason string                  `json:"reason"` //описание статуса/ошибки
}

type VrSort struct {
	CreatedAtDesc bool `json:"created_at_desc"` // порядок сортировки false = ASC/ true = DESC
}

type VrFilter struct {
	apimodels.Pagination
	Favorite      bool                   `json:"favorite"`       // Избранные
	Search        string                 `json:"search"`         // Поиск по названию
	Statuses      []models.VRStatus      `json:"statuses"`       // Фильтр по статусам
	CityID        string                 `json:"city_id"`        // Фильтр по городу
	AuthorID      string                 `json:"author_id"`      // Фильтр по автору
	SelectionType models.VRSelectionType `json:"selection_type"` // Фильтр по виду подбора
	SearchPeriod  SearchPeriod           `json:"search_period"`  // Поиск по дате (1 - За день|2 - за 3 дня|3 - за неделю|4 - за 30 дней|5 - за пероид)
	SearchFrom    string                 `json:"search_from"`    // Период "с", при выборе search_period = 5 (в формате "21.09.2023")
	SearchTo      string                 `json:"search_to"`      // Период "по", при выборе search_period = 5 (в формате "21.09.2023")
	Sort          VrSort                 `json:"sort"`           // Сортировка
}

type SearchPeriod int

const (
	SearchByToday SearchPeriod = iota + 1
	SearchBy3Days
	SearchByWeek
	SearchByMonth
	SearchByPeriod
)

const ruLayout = "02.01.2006"

func (vr VrFilter) Validate() error {
	if vr.SearchPeriod != SearchByPeriod {
		return nil
	}
	if vr.SearchFrom == "" && vr.SearchTo == "" {
		return errors.New("не указан период поиска")
	}
	if vr.SearchFrom != "" {
		_, err := time.ParseInLocation(ruLayout, vr.SearchFrom, time.Now().Location())
		if err != nil {
			return errors.New("некорректный формат Периода \"с\", ожидается формат: \"21.09.2023\"")
		}
	}
	if vr.SearchTo != "" {
		_, err := time.ParseInLocation(ruLayout, vr.SearchTo, time.Now().Location())
		if err != nil {
			return errors.New("некорректный формат Периода \"по\", ожидается формат: \"21.09.2023\"")
		}
	}
	return nil
}

func (vr VrFilter) GetSearchFrom() time.Time {
	if vr.SearchFrom == "" {
		return time.Time{}
	}
	res, _ := time.ParseInLocation(ruLayout, vr.SearchFrom, time.Now().Location())
	return res
}

func (vr VrFilter) GetSearchTo() time.Time {
	if vr.SearchTo == "" {
		return time.Time{}
	}
	res, _ := time.ParseInLocation(ruLayout, vr.SearchTo, time.Now().Location())
	return res
}
