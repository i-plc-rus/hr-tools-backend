package applicanthistoryhandler

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	applicanthistorystore "hr-tools-backend/lib/applicant-history/store"
	applicantstore "hr-tools-backend/lib/applicant/store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	applicantapimodels "hr-tools-backend/models/api/applicant"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	List(spaceID, applicantID string, filter applicantapimodels.ApplicantHistoryFilter) ([]applicantapimodels.ApplicantHistoryView, int64, error)
	Save(spaceID, applicantID, vacancyID, userID string, action dbmodels.ActionType, changes dbmodels.ApplicantChanges)
	SaveNote(spaceID, applicantID, userID string, action applicantapimodels.ApplicantNote) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store:          applicanthistorystore.NewInstance(db.DB),
		userStore:      spaceusersstore.NewInstance(db.DB),
		applicantStore: applicantstore.NewInstance(db.DB),
	}
}

type impl struct {
	store          applicanthistorystore.Provider
	userStore      spaceusersstore.Provider
	applicantStore applicantstore.Provider
}

func (i impl) List(spaceID, applicantID string, filter applicantapimodels.ApplicantHistoryFilter) ([]applicantapimodels.ApplicantHistoryView, int64, error) {

	rowCount, err := i.store.ListCount(spaceID, applicantID, filter)
	if err != nil {
		return nil, 0, err
	}

	page, limit := filter.GetPage()
	offset := (page - 1) * limit
	if int64(offset) > rowCount {
		return []applicantapimodels.ApplicantHistoryView{}, rowCount, nil
	}

	list, err := i.store.List(spaceID, applicantID, filter)
	if err != nil {
		log.WithError(err).Error("ошибка получения списка действий")
		return nil, 0, errors.New("ошибка получения списка действий")
	}
	result := make([]applicantapimodels.ApplicantHistoryView, 0, len(list))
	for _, rec := range list {
		result = append(result, applicantapimodels.Convert(rec))
	}
	return result, rowCount, nil
}

func (i impl) Save(spaceID, applicantID, vacancyID, userID string, action dbmodels.ActionType, changes dbmodels.ApplicantChanges) {
	logger := log.WithField("space_id", spaceID).
		WithField("applicant_id", applicantID).
		WithField("vacancy_id", vacancyID).
		WithField("action", action).
		WithField("description", changes.Description)
	rec := dbmodels.ApplicantHistory{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		ApplicantID: applicantID,
		VacancyID:   vacancyID,
		ActionType:  action,
		Changes:     changes,
	}
	if userID != "" {
		rec.UserID = &userID
		user, err := i.userStore.GetByID(userID)
		if err != nil {
			logger.WithError(err).Error("ошибка сохранения истории действий по кандидату, не удалось получить автора изменений")
			return
		}
		if user == nil {
			logger.Error("ошибка сохранения истории действий по кандидату, автор изменений не найден")
			return
		}
		rec.UserName = user.GetFullName()
	} else {
		rec.UserName = "Система"
	}
	_, err := i.store.Create(rec)
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения истории действий по кандидату")
	}
}

func (i impl) SaveNote(spaceID, applicantID, userID string, note applicantapimodels.ApplicantNote) error {
	logger := log.WithField("space_id", spaceID).
		WithField("applicant_id", applicantID).
		WithField("action", dbmodels.HistoryTypeComment).
		WithField("description", note.Note)

	applicantRec, err := i.applicantStore.GetByID(spaceID, applicantID)
	if err != nil {
		logger.WithError(err).Error("ошибка получения кандидата")
		return errors.New("ошибка получения кандидата")
	}
	if applicantRec == nil {
		return errors.New("кандидат не найден")
	}
	logger = logger.WithField("vacancy_id", applicantRec.VacancyID)
	user, err := i.userStore.GetByID(userID)
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения комментария по кандидату, не удалось получить автора изменений")
		return errors.New("ошибка сохранения комментария по кандидату, не удалось получить автора изменений")
	}
	if user == nil {
		logger.Error("ошибка сохранения комментария по кандидату, автор изменений не найден")
		return errors.New("ошибка сохранения комментария по кандидату, автор изменений не найден")
	}

	rec := dbmodels.ApplicantHistory{
		BaseSpaceModel: dbmodels.BaseSpaceModel{SpaceID: spaceID},
		ApplicantID:    applicantID,
		VacancyID:      applicantRec.VacancyID,
		UserID:         &userID,
		UserName:       user.GetFullName(),
		ActionType:     dbmodels.HistoryTypeComment,
		Changes:        dbmodels.ApplicantChanges{Description: note.Note},
	}
	_, err = i.store.Create(rec)
	if err != nil {
		logger.WithError(err).Error("ошибка сохранения комментария по кандидату")
		return errors.New("ошибка сохранения комментария по кандидату")
	}
	return nil
}
