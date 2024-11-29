package applicanthistorystore

import (
	applicantapimodels "hr-tools-backend/models/api/applicant"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Provider interface {
	Create(rec dbmodels.ApplicantHistory) (id string, err error)
	ListCount(spaceID, userID string, filter applicantapimodels.ApplicantHistoryFilter) (count int64, err error)
	List(spaceID, applicantID string, filter applicantapimodels.ApplicantHistoryFilter) (list []dbmodels.ApplicantHistory, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.ApplicantHistory) (id string, err error) {
	err = i.db.
		Omit("Vacancy").
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) ListCount(spaceID, applicantID string, filter applicantapimodels.ApplicantHistoryFilter) (count int64, err error) {
	var rowCount int64
	tx := i.db.
		Model(dbmodels.ApplicantHistory{}).
		Where("space_id = ?", spaceID).
		Where("applicant_id = ?", applicantID)
	if filter.CommentsOnly {
		tx = tx.Where("action_type = ?", dbmodels.HistoryTypeComment)
	}
	err = tx.Count(&rowCount).Error
	if err != nil {
		log.WithError(err).Error("ошибка получения общего количества действий по профилю кандидата")
		return 0, errors.New("ошибка получения общего количества действий по профилю кандидата")
	}
	return rowCount, nil
}

func (i impl) List(spaceID, applicantID string, filter applicantapimodels.ApplicantHistoryFilter) (list []dbmodels.ApplicantHistory, err error) {
	list = []dbmodels.ApplicantHistory{}
	tx := i.db.
		Model(dbmodels.ApplicantHistory{}).
		Where("space_id = ?", spaceID).
		Where("applicant_id = ?", applicantID)
	if filter.CommentsOnly {
		tx = tx.Where("action_type = ?", dbmodels.HistoryTypeComment)
	}
	page, limit := filter.GetPage()
	i.setPage(tx, page, limit)
	tx.Order("created_at")
	err = tx.Preload("Vacancy").Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) setPage(tx *gorm.DB, page, limit int) {
	offset := (page - 1) * limit
	tx.Limit(limit).Offset(offset)
}
