package negotiationchathandler

import (
	"context"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	externalservices "hr-tools-backend/lib/external-services"
	avitohandler "hr-tools-backend/lib/external-services/avito"
	hhhandler "hr-tools-backend/lib/external-services/hh"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	"hr-tools-backend/models"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"

	"github.com/pkg/errors"
)

type Provider interface {
	IsVailable(spaceID, applicantID string) (negotiationapimodels.MessengerAvailableResponse, error)
	SendMessage(spaceID string, req negotiationapimodels.NewMessageRequest) error
	GetMessages(spaceID, userID string, req negotiationapimodels.MessageListRequest) (list []negotiationapimodels.MessageItem, err error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		applicantStore: applicantstore.NewInstance(db.DB),
		userStore:      spaceusersstore.NewInstance(db.DB),
	}
}

type impl struct {
	applicantStore applicantstore.Provider
	userStore      spaceusersstore.Provider
}

func (i impl) IsVailable(spaceID, applicantID string) (negotiationapimodels.MessengerAvailableResponse, error) {
	applicant, err := i.applicantStore.GetByID(spaceID, applicantID)
	if err != nil {
		return negotiationapimodels.MessengerAvailableResponse{}, err
	}
	if applicant == nil {
		return negotiationapimodels.MessengerAvailableResponse{}, errors.New("кандидат не найден")
	}
	if (applicant.Source != models.ApplicantSourceAvito && applicant.Source != models.ApplicantSourceHh) ||
		applicant.NegotiationID == "" {
		return negotiationapimodels.MessengerAvailableResponse{IsAvailable: false}, nil
	}

	result := negotiationapimodels.MessengerAvailableResponse{IsAvailable: false, Service: string(applicant.Source)}
	switch applicant.Source {
	case models.ApplicantSourceAvito:
		//для авито должен быть идентификатор чата
		if applicant.ChatID == "" {
			return result, nil
		}
		result.IsAvailable = avitohandler.Instance.CheckConnected(context.TODO(), spaceID)
	case models.ApplicantSourceHh:
		result.IsAvailable = hhhandler.Instance.CheckConnected(context.TODO(), spaceID)
	}
	return result, nil
}

func (i impl) SendMessage(spaceID string, req negotiationapimodels.NewMessageRequest) error {
	applicant, err := i.applicantStore.GetByID(spaceID, req.ApplicantID)
	if err != nil {
		return err
	}
	if applicant == nil {
		return errors.New("кандидат не найден")
	}
	var handler externalservices.JobSiteProvider
	switch applicant.Source {
	case models.ApplicantSourceAvito:
		handler = avitohandler.Instance
	case models.ApplicantSourceHh:
		handler = hhhandler.Instance
	default:
		return errors.New("чат не поддерживается")
	}
	return handler.SendMessage(context.TODO(), applicant.Applicant, req.Text)
}

func (i impl) GetMessages(spaceID, userID string, req negotiationapimodels.MessageListRequest) (list []negotiationapimodels.MessageItem, err error) {
	applicant, err := i.applicantStore.GetByID(spaceID, req.ApplicantID)
	if err != nil {
		return nil, err
	}
	if applicant == nil {
		return nil, errors.New("кандидат не найден")
	}

	user, err := i.userStore.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("пользователь не найден")
	}
	var handler externalservices.JobSiteProvider
	switch applicant.Source {
	case models.ApplicantSourceAvito:
		handler = avitohandler.Instance
	case models.ApplicantSourceHh:
		handler = hhhandler.Instance
	default:
		return nil, errors.New("чат не поддерживается")
	}
	return handler.GetMessages(context.TODO(), *user, applicant.Applicant)
}
