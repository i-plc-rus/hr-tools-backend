package dbmodels

import (
	spaceapimodels "hr-tools-backend/models/api/space"
	"time"
)

type Space struct {
	BaseModel
	// Logo             string `gorm:"type:varchar(50)"`
	Referal          string `gorm:"type:varchar(500)"`
	ReferalID        string
	TypeBilling      string `gorm:"type:varchar(10)"`
	StartPay         time.Time
	StopPay          time.Time
	IsActive         bool
	OrganizationName string `gorm:"type:varchar(255)"` // Юридическое название компании
	Inn              string `gorm:"type:varchar(12)"`  // ИНН
	Kpp              string `gorm:"type:varchar(9)"`   // КПП
	OGRN             string `gorm:"type:varchar(15)"`  // ОГРН
	FullName         string `gorm:"type:varchar(255)"`
	DirectorName     string `gorm:"type:varchar(255)"`
	Web              string `gorm:"type:varchar(255)"`
	TimeZone         string `gorm:"type:varchar(255)"`
	Description      string
}

func (rec Space) ToModel() spaceapimodels.ProfileData {
	return spaceapimodels.ProfileData{
		OrganizationName: rec.OrganizationName,
		Web:              rec.Web,
		TimeZone:         rec.TimeZone,
		Description:      rec.Description,
		DirectorName:     rec.DirectorName,
	}
}
