package dbmodels

type LanguageData struct {
	BaseModel
	Code string `gorm:"type:varchar(36);index"`
	Name string `gorm:"type:varchar(255)"`
}
