package spaceusersstore

import (
	spaceapimodels "hr-tools-backend/models/api/space"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
	dbmodels "hr-tools-backend/models/db"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Provider interface {
	Create(rec dbmodels.SpaceUser) (string, error)
	Update(userID string, updMap map[string]interface{}) error
	Delete(userID string) error
	GetCountList(spaceID string, filter spaceapimodels.SpaceUserFilter) (count int64, err error)
	GetList(spaceID string, filter spaceapimodels.SpaceUserFilter) (userList []dbmodels.SpaceUser, err error)
	ExistByEmail(email string) (bool, error)
	FindByEmail(email string, checkNew bool) (rec *dbmodels.SpaceUser, err error)
	GetByID(userID string) (rec *dbmodels.SpaceUser, err error)
	GetByResetCode(code string) (rec *dbmodels.SpaceUser, err error)
	GetListForVacancy(spaceID, vacancyID string, filter vacancyapimodels.PersonFilter) (userList []dbmodels.SpaceUser, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) GetCountList(spaceID string, filter spaceapimodels.SpaceUserFilter) (count int64, err error) {
	var rowCount int64
	tx := i.db.
		Model(dbmodels.SpaceUser{}).
		Where("space_id = ?", spaceID)
	i.addFilter(tx, filter)
	err = tx.Count(&rowCount).Error
	if err != nil {
		return 0, err
	}
	return rowCount, nil
}

func (i impl) GetList(spaceID string, filter spaceapimodels.SpaceUserFilter) (userList []dbmodels.SpaceUser, err error) {
	tx := i.db.
		Model(dbmodels.SpaceUser{}).
		Select("space_users.*, (last_name || ' ' || first_name) as fio").
		Where("space_id = ?", spaceID)
	i.addFilter(tx, filter)
	i.addSort(tx, filter.Sort)
	page, limit := filter.GetPage()
	i.setPage(tx, page, limit)
	err = tx.
		Preload(clause.Associations).
		Find(&userList).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return userList, nil
}

func (i impl) Delete(userID string) error {
	return i.db.
		Where("id = ?", userID).
		Delete(&dbmodels.SpaceUser{}).
		Error
}

func (i impl) Update(userID string, updMap map[string]interface{}) error {
	err := i.db.
		Model(&dbmodels.SpaceUser{}).
		Where("id = ?", userID).
		Updates(updMap).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) GetByID(userID string) (rec *dbmodels.SpaceUser, err error) {
	err = i.db.Model(dbmodels.SpaceUser{}).
		Where("id = ?", userID).
		Preload(clause.Associations).
		First(&rec).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rec, nil
}

func (i impl) FindByEmail(email string, checkNew bool) (rec *dbmodels.SpaceUser, err error) {
	tx := i.db.Model(dbmodels.SpaceUser{}).
		Where("email = ?", email)
	if checkNew {
		tx.Or("new_email = ?", email)
	}
	err = tx.Preload(clause.Associations).First(&rec).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rec, nil
}

func (i impl) Create(rec dbmodels.SpaceUser) (string, error) {
	err := i.db.
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) ExistByEmail(email string) (bool, error) {
	err := i.db.
		Where("email = ?", email).
		First(&dbmodels.SpaceUser{}).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (i impl) GetByResetCode(code string) (rec *dbmodels.SpaceUser, err error) {
	err = i.db.Model(dbmodels.SpaceUser{}).
		Where("reset_code = ?", code).
		Preload(clause.Associations).
		First(&rec).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rec, nil
}

func (i impl) GetListForVacancy(spaceID, vacancyID string, filter vacancyapimodels.PersonFilter) (userList []dbmodels.SpaceUser, err error) {
	tx := i.db.Model(dbmodels.SpaceUser{})
	tx = tx.
		Where("space_id = ?", spaceID).
		Where("id not in (select id from vacancy_teams where vacancy_id = ?)", vacancyID)
	if filter.Search != "" {
		tx = tx.Where("LOWER(first_name|| ' ' || last_name) like ?", "%"+strings.ToLower(filter.Search)+"%")
	}
	err = tx.Find(&userList).Error
	if err != nil {
		return nil, err
	}
	return userList, nil
}

func (i impl) setPage(tx *gorm.DB, page, limit int) {
	offset := (page - 1) * limit
	tx.Limit(limit).Offset(offset)
}

func (i impl) addFilter(tx *gorm.DB, filter spaceapimodels.SpaceUserFilter) {
	if filter.Search != "" {
		tx.Where("LOWER(last_name || ' ' || first_name) like ?", "%"+strings.ToLower(filter.Search)+"%").
			Or("LOWER(email) like ?", "%"+strings.ToLower(filter.Search)+"%")
	}
}

func (i impl) addSort(tx *gorm.DB, sort spaceapimodels.SpaceUserSort) {
	if sort.NameDesc != nil {
		if *sort.NameDesc {
			tx = tx.Order("fio desc")
		} else {
			tx = tx.Order("fio asc")
		}
	}
	if sort.EmailDesc != nil {
		if *sort.EmailDesc {
			tx = tx.Order("email desc")
		} else {
			tx = tx.Order("email asc")
		}
	}
	if sort.RoleDesc != nil {
		if *sort.RoleDesc {
			tx = tx.Order("role desc")
		} else {
			tx = tx.Order("role asc")
		}
	}
}
