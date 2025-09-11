package surveyapimodels

type VkAiProvider interface {
	VkStep1(spaceID, vacancyID, vacancyInfo, applicantInfo, questions, applicantAnswers string) (resp VkStep1, err error)
}

type VkStep1 struct {
	Questions   []VkStep1Question `json:"questions"`
	ScriptIntro string            `json:"script_intro"`
	ScriptOutro string            `json:"script_outro"`
	Comments    map[string]string `json:"comments"`
}

type VkStep1Question struct {
	ID      string   `json:"id"`      // Идентификатор вопроса
	Text    string   `json:"text"`    // Текст вопроса
	Type    string   `json:"type"`    // Тип вопроса
	Options []string `json:"options"` // Варианты ответов
}
