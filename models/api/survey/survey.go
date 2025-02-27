package surveyapimodels

import (
	"encoding/json"
	dbmodels "hr-tools-backend/models/db"

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
