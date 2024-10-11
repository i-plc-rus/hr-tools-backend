package initializers

import (
	"hr-tools-backend/config"
	"hr-tools-backend/db"
)

func InitDBConnection() {
	err := db.Connect(config.Conf.Database.Host, config.Conf.Database.Port, config.Conf.Database.Name,
		config.Conf.Database.User, config.Conf.Database.Password, *config.Conf.Database.DebugMode, *config.Conf.Database.MigrateOnStart)
	if err != nil {
		panic(err.Error())
	}

	db.InitPreload()
}
