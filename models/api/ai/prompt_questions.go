package aiapimodels

type ValidateQPromptResponse struct {
	ResultPrompt string   `json:"result_prompt"`
	MissedTags   []string `json:"missed_tags"`
}

type ApplicantTemplateData struct {
	Vacancy                 string
	Requirements            string
	Applicant               string
	TypicalQuestions        string
	TypicalQuestionsAnswers string
}

func (q ApplicantTemplateData) GetTagList() []string {
	return []string{"{{.Vacancy}}", "{{.Requirements}}", "{{.Applicant}}", "{{.TypicalQuestions}}", "{{.TypicalQuestionsAnswers}}"}
}
