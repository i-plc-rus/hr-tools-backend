package dictapimodels

import (
	"errors"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
)

type RejectReasonFind struct {
	Search    string                 `json:"search"`    // Поиск по содержимому причины
	Initiator models.RejectInitiator `json:"initiator"` // Фильтр по инициатору отказа
}

type RejectReasonData struct {
	Initiator models.RejectInitiator `json:"initiator"` // Инициатор отказа
	Name      string                 `json:"name"`      // Причина отказа
}

type RejectReasonView struct {
	RejectReasonData
	ID        string `json:"id"`         // Идентификатор записи
	CanChange bool   `json:"can_change"` // Можно изменять
}

func (j RejectReasonData) Validate() error {
	if j.Initiator == "" {
		return errors.New("не указан инициатор отказа")
	}
	err := j.Initiator.IsValid()
	if err != nil {
		return err
	}
	if j.Name == "" {
		return errors.New("не указана причина отказа")
	}
	return nil
}

func RejectReasonConvert(rec dbmodels.RejectReason) RejectReasonView {
	return RejectReasonView{
		RejectReasonData: RejectReasonData{
			Name:      rec.Name,
			Initiator: rec.Initiator,
		},
		ID:        rec.ID,
		CanChange: true,
	}
}
