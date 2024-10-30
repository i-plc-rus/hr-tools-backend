package externalservices

import (
	"context"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
)

type JobSiteProvider interface {
	GetConnectUri(spaceID string) (uri string, err error)
	RequestToken(spaceID, code string)
	CheckConnected(spaceID string) bool
	VacancyPublish(ctx context.Context, spaceID, vacancyID string) (err error)
	VacancyUpdate(ctx context.Context, spaceID, vacancyID string) error
	VacancyClose(ctx context.Context, spaceID, vacancyID string) error
	VacancyAttach(ctx context.Context, spaceID, vacancyID string, extID string) error
	GetVacancyInfo(ctx context.Context, spaceID, vacancyID string) (*vacancyapimodels.ExtVacancyInfo, error)
}
