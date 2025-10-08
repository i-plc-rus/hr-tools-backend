package dbmodels

type MasaiSession struct {
	BaseModel
	VkStepID    string
	QuestionID  string
	ApplicantID string
	VideoPath   string
	EventID     string
}
