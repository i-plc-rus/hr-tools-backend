package dbmodels

type VacancyTeam struct {
	BaseSpaceModel
	VacancyID   string     `gorm:"type:varchar(36);index:idx_vacancy"`
	SpaceUser   *SpaceUser `gorm:"foreignKey:ID"`
	Responsible bool
}
