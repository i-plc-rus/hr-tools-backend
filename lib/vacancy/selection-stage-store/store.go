package selectionstagestore

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	Create(rec dbmodels.SelectionStage) (id string, err error)
	Update(spaceID, vacancyID, id string, updMap map[string]interface{}) error
	GetByID(spaceID, vacancyID, id string) (*dbmodels.SelectionStage, error)
	List(spaceID, vacancyID string) (list []dbmodels.SelectionStage, err error)
	Delete(spaceID, vacancyID, id string) (err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Create(rec dbmodels.SelectionStage) (id string, err error) {
	maxOrder, err := i.maxOrder(rec.SpaceID, rec.VacancyID)
	if err != nil {
		return "", err
	}
	rec.StageOrder = maxOrder + 1
	err = i.db.
		Save(&rec).
		Error
	if err != nil {
		return "", err
	}
	return rec.ID, nil
}

func (i impl) GetByID(spaceID, vacancyID, id string) (*dbmodels.SelectionStage, error) {
	rec := dbmodels.SelectionStage{}
	err := i.db.
		Model(&dbmodels.SelectionStage{}).
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
		Where("vacancy_id = ?", vacancyID).
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

func (i impl) Update(spaceID, vacancyID, id string, updMap map[string]interface{}) error {
	if len(updMap) == 0 {
		return nil
	}
	err := i.db.
		Model(&dbmodels.SelectionStage{}).
		Where("id = ?", id).
		Where("space_id = ?", spaceID).
		Where("vacancy_id = ?", vacancyID).
		Updates(updMap).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) List(spaceID, vacancyID string) (list []dbmodels.SelectionStage, err error) {
	list = []dbmodels.SelectionStage{}
	tx := i.db.
		Where("space_id = ?", spaceID).
		Where("vacancy_id = ?", vacancyID)
	err = tx.Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (i impl) Delete(spaceID, vacancyID, id string) (err error) {
	delRec := dbmodels.SelectionStage{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			BaseModel: dbmodels.BaseModel{ID: id},
			SpaceID:   spaceID,
		},
		VacancyID: vacancyID,
	}
	err = i.db.
		Delete(&delRec).
		Error

	if err != nil {
		return err
	}
	return nil
}

func (i impl) maxOrder(spaceID, vacancyID string) (order int, err error) {
	type result struct {
		MaxOrder int
	}
	res := result{}
	err = i.db.Table("selection_stages").
		Where("space_id = ?", spaceID).
		Where("vacancy_id = ?", vacancyID).
		Select("max(stage_order) as max_order").Find(&res).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return res.MaxOrder, nil
}
