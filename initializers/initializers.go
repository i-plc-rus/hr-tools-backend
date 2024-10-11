package initializers

import (
	"context"
	"hr-tools-backend/config"
	"hr-tools-backend/fiberlog"
	adminpanelhandler "hr-tools-backend/lib/admin-panel"
	adminpanelauthhandler "hr-tools-backend/lib/admin-panel/auth"
	companyprovider "hr-tools-backend/lib/dicts/company"
	departmentprovider "hr-tools-backend/lib/dicts/department"
	jobtitleprovider "hr-tools-backend/lib/dicts/job-title"
	spaceauthhandler "hr-tools-backend/lib/space/auth"
	spacehandler "hr-tools-backend/lib/space/handler"
)

var LoggerConfig *fiberlog.Config

func InitAllServices(ctx context.Context) {
	LoggerConfig = InitLogger()
	config.InitConfig()
	InitDBConnection()
	InitSmtp()
	spacehandler.NewHandler()
	spaceauthhandler.NewHandler()
	adminpanelauthhandler.NewHandler()
	adminpanelhandler.NewHandler()
	companyprovider.NewHandler()
	departmentprovider.NewHandler()
	jobtitleprovider.NewHandler()
}
