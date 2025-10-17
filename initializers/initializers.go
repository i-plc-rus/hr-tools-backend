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
	"hr-tools-backend/lib/utils/lock"
	vacancyhandler "hr-tools-backend/lib/vacancy"
	vacancyreqhandler "hr-tools-backend/lib/vacancy-req"
	"hr-tools-backend/lib/vk"
	vkstep0runworker "hr-tools-backend/lib/vk/step0-run-worker"
	vkstep1runworker "hr-tools-backend/lib/vk/step1-run-worker"
	vkstep10runworker "hr-tools-backend/lib/vk/step10-run-worker"
	vkstep11runworker "hr-tools-backend/lib/vk/step11-run-worker"
	vkstep9doneworker "hr-tools-backend/lib/vk/step9-done-worker"
	vkstep9runworker "hr-tools-backend/lib/vk/step9-run-worker"
	vkstep9scoreworker "hr-tools-backend/lib/vk/step9-score-worker"
	connectionhub "hr-tools-backend/lib/ws/hub/connection-hub"
	"time"
)

var LoggerConfig *fiberlog.Config

func InitAllServices(ctx context.Context) {
	LoggerConfig = InitLogger()
	config.InitConfig()
	InitDBConnection()
	InitS3()
	InitSmtp()
	connectionhub.Init()
	lock.InitResourceLock(ctx)
	filestorage.NewHandler()
	cityprovider.NewHandler()
	pushhandler.NewHandler()
	hhclient.NewProvider(config.Conf.HH.RedirectUri)
	avitoclient.NewProvider()
	applicanthistoryhandler.NewHandler()
	spaceusershander.NewHandler()
	spacehandler.NewHandler(config.Conf.Sales.Email)
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
	gpthandler.NewHandler(false)
	hhhandler.NewHandler()
	avitohandler.NewHandler()
	applicant.NewHandler()
	messagetemplate.NewHandler()
	xlsexport.NewHandler()
	analytics.NewHandler()
	negotiationchathandler.NewHandler()
	survey.NewHandler()
	vk.NewHandler(ctx)
	supersethandler.NewHandler(config.Conf.Superset.Host, config.Conf.Superset.Username, config.Conf.Superset.Password, config.Conf.Superset.DashboardParams)
	go initWorkers(ctx)
}

// запускаем с промежутком в 10 сек чтоб размыть нагрузку
func initWorkers(ctx context.Context) {
	//Задача проверки статусов модерации/публикации в HH/Avito
	externalserviceworker.StartWorker(ctx)

	// Задача  ВК. Шаг 0. отправка ссылки на анкету с типовыми вопросами
	vkstep0runworker.StartWorker(ctx)

	// Задача ВК. Шаг 1. Генерация черновика скрипта
	vkstep1runworker.StartWorker(ctx)

	// Задача ВК. Шаг 9. Транскрибация видео ответов
	vkstep9runworker.StartWorker(ctx)

	// Задача ВК. Шаг 9. Cемантическая оценка ответов
	vkstep9scoreworker.StartWorker(ctx)

	// Задача ВК. Шаг 10. Подсчёт баллов и адаптивный фильтр
	vkstep10runworker.StartWorker(ctx)

	// Задача ВК. Шаг 9. семантическая оценка для опроса завершена
	vkstep9doneworker.StartWorker(ctx)

	// Задача ВК. Шаг 11. Генерация отчёта и рекомендаций
	vkstep11runworker.StartWorker(ctx)

	if makeTimeGap(ctx) {
		//Задача получения откликов по вакансиям из HH/Avito
		negotiationworker.StartWorker(ctx)
	}
	if makeTimeGap(ctx) {
		// Задача получения сообщений из HH/Avito от кандидатов
		newmsgworker.StartWorker(ctx)
	}
	// Deprecated: используются vkstep
	/*
		if makeTimeGap(ctx) {
			// Задача генерации опросов для кандидатов отправивших отклик
			applicantsurveyworker.StartWorker(ctx)
		}
		if makeTimeGap(ctx) {
			// Задача отправки ссылок на опрос кандидатам
			applicantsurveysuggestworker.StartWorker(ctx)
		}
		if makeTimeGap(ctx) {
			// Задача оценки кандидатов
			applicantsurveyscoreworker.StartWorker(ctx)
		}
	*/
}

func makeTimeGap(ctx context.Context) (canRun bool) {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(time.Second * 10):
		return true
	}
}
