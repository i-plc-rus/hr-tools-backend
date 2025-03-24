package vacancyreqstore

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
	"strings"
	"time"
)

type Provider interface {
	Create(rec dbmodels.VacancyRequest) (id string, err error)
	GetByID(spaceID, id string) (rec *dbmodels.VacancyRequest, err error)
	Update(spaceID, id string, updMap map[string]interface{}) error
	Delete(spaceID, id string) error
	ListCount(spaceID, userID string, filter vacancyapimodels.VrFilter) (count int64, err error)
	List(spaceID, userID string, filter vacancyapimodels.VrFilter) (list []dbmodels.VacancyRequest, err error)
	SetPin(id, userID string) error
	RemovePin(id, userID string) error
	SetFavorite(id, userID string) error
	RemoveFavorite(id, userID string) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.VacancyRequest) (id string, err error) {
	err = i.db.Omit(clause.Associations).
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) GetByID(spaceID, id string) (*dbmodels.VacancyRequest, error) {
	rec := dbmodels.VacancyRequest{}
	err := i.db.
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
		Preload(clause.Associations).
		Preload("ApprovalStages.SpaceUser").
		First(&rec).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (i impl) Update(spaceID, id string, updMap map[string]interface{}) error {
	if len(updMap) == 0 {
		return nil
	}
	tx := i.db.
		Model(&dbmodels.VacancyRequest{}).
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
		Updates(updMap)
	if tx.RowsAffected == 0 {
		return errors.New("запись не найдена")
	}
	err := tx.Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) Delete(spaceID, id string) error {
	rec := dbmodels.VacancyRequest{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			BaseModel: dbmodels.BaseModel{ID: id},
			SpaceID:   spaceID,
		},
	}
	err := i.db.
		Delete(&rec).
		Error

	if err != nil {
		return err
	}
	return nil
}

func (i impl) ListCount(spaceID, userID string, filter vacancyapimodels.VrFilter) (count int64, err error) {
	var rowCount int64
	tx := i.db.
		Model(dbmodels.VacancyRequest{}).
		Joins("left join vr_favorites as f on vacancy_requests.id = f.vacancy_request_id and f.space_user_id = ?", userID).
		Joins("left join vr_pinneds as p on vacancy_requests.id = p.vacancy_request_id and p.space_user_id = ?", userID).
		Where("space_id = ?", spaceID)
	i.addFilter(tx, filter)
	err = tx.Count(&rowCount).Error
	if err != nil {
		log.WithError(err).Error("ошибка получения общего количества заявок на вакансию")
		return 0, errors.New("ошибка получения общего количества заявок на вакансию")
	}
	return rowCount, nil
}

func (i impl) List(spaceID, userID string, filter vacancyapimodels.VrFilter) (list []dbmodels.VacancyRequest, err error) {
	list = []dbmodels.VacancyRequest{}
	tx := i.db.
		Model(dbmodels.VacancyRequest{}).
		Select("vacancy_requests.*, f.selected as favorite, p.selected as pinned").
		Joins("left join vr_favorites as f on vacancy_requests.id = f.vacancy_request_id and f.space_user_id = ?", userID).
		Joins("left join vr_pinneds as p on vacancy_requests.id = p.vacancy_request_id and p.space_user_id = ?", userID).
		Where("space_id = ?", spaceID)
	i.addFilter(tx, filter)
	page, limit := filter.GetPage()
	i.setPage(tx, page, limit)
	tx = tx.Preload(clause.Associations).Preload("ApprovalStages.SpaceUser")
	err = tx.Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) SetPin(id, userID string) error {
	rec := dbmodels.VrPinned{
		VacancyRequestID: id,
		SpaceUserID:      userID,
		Selected:         true,
	}
	err := i.db.
		Save(&rec).
		Error
	if err != nil {
		if strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			return nil
		}
		return errors.Wrap(err, "ошибка закрепления заявки на вакансию")
	}
	return nil
}

func (i impl) RemovePin(id, userID string) error {
	rec := dbmodels.VrPinned{}
	err := i.db.Model(&dbmodels.VrPinned{}).
		Where("space_user_id = ?", userID).
		Where("vacancy_request_id = ?", id).
		Delete(&rec).Error
	if err != nil {
		return errors.Wrap(err, "ошибка открепления заявки на вакансию")
	}
	return nil
}

func (i impl) SetFavorite(id, userID string) error {
	rec := dbmodels.VrFavorite{
		VacancyRequestID: id,
		SpaceUserID:      userID,
		Selected:         true,
	}
	err := i.db.
		Save(&rec).
		Error
	if err != nil {
		if strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			return nil
		}
		return errors.Wrap(err, "ошибка добавления в избранное")
	}
	return nil
}

func (i impl) RemoveFavorite(id, userID string) error {
	rec := dbmodels.VrFavorite{}
	err := i.db.Model(&dbmodels.VrFavorite{}).
		Where("space_user_id = ?", userID).
		Where("vacancy_request_id = ?", id).
		Delete(&rec).Error
	if err != nil {
		return errors.Wrap(err, "ошибка удаления из избранного")
	}
	return nil
}

func (i impl) addSort(tx *gorm.DB, sort vacancyapimodels.VrSort) {
	tx.Order("p.selected")
	if sort.CreatedAtDesc {
		tx = tx.Order("vacancy_requests.created_at desc")
	} else {
		tx = tx.Order("vacancy_requests.created_at asc")
	}
}

func (i impl) addFilter(tx *gorm.DB, filter vacancyapimodels.VrFilter) {
	if filter.Favorite {
		tx = tx.Where("f.selected = true")
	}
	if len(filter.Statuses) != 0 {
		tx = tx.Where("status in (?)", filter.Statuses)
	}
	if filter.CityID != "" {
		tx = tx.Where("city_id = ?", filter.CityID)
	}
	if filter.Search != "" {
		tx.Where("LOWER(vacancy_name) like ?", "%"+strings.ToLower(filter.Search)+"%")
	}
	if filter.SelectionType != "" {
		tx = tx.Where("selection_type = ?", filter.SelectionType)
	}
	if filter.AuthorID != "" {
		tx = tx.Where("author_id = ?", filter.AuthorID)
	}
	if filter.SearchPeriod != 0 {
		now := time.Now()
		toDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		toDate = toDate.AddDate(0, 0, 1)
		switch filter.SearchPeriod {
		case 1: //За день
			tx = tx.Where("vacancy_requests.created_at between ? and ?", toDate.AddDate(0, 0, -1), toDate)
			break
		case 2: //за 3 дня
			tx = tx.Where("vacancy_requests.created_at between ? and ?", toDate.AddDate(0, 0, -3), toDate)
			break
		case 3: //за неделю
			tx = tx.Where("vacancy_requests.created_at between ? and ?", toDate.AddDate(0, 0, -7), toDate)
			break
		case 4: //за 30 дней
			tx = tx.Where("vacancy_requests.created_at between ? and ?", toDate.AddDate(0, 0, -30), toDate)
			break
		case 5: //за пероид
			filterFrom := filter.GetSearchFrom()
			if !filterFrom.IsZero() {
				searchFrom := time.Date(filterFrom.Year(), filterFrom.Month(), filterFrom.Day(), 0, 0, 0, 0, filterFrom.Location())
				tx = tx.Where("vacancy_requests.created_at >= ?", searchFrom)
			}
			filterTo := filter.GetSearchTo()
			if !filterTo.IsZero() {
				searchTo := time.Date(filterTo.Year(), filterTo.Month(), filterTo.Day(), 0, 0, 0, 0, filterTo.Location())
				searchTo = searchTo.AddDate(0, 0, 1)
				tx = tx.Where("vacancy_requests.created_at <= ?", searchTo)
			}
			break
		}
	}
	i.addSort(tx, filter.Sort)
}

func (i impl) setPage(tx *gorm.DB, page, limit int) {
	offset := (page - 1) * limit
	tx.Limit(limit).Offset(offset)
}
