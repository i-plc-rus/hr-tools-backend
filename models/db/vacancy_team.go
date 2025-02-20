package dbmodels

type VacancyTeam struct {
	BaseSpaceModel
	UserID      string
	VacancyID   string     `gorm:"type:varchar(36);index:idx_vacancy"`
	SpaceUser   *SpaceUser `gorm:"foreignKey:UserID"`
	Responsible bool
}
