package vacancyapimodels

import (
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"
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
	CompanyName       string               `json:"company_name"`
	DepartmentName    string               `json:"department_name"`
	JobTitleName      string               `json:"job_title_name"`
	City              string               `json:"city"`
	CompanyStructName string               `json:"company_struct_name"`
	Status            models.VacancyStatus `json:"status"`
	Pinned            bool                 `json:"pinned"`
	Favorite          bool                 `json:"favorite"`
	HH                ExternalLink         `json:"hh"`
}

type VacancyView struct {
	VacancyData
	VacancyInfo
	External     ExternalData `json:"external"`
	ID           string       `json:"id"`
	CreationDate time.Time    `json:"creation_date"`
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
		},
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
	return result
}
