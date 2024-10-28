package dbmodels

type ExtData struct {
	BaseModel
	SpaceID string `gorm:"type:varchar(36);index:idx_code"`
	Code    string `gorm:"type:varchar(255);index:idx_code"`
	Value   []byte
}
