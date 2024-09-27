package db

import (
	"fmt"

	gorm_logrus "github.com/onrik/gorm-logrus"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(host string, port string, database string, user string, pass string, debugMode bool, migrate bool) (err error) {
	if DB == nil {
		dbConnString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", host, port, user, database, pass)
		db, err := gorm.Open(postgres.Open(dbConnString), &gorm.Config{
			Logger: gorm_logrus.New(),
		})

		if debugMode {
			db.Logger = logger.Default.LogMode(logger.Info)
		}
		if err != nil {
			return errors.Wrap(err, "Ошибка подключения к БД")
		}
		if debugMode {
			DB = db.Debug()
		} else {
			DB = db
		}
		if migrate {
			err = AutoMigrateDB()
		}
		log.Info("Сервис успешно подключен к БД")
	}
	return err
}

func PingDB() error {
	db, err := DB.DB()
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}
