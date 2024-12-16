package vacancystore

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"hr-tools-backend/models"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
	"strings"
)

type Provider interface {
	Create(rec dbmodels.Vacancy) (id string, err error)
	GetByID(spaceID, id string) (rec *dbmodels.Vacancy, err error)
	Update(spaceID, id string, updMap map[string]interface{}) error
	Delete(spaceID, id string) error
	ListCount(spaceID, userID string, filter vacancyapimodels.VacancyFilter) (count int64, err error)
	List(spaceID, userID string, filter vacancyapimodels.VacancyFilter) (list []dbmodels.VacancyExt, err error)
	SetPin(vacancyID, userID string) error
	RemovePin(vacancyID, userID string) error
	SetFavorite(vacancyID, userID string) error
	RemoveFavorite(vacancyID, userID string) error
	ListAvitoByStatus(spaceID string, status models.VacancyPubStatus) (list []dbmodels.Vacancy, err error)
	ListHhByStatus(spaceID string, status models.VacancyPubStatus) (list []dbmodels.Vacancy, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.Vacancy) (id string, err error) {
	err = i.db.Omit(clause.Associations).
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) GetByID(spaceID, id string) (*dbmodels.Vacancy, error) {
	rec := dbmodels.Vacancy{}
	err := i.db.
		Model(&dbmodels.Vacancy{}).
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
		Preload(clause.Associations).
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
		Model(&dbmodels.Vacancy{}).
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
	rec := dbmodels.Vacancy{
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

func (i impl) ListCount(spaceID, userID string, filter vacancyapimodels.VacancyFilter) (count int64, err error) {
	var rowCount int64
	tx := i.db.
		Model(dbmodels.Vacancy{}).
		Joins("left join favorites as f on vacancies.id = f.vacancy_id and f.space_user_id = ?", userID).
		Joins("left join pinneds as p on vacancies.id = p.vacancy_id and p.space_user_id = ?", userID).
		Where("space_id = ?", spaceID)
	i.addFilter(tx, filter, userID)
	err = tx.Count(&rowCount).Error
	if err != nil {
		log.WithError(err).Error("ошибка получения общего количества вакансий")
		return 0, errors.New("ошибка получения общего количества вакансий")
	}
	return rowCount, nil
}

func (i impl) List(spaceID, userID string, filter vacancyapimodels.VacancyFilter) (list []dbmodels.VacancyExt, err error) {
	list = []dbmodels.VacancyExt{}
	tx := i.db.
		Model(dbmodels.Vacancy{}).
		Select("vacancies.*, f.selected as favorite, p.selected as pinned").
		Joins("left join favorites as f on vacancies.id = f.vacancy_id and f.space_user_id = ?", userID).
		Joins("left join pinneds as p on vacancies.id = p.vacancy_id and p.space_user_id = ?", userID).
		Where("space_id = ?", spaceID)
	i.addFilter(tx, filter, userID)
	page, limit := filter.GetPage()
	i.setPage(tx, page, limit)
	err = tx.Preload(clause.Associations).Find(&list).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) SetPin(vacancyID, userID string) error {
	rec := dbmodels.Pinned{
		VacancyID:   vacancyID,
		SpaceUserID: userID,
		Selected:    true,
	}
	err := i.db.
		Save(&rec).
		Error
	if err != nil {
		if strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			return nil
		}
		return errors.Wrap(err, "ошибка закрепления вакансии")
	}
	return nil
}

func (i impl) RemovePin(vacancyID, userID string) error {
	rec := dbmodels.Pinned{}
	err := i.db.Model(&dbmodels.Pinned{}).
		Where("space_user_id = ?", userID).
		Where("vacancy_id = ?", vacancyID).
		Delete(&rec).Error
	if err != nil {
		return errors.Wrap(err, "ошибка открепления вакансии")
	}
	return nil
}

func (i impl) SetFavorite(vacancyID, userID string) error {
	rec := dbmodels.Favorite{
		VacancyID:   vacancyID,
		SpaceUserID: userID,
		Selected:    true,
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

func (i impl) RemoveFavorite(vacancyID, userID string) error {
	rec := dbmodels.Favorite{}
	err := i.db.Model(&dbmodels.Favorite{}).
		Where("space_user_id = ?", userID).
		Where("vacancy_id = ?", vacancyID).
		Delete(&rec).Error
	if err != nil {
		return errors.Wrap(err, "ошибка удаления из избранного")
	}
	return nil
}

func (i impl) ListAvitoByStatus(spaceID string, status models.VacancyPubStatus) (list []dbmodels.Vacancy, err error) {
	list = []dbmodels.Vacancy{}
	tx := i.db.
		Model(dbmodels.Vacancy{}).
		Where("space_id = ?", spaceID).
		Where("avito_status = ?", status).
		Preload(clause.Associations)
	err = tx.Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) ListHhByStatus(spaceID string, status models.VacancyPubStatus) (list []dbmodels.Vacancy, err error) {
	list = []dbmodels.Vacancy{}
	tx := i.db.
		Model(dbmodels.Vacancy{}).
		Where("space_id = ?", spaceID).
		Where("hh_status = ?", status).
		Preload(clause.Associations)
	err = tx.Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) addSort(tx *gorm.DB, sort vacancyapimodels.VacancySort) {
	tx.Order("p.selected")
	if sort.CreatedAtDesc {
		tx = tx.Order("vacancies.created_at desc")
	} else {
		tx = tx.Order("vacancies.created_at asc")
	}
}

func (i impl) addFilter(tx *gorm.DB, filter vacancyapimodels.VacancyFilter, userID string) {
	if filter.Tab > 0 {
		switch filter.Tab {
		case vacancyapimodels.VacancyTabMy:
			tx = tx.Where("author_id = ?", userID)
		case vacancyapimodels.VacancyTabOther:
			tx = tx.Where("author_id <> ?", userID)
		case vacancyapimodels.VacancyTabArch:
			tx = tx.Where("status = ?", models.VacancyStatusClosed)
		}
	}
	if filter.VacancyRequestID != "" {
		tx = tx.Where("vacancy_request_id = ?", filter.VacancyRequestID)
	}
	if filter.Favorite {
		tx = tx.Where("f.selected = true")
	}
	if len(filter.Statuses) != 0 {
		tx = tx.Where("status in (?)", filter.Statuses)
	}
	if filter.CityID != "" {
		tx = tx.Where("city_id = ?", filter.CityID)
	}
	if filter.DepartmentID != "" {
		tx = tx.Where("department_id = ?", filter.DepartmentID)
	}
	if filter.Search != "" {
		tx.Where("LOWER(vacancy_name) like ?", "%"+strings.ToLower(filter.Search)+"%")
	}
	if filter.SelectionType != "" {
		tx = tx.Where("selection_type = ?", filter.SelectionType)
	}
	if filter.RequestType != "" {
		tx = tx.Where("request_type = ?", filter.RequestType)
	}
	if filter.Urgency != "" {
		tx = tx.Where("urgency = ?", filter.Urgency)
	}
	if filter.AuthorID != "" {
		tx = tx.Where("author_id = ?", filter.AuthorID)
	}
	if filter.RequestAuthorID != "" {
		subQuery := i.db.Select("id").Where("author_id = ?", filter.RequestAuthorID).Table("vacancy_requests")
		tx.Where("vacancy_request_id in (?)", subQuery)
	}
	i.addSort(tx, filter.Sort)
}

func (i impl) setPage(tx *gorm.DB, page, limit int) {
	offset := (page - 1) * limit
	tx.Limit(limit).Offset(offset)
}
