package dbmodels

import "time"

type Space struct {
	BaseModel
	Logo             string `gorm:"type:varchar(50)"`
	Referal          string `gorm:"type:varchar(500)"`
	ReferalID        string
	TypeBilling      string `gorm:"type:varchar(10)"`
	StartPay         time.Time
	StopPay          time.Time
	IsActive         bool
	OrganizationType string `gorm:"type:varchar(3)"`
	OrganizationName string `gorm:"type:varchar(255)"` // Юридическое название компании
	Inn              string `gorm:"type:varchar(12)"`  // ИНН
	Kpp              string `gorm:"type:varchar(9)"`   // КПП
	OGRN             string `gorm:"type:varchar(15)"`  // ОГРН
	FullName         string `gorm:"type:varchar(255)"`
	DirectorName     string `gorm:"type:varchar(255)"`
}
