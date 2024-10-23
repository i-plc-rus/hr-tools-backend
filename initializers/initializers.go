package initializers

import (
	"context"
	"hr-tools-backend/config"
	"hr-tools-backend/fiberlog"
	adminpanelhandler "hr-tools-backend/lib/admin-panel"
	adminpanelauthhandler "hr-tools-backend/lib/admin-panel/auth"
	aprovalstageshandler "hr-tools-backend/lib/aproval-stages"
	companyprovider "hr-tools-backend/lib/dicts/company"
	companystructprovider "hr-tools-backend/lib/dicts/company-struct"
	departmentprovider "hr-tools-backend/lib/dicts/department"
	jobtitleprovider "hr-tools-backend/lib/dicts/job-title"
	spaceauthhandler "hr-tools-backend/lib/space/auth"
	spacehandler "hr-tools-backend/lib/space/handler"
	spacesettingshandler "hr-tools-backend/lib/space/settings/handler"
	spaceusershander "hr-tools-backend/lib/space/users/hander"
	vacancyhandler "hr-tools-backend/lib/vacancy"
	vacancyreqhandler "hr-tools-backend/lib/vacancy-req"
)

var LoggerConfig *fiberlog.Config

func InitAllServices(ctx context.Context) {
	LoggerConfig = InitLogger()
	config.InitConfig()
	InitDBConnection()
	InitSmtp()
	spaceusershander.NewHandler()
	spacehandler.NewHandler()
	spaceauthhandler.NewHandler()
	adminpanelauthhandler.NewHandler()
	adminpanelhandler.NewHandler()
	companyprovider.NewHandler()
	departmentprovider.NewHandler()
	jobtitleprovider.NewHandler()
	companystructprovider.NewHandler()
	aprovalstageshandler.NewHandler()
	vacancyhandler.NewHandler()
	vacancyreqhandler.NewHandler()
	spacesettingshandler.NewHandler()
}
