package dbmodels

type Favorite struct {
	BaseModel
	VacancyID   string `gorm:"type:varchar(36);uniqueIndex:idx_user"`
	SpaceUserID string `gorm:"type:varchar(36);uniqueIndex:idx_user"`
	Selected    bool
}
