package vacancyapimodels

import (
	"fmt"
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"strings"
)

type ApprovalStages struct {
	ApprovalStages []ApprovalStageData `json:"approval_stages"`
}

func (v ApprovalStages) Validate() error {
	for _, item := range v.ApprovalStages {
		err := item.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

type ApprovalStageData struct {
	Stage          int                   `json:"stage"`
	SpaceUserID    string                `json:"space_user_id"`
	ApprovalStatus models.ApprovalStatus `json:"approval_status"`
}

func (a ApprovalStageData) Validate() error {
	if a.SpaceUserID == "" {
		return errors.New("отсутсвует идентификатор пользователя")
	}
	return nil
}

type ApprovalStageView struct {
	ApprovalStageData
	SpaceUserName string `json:"space_user_name"`
}

func ApprovalStageConvert(rec dbmodels.ApprovalStage) ApprovalStageView {
	userName := ""
	if rec.SpaceUser != nil {
		userName = strings.TrimSpace(fmt.Sprintf("%v %v", rec.SpaceUser.FirstName, rec.SpaceUser.LastName))
	}
	return ApprovalStageView{
		ApprovalStageData: ApprovalStageData{
			Stage:          rec.Stage,
			SpaceUserID:    rec.SpaceUserID,
			ApprovalStatus: rec.ApprovalStatus,
		},
		SpaceUserName: userName,
	}
}
