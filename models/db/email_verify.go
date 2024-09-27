package dbmodels

import "time"

type EmailVerify struct {
	Email         string `gorm:"type:varchar(255)"`
	Code          string `gorm:"type:varchar(24)"`
	DateGenerated time.Time
	DateExpires   time.Time
	DateUsed      time.Time
}
