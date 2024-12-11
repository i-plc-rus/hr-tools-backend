package dbmodels

import "time"

type EmailVerify struct {
	ID            string `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	Email         string `gorm:"type:varchar(255)"`
	Code          string `gorm:"type:varchar(24)"`
	DateGenerated time.Time
	DateExpires   time.Time
	DateUsed      time.Time
}
