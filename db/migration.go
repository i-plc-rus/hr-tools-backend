package db

import (
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
		return errors.Wrap(err, "ошибка создания структуры SpaceUser")
	}

	// Миграция: проставить существующим пользователям status=WORKING и status_changed_at=NOW()
	if err := DB.Exec(`
		UPDATE space_users 
		SET status = 'WORKING', status_changed_at = NOW() 
		WHERE status IS NULL OR status = '' OR status_changed_at = '0001-01-01 00:00:00'::timestamp
	`).Error; err != nil {
		return errors.Wrap(err, "ошибка обновления статусов пользователей")
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

	if err := DB.AutoMigrate(&dbmodels.VacancyComment{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры VacancyComment")
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

	if err := DB.AutoMigrate(&dbmodels.AiLog{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры AiLog")
	}

	if err := DB.AutoMigrate(&dbmodels.VacancyRequestComment{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры VacancyRequestComment")
	}

	if err := DB.AutoMigrate(&dbmodels.ApplicantVkStep{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры ApplicantVkStep")
	}

	if err := DB.AutoMigrate(&dbmodels.QuestionHistory{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры QuestionHistory")
	}

	if err := DB.AutoMigrate(&dbmodels.MasaiSession{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры MasaiSession")
	}

	if err := DB.AutoMigrate(&dbmodels.ApplicantVkVideoSurvey{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры ApplicantVkVideoSurvey")
	}

	if err := DB.AutoMigrate(&dbmodels.LanguageData{}); err != nil {
		return errors.Wrap(err, "ошибка создания структуры LanguageData")
	}

	log.Info("Миграция прошла успешно")
	return nil
}
