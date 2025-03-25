package externalservices

import (
	"context"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
)

type JobSiteProvider interface {
	GetConnectUri(spaceID string) (uri string, err error)
	RequestToken(spaceID, code string)
	CheckConnected(ctx context.Context, spaceID string) bool
	VacancyPublish(ctx context.Context, spaceID, vacancyID string) (hMsg string, err error)
	VacancyUpdate(ctx context.Context, spaceID, vacancyID string) (hMsg string, err error)
	VacancyClose(ctx context.Context, spaceID, vacancyID string) (hMsg string, err error)
	VacancyAttach(ctx context.Context, spaceID, vacancyID string, extID string) (hMsg string, err error)
	GetVacancyInfo(ctx context.Context, spaceID, vacancyID string) (*vacancyapimodels.ExtVacancyInfo, error)
	SendMessage(ctx context.Context, data dbmodels.Applicant, msg string) error
	GetMessages(ctx context.Context, user dbmodels.SpaceUser, data dbmodels.Applicant) ([]negotiationapimodels.MessageItem, error)
	GetLastInMessage(ctx context.Context, data dbmodels.Applicant) (*negotiationapimodels.MessageItem, error)
}
