package questionhistorystore

import (
	dbmodels "hr-tools-backend/models/db"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Provider interface {
	Save(rec dbmodels.QuestionHistory) (err error)
	FindByText(text string, vacancyID *string) (rec *dbmodels.QuestionHistory, err error)
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct {
	db *gorm.DB
}

func (i impl) Save(rec dbmodels.QuestionHistory) (err error) {
	// сохраняем только уникальные вопросы по вакансии
	existedRec, err := i.FindByText(rec.Text, &rec.VacancyID)
	if err != nil {
		return err
	}
	if existedRec != nil {
		return nil
	}
	err = i.db.
		Save(&rec).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (i impl) FindByText(text string, vacancyID *string) (*dbmodels.QuestionHistory, error) {
	rec := dbmodels.QuestionHistory{}
	tx := i.db.
		Where("LOWER(text) = ?", strings.ToLower(text))
	if vacancyID != nil {
		tx = tx.Where("vacancy_id = ?", *vacancyID)
	}
	err := tx.First(&rec).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}
