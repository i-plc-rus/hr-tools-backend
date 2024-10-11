package db

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	dbmodels "hr-tools-backend/models/db"
)

func AutoMigrateDB() error {
	DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
	log.Info("Запуск миграций")
	if err := DB.AutoMigrate(&dbmodels.Space{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры Space")
	}
	if err := DB.AutoMigrate(&dbmodels.EmailVerify{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры EmailVerify")
	}
	if err := DB.AutoMigrate(&dbmodels.SpaceUser{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры EmailVerify")
	}

	if err := DB.AutoMigrate(&dbmodels.AdminPanelUser{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры AdminPanelUser")
	}
	log.Info("Миграция прошла успешно")
	return nil
}
