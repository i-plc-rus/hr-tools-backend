package vacancyapimodels

import (
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"
)

type VacancyRequestData struct {
	CompanyID       string                 `json:"company_id"`        // ид компании
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

type VacancyRequestEditData struct {
	VacancyRequestData
	ApprovalStages
}

func (v VacancyRequestEditData) Validate() error {
	err := v.VacancyRequestData.Validate()
	if err != nil {
		return err
	}
	return v.ApprovalStages.Validate()
}

type VacancyRequestView struct {
	VacancyRequestData
	ID                   string              `json:"id"`
	CreationDate         time.Time           `json:"creation_date"`
	Status               models.VRStatus     `json:"status"`
	CompanyName          string              `json:"company_name"`
	DepartmentName       string              `json:"department_name"`
	JobTitleName         string              `json:"job_title_name"`
	City                 string              `json:"city"`
	CompanyStructName    string              `json:"company_struct_name"`
	ApprovalStages       []ApprovalStageView `json:"approval_stages"`
	ApprovalStageCurrent int                 `json:"approval_stage_current"`
	ApprovalStageIsLast  bool                `json:"approval_stage_is_last"`
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
		},
		ID:           rec.ID,
		CreationDate: rec.CreatedAt,
		Status:       rec.Status,
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
	approvalStages := []ApprovalStageView{}
	for _, item := range rec.ApprovalStages {
		approvalStages = append(approvalStages, ApprovalStageConvert(*item))
	}
	isLast, stage := rec.GetCurrentApprovalStage()
	if stage != nil {
		result.ApprovalStageCurrent = stage.Stage
		result.ApprovalStageIsLast = isLast
	}
	result.ApprovalStages = approvalStages
	return result
}
