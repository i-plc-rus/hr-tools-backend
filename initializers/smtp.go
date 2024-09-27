package initializers

import (
	"hr-tools-backend/config"
	"hr-tools-backend/lib/smtp"
)

func InitSmtp() {
	err := smtp.Connect(config.Conf.Smtp.User, config.Conf.Smtp.Password,
		config.Conf.Smtp.Host, config.Conf.Smtp.Port, *config.Conf.Smtp.TLSEnabled)
	if err != nil {
		panic(err.Error())
	}
}
