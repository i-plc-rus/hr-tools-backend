package config

import (
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
	Conf = conf
}
