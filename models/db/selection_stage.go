package dbmodels

import "hr-tools-backend/models"

type SelectionStage struct {
	BaseSpaceModel
	VacancyID  string `gorm:"type:varchar(36)"`
	StageOrder int
	Name       string `gorm:"type:varchar(255)"`
	StageType  string `gorm:"type:varchar(255)"`
	CanDelete  bool
	LimitValue int64
	LimitType  models.LimitType `gorm:"type:varchar(50)"`
}

const (
	NegotiationStage      string = "Откликнулся"
	AddedStage            string = "Добавлен"
	ScreenStage           string = "Скриннинг"
	ManagerInterviewStage string = "Интервью с менеджером"
	ClientInterviewStage  string = "Интервью с заказчиком"
	OfferStage            string = "Оффер"
	HiredStage            string = "Принят"
)

var DefaultSelectionStages = []string{NegotiationStage, AddedStage, ScreenStage, ManagerInterviewStage, ClientInterviewStage, OfferStage, HiredStage}
