package config

import (
	"hr-tools-backend/models"
	"strings"

	"github.com/gotify/configor"
)

var Conf *Configuration

type Configuration struct {
	App struct {
		ListenAddr string `default:"" env:"APP_HOST"`
		Port       int    `default:"8080"  env:"APP_PORT"`
	}
	Database struct {
		Host           string `default:"127.0.0.1" env:"DB_HOST"`
		Port           string `default:"5432" env:"DB_PORT"`
		Name           string `default:"hr-tools" env:"DB_NAME"`
		User           string `default:"postgres" env:"DB_USER"`
		Password       string `default:"postgres" env:"DB_PASSWORD"`
		MigrateOnStart *bool  `default:"true" env:"DB_MIGRATE_ON_START"`
		DebugMode      *bool  `default:"false" env:"DB_DEBUG_MODE"`
	}
	Smtp struct {
		User                  string `default:"" env:"SMTP_USER"`
		Password              string `default:"" env:"SMTP_PASSWORD"`
		Host                  string `default:"" env:"SMTP_HOST"`
		Port                  string `default:"" env:"SMTP_PORT"`
		TLSEnabled            *bool  `default:"true" env:"SMTP_TLS_ENABLED"`
		EmailSendVerification string `default:"" env:"EMAIL_SEND_VERIFICATION"`
		DomainForVerifyLink   string `default:"http://localhost:8000" env:"DOMAIN_FOR_VERIFY_LINK"`
		ResetUI               string `default:"https://s.hr-tools.pro/auth/password-recovery-reset" env:"PASSWORD_RESET_UI_LINK"`
	}
	DaData struct {
		ApiKey  string `default:"" env:"DADATA_API_KEY"`
		Timeout int64  `default:"20" env:"DADATA_TIMEOUT_SEC"`
	}
	Auth struct {
		JWTExpireInSec        int64  `default:"2678400" env:"JWT_EXPIRE"`
		JWTRefreshExpireInSec int64  `default:"5356800" env:"JWT_REFRESH_EXPIRE"`
		JWTSecret             string `default:"secret-key-123" env:"JWT_SECRET"`
	}
	AdminPanelAuth struct {
		JWTExpireInSec        int64  `default:"2678400" env:"ADMIN_PANEL_JWT_EXPIRE"`
		JWTRefreshExpireInSec int64  `default:"5356800" env:"ADMIN_PANEL_JWT_REFRESH_EXPIRE"`
		JWTSecret             string `default:"secret-key-321" env:"ADMIN_PANEL_JWT_SECRET"`
	}
	Admin struct {
		FirstName   string `default:"Admin" env:"SUPER_ADMIN_FIRST_NAME"`
		LastName    string `default:"Admin" env:"SUPER_ADMIN_LAST_NAME"`
		Email       string `default:"admin@admin.ad" env:"SUPER_ADMIN_EMAIL"`
		PhoneNumber string `default:"" env:"SUPER_ADMIN_PHONE"`
		Password    string `default:"123hygAS" env:"SUPER_ADMIN_PASSWORD"`
	}
	AI struct {
		VkStep1AI string `default:"Ollama" env:"AI_VK_STEP1"` //Ollama | YandexGPT
		YandexGPT struct {
			IAMToken  string `default:"" env:"YANDEXGPT_IAM_TOKEN"`
			CatalogID string `default:"" env:"YANDEXGPT_CATALOG_ID"`
		}
		Ollama struct {
			OllamaURL   string `default:"http://localhost:11434/api/generate" env:"OLLAMA_URL"`
			OllamaModel string `default:"deepseek-r1:8b" env:"OLLAMA_MODEL"` //deepseek-r1:8b/llama3 //https://ollama.com/search
		}
	}
	HH struct {
		RedirectUri string `default:"https://a.hr-tools.pro/api/v1/oauth/callback/hh" env:"HH_REDIRECT"`
	}
	S3 struct {
		Endpoint         string `default:"minio" env:"S3_ENDPOINT"`
		AccessKeyID      string `default:"" env:"S3_ACCESS_KEY_ID"`
		SecretAccessKey  string `default:"" env:"S3_SECRET_KEY"`
		UseSSL           *bool  `default:"false" env:"S3_USE_SSL"`
		BucketNamePrefix string `default:"hr-tools" env:"S3_BUCKET_NAME_PREFIX"`
	}
	Recovery struct {
		MailTitle string `default:"Восстановление пароля" env:"RECOVERY_MAIL_TITLE"`
		MailBody  string `default:"Здравствуйте,<br>Вы запросили сброс пароля вашей учетной записи.<br>Пожалуйста, нажмите кнопку ниже, чтобы создать новый пароль. Если вы не хотели сбрасывать свой пароль, просто проигнорируйте это письмо.<br>[link]<br>Обратите внимание, что эту ссылку можно использовать только один раз. Если вы отправили более 1 запроса на сброс пароля, используйте последнюю полученную вами ссылку." env:"RECOVERY_MAIL_BODY"`
	}
	Superset struct {
		Host          string `default:"https://superset.hr-tools.pro" env:"SUPERSET_HOST"`
		Username      string `default:"admin" env:"SUPERSET_USERNAME"`
		Password      string `default:"P@SSW0RD" env:"SUPERSET_PASSWORD"`
		ResourcesType string `default:"dashboard" env:"SUPERSET_RESOURCES_TYPE"`
		//список дашбордов ';' Формат эллемента списка: Code:DashboardID; Пример: recruiter_funnel:9df52fe6-4e65-43b2-a58c-a717804fd913;cohort_funnel:d7c63a3f-897d-402d-ac9c-ea06254c3238;
		Dashboards      string `default:"" env:"SUPERSET_DASHBOARDS"`
		DashboardParams models.DashboardParams
	}
	Sales struct {
		Email string `default:"info@it-tech.group" env:"SALES_EMAIL"` // sales@hr.tools.pro - previous one
	}
	UIParams struct {
		SurveyStep0Path string `default:"https://s.hr-tools.pro/public/survey/step0/" env:"PUBLIC_SURVEY_STEP0_UI_URL"`
		SurveyPath      string `default:"https://s.hr-tools.pro/public/survey/" env:"PUBLIC_SURVEY_UI_URL"`
	}
}

func configFiles() []string {
	return []string{"config.yml"}
}

func InitConfig() {
	if Conf != nil {
		return
	}
	conf := new(Configuration)
	err := configor.New(&configor.Config{}).Load(conf, configFiles()...)
	if err != nil {
		panic(err)
	}
	conf.Superset.DashboardParams = models.DashboardParams{}
	if conf.Superset.Dashboards != "" {
		rules := strings.Split(conf.Superset.Dashboards, ";")
		for _, rule := range rules {
			values := strings.Split(rule, ":")
			if len(values) != 2 {
				continue
			}
			r := models.DashboardParam{}
			for i, value := range values {
				switch i {
				case 0:
					r.Code = value
					break
				case 1:
					r.DashboardID = value
					break
				}
			}
			conf.Superset.DashboardParams = append(conf.Superset.DashboardParams, r)
		}
	}
	Conf = conf
}
