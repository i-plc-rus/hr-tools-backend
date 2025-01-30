package externalservices

import (
	"context"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
)

type JobSiteProvider interface {
	GetConnectUri(spaceID string) (uri string, err error)
	RequestToken(spaceID, code string)
	CheckConnected(spaceID string) bool
	VacancyPublish(ctx context.Context, spaceID, vacancyID string) (hMsg string, err error)
	VacancyUpdate(ctx context.Context, spaceID, vacancyID string) (hMsg string, err error)
	VacancyClose(ctx context.Context, spaceID, vacancyID string) (hMsg string, err error)
	VacancyAttach(ctx context.Context, spaceID, vacancyID string, extID string) (hMsg string, err error)
	GetVacancyInfo(ctx context.Context, spaceID, vacancyID string) (*vacancyapimodels.ExtVacancyInfo, error)
}
