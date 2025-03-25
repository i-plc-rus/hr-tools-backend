package initializers

import (
	"context"
	"hr-tools-backend/config"
	"hr-tools-backend/fiberlog"
	adminpanelhandler "hr-tools-backend/lib/admin-panel"
	adminpanelauthhandler "hr-tools-backend/lib/admin-panel/auth"
	"hr-tools-backend/lib/analytics"
	"hr-tools-backend/lib/applicant"
	applicanthistoryhandler "hr-tools-backend/lib/applicant-history"
	aprovalstageshandler "hr-tools-backend/lib/aproval-stages"
	cityprovider "hr-tools-backend/lib/dicts/city"
	companyprovider "hr-tools-backend/lib/dicts/company"
	companystructprovider "hr-tools-backend/lib/dicts/company-struct"
	departmentprovider "hr-tools-backend/lib/dicts/department"
	jobtitleprovider "hr-tools-backend/lib/dicts/job-title"
	rejectreasonprovider "hr-tools-backend/lib/dicts/reject-reason"
	xlsexport "hr-tools-backend/lib/export/xls"
	avitohandler "hr-tools-backend/lib/external-services/avito"
	avitoclient "hr-tools-backend/lib/external-services/avito/client"
	hhhandler "hr-tools-backend/lib/external-services/hh"
	"hr-tools-backend/lib/external-services/hh/hhclient"
	negotiationchathandler "hr-tools-backend/lib/external-services/negotiation-chat"
	newmsgworker "hr-tools-backend/lib/external-services/negotiation-chat/new-msg-worker"
	negotiationworker "hr-tools-backend/lib/external-services/negotiation-worker"
	externalserviceworker "hr-tools-backend/lib/external-services/worker"
	filestorage "hr-tools-backend/lib/file-storage"
	gpthandler "hr-tools-backend/lib/gpt"
	messagetemplate "hr-tools-backend/lib/message-template"
	spaceauthhandler "hr-tools-backend/lib/space/auth"
	spacehandler "hr-tools-backend/lib/space/handler"
	pushhandler "hr-tools-backend/lib/space/push/handler"
	spacesettingshandler "hr-tools-backend/lib/space/settings/handler"
	spaceusershander "hr-tools-backend/lib/space/users/hander"
	supersethandler "hr-tools-backend/lib/superset"
	"hr-tools-backend/lib/survey"
	applicantsurveyscoreworker "hr-tools-backend/lib/survey/applicant-survey-score-worker"
	applicantsurveyworker "hr-tools-backend/lib/survey/applicant-survey-worker"
	vacancyhandler "hr-tools-backend/lib/vacancy"
	vacancyreqhandler "hr-tools-backend/lib/vacancy-req"
	connectionhub "hr-tools-backend/lib/ws/hub/connection-hub"
)

var LoggerConfig *fiberlog.Config

func InitAllServices(ctx context.Context) {
	LoggerConfig = InitLogger()
	config.InitConfig()
	InitDBConnection()
	InitS3()
	InitSmtp()
	connectionhub.Init()
	filestorage.NewHandler()
	cityprovider.NewHandler()
	pushhandler.NewHandler()
	hhclient.NewProvider(config.Conf.HH.RedirectUri)
	avitoclient.NewProvider()
	applicanthistoryhandler.NewHandler()
	spaceusershander.NewHandler()
	spacehandler.NewHandler()
	spaceauthhandler.NewHandler()
	adminpanelauthhandler.NewHandler()
	adminpanelhandler.NewHandler()
	companyprovider.NewHandler()
	departmentprovider.NewHandler()
	jobtitleprovider.NewHandler()
	rejectreasonprovider.NewHandler()
	companystructprovider.NewHandler()
	aprovalstageshandler.NewHandler()
	vacancyhandler.NewHandler()
	vacancyreqhandler.NewHandler()
	spacesettingshandler.NewHandler()
	gpthandler.NewHandler()
	hhhandler.NewHandler()
	avitohandler.NewHandler()
	applicant.NewHandler()
	externalserviceworker.StartWorker(ctx)
	negotiationworker.StartWorker(ctx)
	messagetemplate.NewHandler()
	xlsexport.NewHandler()
	analytics.NewHandler()
	negotiationchathandler.NewHandler()
	survey.NewHandler()
	newmsgworker.StartWorker(ctx)
	supersethandler.NewHandler(config.Conf.Superset.Host, config.Conf.Superset.Username, config.Conf.Superset.Password, config.Conf.Superset.DashboardParams)
	applicantsurveyworker.StartWorker(ctx)
	applicantsurveyscoreworker.StartWorker(ctx)
}
