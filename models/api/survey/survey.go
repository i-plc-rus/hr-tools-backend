package surveyapimodels

import (
	"encoding/json"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"time"

	"github.com/pkg/errors"
)

type HRSurvey struct {
	Questions []dbmodels.HRSurveyQuestion `json:"questions"`
}

type HRSurveyView struct {
	HRSurvey
	IsFilledOut bool `json:"is_filled_out"` // "анкета полностью заполнена"
}

func (s HRSurvey) Validate() error {
	if len(s.Questions) == 0 {
		return errors.New("в анкете отсутсвуют вопросы")
	}
	for _, question := range s.Questions {
		if question.QuestionID == "" {
			return errors.New("в одном из вопросов анкеты отсутсвует идентификатор вопроса")
		}
		if question.Selected != "Не подходит" {
			if question.QuestionText == "" {
				return errors.New("в одном из вопросов анкеты отсутсвует текст вопроса, для перегенерации выберите вариант \"Не подходит\"")
			}
			if question.QuestionType == "" {
				return errors.New("в одном из вопросов анкеты отсутсвует тип вопроса, для перегенерации выберите вариант \"Не подходит\"")
			}
		}
	}

	return nil
}

type ReGenerateSurvey struct {
	Questions []NotSuitableQuestion `json:"questions"`
}

type NotSuitableQuestion struct {
	QuestionID   string `json:"question_id"`
	QuestionText string `json:"question_text"`
	QuestionType string `json:"question_type"`
}

type VacancyPubData struct {
	Title        string `json:"title"`
	Requirements string `json:"requirements"`
	Employment   string `json:"employment,omitempty"`
	Experience   string `json:"experience,omitempty"`
	Schedule     string `json:"schedule,omitempty"`
	SalaryFrom   int    `json:"salary_from,omitempty"`
	SalaryTo     int    `json:"salary_to,omitempty"`
	JobTitle     string `json:"job_title,omitempty"`
}

type ApplicantParamsData struct {
	Education               string                     `json:"education"`
	HaveAdditionalEducation bool                       `json:"have_additional_education"`
	Employments             []string                   `json:"employments"`
	Schedules               []string                   `json:"schedules"`
	Languages               []dbmodels.Language        `json:"languages"`
	TripReadiness           string                     `json:"trip_readiness"`
	DriverLicenseTypes      []models.DriverLicenseType `json:"driver_license_types"`
	SearchStatus            string                     `json:"search_status"`
}

type ApplicantPubData struct {
	ResumeTitle     string              `json:"resume_title"`
	Salary          int                 `json:"salary"`
	BirthDate       time.Time           `json:"birth_date"`
	Citizenship     string              `json:"citizenship"`
	Gender          string              `json:"gender"`
	Relocation      string              `json:"relocation"`
	TotalExperience int                 `json:"total_experience"`
	Params          ApplicantParamsData `json:"params"`
}

type HRSurveyQuestionsPubData struct {
	Questions []HRSurveyQuestionPubData `json:"questions"`
}

type HRSurveyQuestionPubData struct {
	QuestionID   string `json:"question_id"`
	QuestionText string `json:"question_text"`
	QuestionType string `json:"question_type"`
	Answer       string `json:"answer"`
	Weight       int    `json:"weight"`
	Comment      string `json:"comment"`
}

func GetVacancyDataContent(rec dbmodels.Vacancy) (string, error) {
	result := VacancyPubData{
		Requirements: rec.Requirements,
		Employment:   rec.Employment.ToString(),
		Experience:   rec.Experience.ToString(),
		Schedule:     rec.Schedule.ToString(),
		Title:        rec.VacancyName,
	}
	if rec.JobTitle != nil {
		result.JobTitle = rec.JobTitle.Name
	}

	if rec.Salary.From != 0 || rec.Salary.To != 0 {
		result.SalaryFrom = rec.From
		result.SalaryTo = rec.To
	} else if rec.Salary.InHand != 0 {
		result.SalaryFrom = rec.Salary.InHand
		result.SalaryTo = rec.Salary.InHand

	}
	body, err := json.Marshal(result)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации структуры вакансии")
	}
	return string(body), nil
}

func GetApplicantDataContent(rec dbmodels.Applicant) (string, error) {
	result := ApplicantPubData{
		ResumeTitle:     rec.ResumeTitle,
		Salary:          rec.Salary,
		BirthDate:       rec.BirthDate,
		Citizenship:     rec.Citizenship,
		Gender:          rec.Gender.ToString(),
		Relocation:      rec.Relocation.ToString(),
		TotalExperience: rec.TotalExperience,
		Params: ApplicantParamsData{
			Education:               rec.Params.Education.ToString(),
			HaveAdditionalEducation: false,
			Employments:             []string{},
			Schedules:               []string{},
			Languages:               rec.Params.Languages,
			TripReadiness:           rec.Params.TripReadiness.ToString(),
			DriverLicenseTypes:      rec.Params.DriverLicenseTypes,
			SearchStatus:            rec.Params.SearchStatus.ToString(),
		},
	}
	for _, employment := range rec.Params.Employments {
		result.Params.Employments = append(result.Params.Employments, employment.ToString())
	}

	for _, schedule := range rec.Params.Schedules {
		result.Params.Schedules = append(result.Params.Schedules, schedule.ToString())
	}

	body, err := json.Marshal(result)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации структуры кандидата")
	}
	return string(body), nil
}

func GetHRDataContent(rec dbmodels.HRSurvey) (string, error) {
	result := HRSurveyQuestionsPubData{
		Questions: make([]HRSurveyQuestionPubData, 0, len(rec.Survey.Questions)),
	}

	for _, question := range rec.Survey.Questions {
		result.Questions = append(result.Questions, HRSurveyQuestionPubData{
			QuestionID:   question.QuestionID,
			QuestionText: question.QuestionText,
			QuestionType: question.QuestionType,
			Answer:       question.Selected,
			Weight:       question.Weight,
			Comment:      question.Comment,
		})
	}

	body, err := json.Marshal(result)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации структуры опросника HR")
	}
	return string(body), nil
}

type ApplicantSurveyView struct {
	ApplicantSurvey
	IsFilledOut bool `json:"is_filled_out"` // "анкета полностью заполнена"
}

type ApplicantSurvey struct {
	Questions []dbmodels.ApplicantSurveyQuestion `json:"questions"`
}

type ApplicantSurveyResponses struct {
	Responses []ApplicantSurveyAnswer `json:"responses"`
}

type ApplicantSurveyAnswer struct {
	QuestionID string `json:"question_id"` // Идентификатор вопроса
	Answer     string `json:"answer"`      // Ответ кандидата
}

func GetApplicantAnswersContent(rec dbmodels.ApplicantSurvey) (string, error) {
	result := ApplicantSurveyResponses{
		Responses: make([]ApplicantSurveyAnswer, 0, len(rec.Survey.Questions)),
	}

	for _, question := range rec.Survey.Questions {
		result.Responses = append(result.Responses, ApplicantSurveyAnswer{
			QuestionID: question.QuestionID,
			Answer:     question.Selected,
		})
	}

	body, err := json.Marshal(result)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации структуры опросника HR")
	}
	return string(body), nil
}

type VkStep0SurveyView struct {
	Questions []VkStep0Question `json:"questions"`
}

func (v VkStep0SurveyView) Content() (string, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации структуры вопросов шага 0")
	}
	return string(body), nil
}

type VkStep0Question struct {
	QuestionID   string   `json:"question_id"`   // Идентификатор вопроса
	QuestionText string   `json:"question_text"` // Текст вопроса
	QuestionType string   `json:"question_type"` // Тип вопроса
	Answers      []string `json:"answers"`       // Варианты ответов
}

type VkStep0SurveyAnswers struct {
	Answers []VkStep0Answer `json:"answers"`
}

type VkStep0Answer struct {
	QuestionID string `json:"question_id"` // Идентификатор вопроса
	Answer     string `json:"answer"`      // Варианты ответов
}

type VkStep0SurveyResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
