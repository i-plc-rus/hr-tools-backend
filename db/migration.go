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

	if err := DB.AutoMigrate(&dbmodels.Company{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры Company")
	}
	if err := DB.AutoMigrate(&dbmodels.Department{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры Department")
	}
	if err := DB.AutoMigrate(&dbmodels.JobTitle{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры JobTitle")
	}

	if err := DB.AutoMigrate(&dbmodels.City{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры City")
	}

	if err := DB.AutoMigrate(&dbmodels.CompanyStruct{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры CompanyStruct")
	}

	if err := DB.AutoMigrate(&dbmodels.ApprovalStage{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры ApprovalStage")
	}

	if err := DB.AutoMigrate(&dbmodels.VacancyRequest{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры VacancyRequest")
	}
	log.Info("Миграция прошла успешно")
	return nil
}
