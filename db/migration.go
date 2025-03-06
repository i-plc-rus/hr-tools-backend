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

	if err := DB.AutoMigrate(&dbmodels.Vacancy{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры Vacancy")
	}

	if err := DB.AutoMigrate(&dbmodels.Favorite{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры Favorite")
	}

	if err := DB.AutoMigrate(&dbmodels.Pinned{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры Pinned")
	}

	if err := DB.AutoMigrate(&dbmodels.SpaceSetting{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры SpaceSetting")
	}

	if err := DB.AutoMigrate(&dbmodels.ExtData{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры ExtData")
	}
	if err := DB.AutoMigrate(&dbmodels.Applicant{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры Applicant")
	}
	if err := DB.AutoMigrate(&dbmodels.MessageTemplate{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры MessageTemplate")
	}

	if err := DB.AutoMigrate(&dbmodels.VrFavorite{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры VrFavorite")
	}

	if err := DB.AutoMigrate(&dbmodels.VrPinned{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры VrPinned")
	}

	if err := DB.AutoMigrate(&dbmodels.SelectionStage{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры SelectionStage")
	}

	if err := DB.AutoMigrate(&dbmodels.FileStorage{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры FileStorage")
	}

	if err := DB.AutoMigrate(&dbmodels.ApplicantHistory{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры ApplicantHistory")
	}

	if err := DB.AutoMigrate(&dbmodels.VacancyTeam{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры VacancyTeam")
	}

	if err := DB.AutoMigrate(&dbmodels.RejectReason{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры RejectReason")
	}

	if err := DB.AutoMigrate(&dbmodels.SpacePushSetting{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры SpacePushSetting")
	}

	if err := DB.AutoMigrate(&dbmodels.PushData{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры PushData")
	}

	if err := DB.AutoMigrate(&dbmodels.HRSurvey{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры HRSurvey")
	}

	if err := DB.AutoMigrate(&dbmodels.ApplicantSurvey{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры ApplicantSurvey")
	}

	log.Info("Миграция прошла успешно")
	return nil
}
