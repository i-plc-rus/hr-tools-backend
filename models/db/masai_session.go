package dbmodels

import "time"

type MasaiSession struct {
	BaseModel
	VkStepID    string
	QuestionID  string
	ApplicantID string
	VideoPath   string
	EventID     string
	ExpiresAt   *time.Time
}
