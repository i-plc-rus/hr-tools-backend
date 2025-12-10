package vacancyapimodels

import (
	"fmt"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type ApprovalTasks struct {
	ApprovalTasks []ApprovalTaskData `json:"approval_stages"`
}

func (v ApprovalTasks) Validate() error {
	for _, item := range v.ApprovalTasks {
		err := item.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

type ApprovalTaskData struct {
	AssigneeUserID string `json:"assignee_user_id"`
}

func (a ApprovalTaskData) Validate() error {
	if a.AssigneeUserID == "" {
		return errors.New("отсутсвует идентификатор пользователя")
	}
	return nil
}

type ApprovalRequestChanges struct {
	Comment string `json:"comment"`
}

func (v ApprovalRequestChanges) Validate() error {
	if v.Comment == "" {
		return errors.New("отсутсвует комментарий")
	}
	return nil
}

type ApprovalReject struct {
	Comment string `json:"comment"`
}

func (v ApprovalReject) Validate() error {
	if v.Comment == "" {
		return errors.New("отсутсвует комментарий")
	}
	return nil
}

type ApprovalTaskView struct {
	ApprovalTaskData
	ID               string               `json:"id"`
	AssigneeUserName string               `json:"assignee_user_name"`
	State            models.ApprovalState `json:"state"`
	Comment          string               `json:"comment"`
	DecidedAt        *time.Time           `json:"decided_at"`
}

func ApprovalStageConvert(rec dbmodels.ApprovalTask) ApprovalTaskView {
	userName := ""
	if rec.AssigneeUser != nil {
		userName = strings.TrimSpace(fmt.Sprintf("%v %v", rec.AssigneeUser.FirstName, rec.AssigneeUser.LastName))
	}
	return ApprovalTaskView{
		ApprovalTaskData: ApprovalTaskData{
			AssigneeUserID: rec.AssigneeUserID,
		},
		ID:               rec.ID,
		AssigneeUserName: userName,
		State:            rec.State,
		Comment:          rec.Comment,
		DecidedAt:        rec.DecidedAt,
	}
}

type ApprovalHistoryData struct {
	SpaceID        string               `json:"space_id"`
	RequestID      string               `json:"request_id"`
	TaskID         string               `json:"task_id"`
	AssigneeUserID string               `json:"assignee_user_id"`
	State          models.ApprovalState `json:"state"`
	Comment        string               `json:"comment"`
}

type ApprovalHistoryView struct {
	ApprovalHistoryData
	CreatedAt        time.Time              `json:"created_at"`
	AssigneeUserName string                 `json:"assignee_user_name"`
	Changes          dbmodels.EntityChanges `json:"changes"` // Изменения
}

func ApprovalHistoryConvert(rec dbmodels.ApprovalHistory) ApprovalHistoryView {
	userName := ""
	if rec.AssigneeUser != nil {
		userName = strings.TrimSpace(fmt.Sprintf("%v %v", rec.AssigneeUser.FirstName, rec.AssigneeUser.LastName))
	}
	return ApprovalHistoryView{
		ApprovalHistoryData: ApprovalHistoryData{
			RequestID:      rec.RequestID,
			TaskID:         rec.TaskID,
			AssigneeUserID: rec.AssigneeUserID,
			State:          rec.State,
			Comment:        rec.Comment,
			SpaceID:        rec.SpaceID,
		},
		CreatedAt:        rec.CreatedAt,
		AssigneeUserName: userName,
		Changes:          rec.Changes,
	}
}
