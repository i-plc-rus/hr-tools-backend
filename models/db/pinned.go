package dbmodels

type Pinned struct {
	BaseModel
	VacancyID   string `gorm:"type:varchar(36);uniqueIndex:idx_user"`
	SpaceUserID string `gorm:"type:varchar(36);uniqueIndex:idx_user"`
	Selected    bool
}

type VrPinned struct {
	BaseModel
	VacancyRequestID string `gorm:"type:varchar(36);uniqueIndex:idx_user"`
	SpaceUserID      string `gorm:"type:varchar(36);uniqueIndex:idx_user"`
	Selected         bool
}
