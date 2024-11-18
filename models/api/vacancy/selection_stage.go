package vacancyapimodels

import (
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
)

type SelectionStageAdd struct {
	Name       string           `json:"name"`        // Название этапа
	StageType  string           `json:"stage_type"`  // Тип этапа
	LimitValue int64            `json:"limit_value"` // Лимит времени на этапе
	LimitType  models.LimitType `json:"limit_type"`  // Лимит времени на этапе - единицы измерения
}

func (s SelectionStageAdd) Validate() error {
	if s.Name == "" {
		return errors.New("не указано название этапа подбора кандидата")
	}
	if s.LimitValue > 0 &&
		(s.LimitType != models.LimitTypeMin && s.LimitType != models.LimitTypeHour &&
			s.LimitType != models.LimitTypDay && s.LimitType != models.LimitTypeWeek) {
		return errors.New("не указаны единицы измерения лимита времени на этапе")
	}
	return nil
}

type SelectionStageView struct {
	ID         string           `json:"id"`          // Идентификатор этапа подбора кандидата
	StageOrder int              `json:"stage_order"` // Порядковый номер этапа
	Name       string           `json:"name"`        // Название этапа
	StageType  string           `json:"stage_type"`  // Тип этапа
	CanDelete  bool             `json:"can_delete"`  // Возможность удаления этапа
	LimitValue int64            `json:"limit_value"` // Лимит времени на этапе
	LimitType  models.LimitType `json:"limit_type"`  // Лимит времени на этапе - единицы измерения
}

type SelectionStageID struct {
	ID string `json:"id"` // Идентификатор этапа подбора кандидата
}

type SelectionStageOrderData struct {
	ID       string `json:"id"`        // Идентификатор этапа подбора кандидата
	NewOrder int    `json:"new_order"` // Новый порядковый номер
}

func SelectionStageConvert(rec dbmodels.SelectionStage) SelectionStageView {
	return SelectionStageView{
		ID:         rec.ID,
		StageOrder: rec.StageOrder,
		Name:       rec.Name,
		StageType:  rec.StageType,
		CanDelete:  rec.CanDelete,
		LimitValue: rec.LimitValue,
		LimitType:  rec.LimitType,
	}
}
