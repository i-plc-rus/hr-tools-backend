package xlsexport

import (
	"bytes"
	"fmt"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type Provider interface {
	ExportApplicantList(list []dbmodels.ApplicantWithJob) (*bytes.Buffer, error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{}
}

type impl struct{}

var applicantHeaders = []string{"ФИО", "Контакты", "Вакансия", "Должность", "Источник кандидата", "Дата отбора", "Желаемая ЗП", "Дата выхода", "Статус"}

func (i impl) ExportApplicantList(list []dbmodels.ApplicantWithJob) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.WithError(err).Error("ошибка закрытия файла")
		}
	}()
	sheet := "Sheet1"
	row := 0
	row, err := writeHeader(f, sheet, row, applicantHeaders)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка формирования заголовка в xlsx")
	}
	if len(list) != 0 {
		row, err = writeApplicantData(f, sheet, list, row)
		if err != nil {
			return nil, errors.Wrap(err, "ошибка формирования таблицы с данными в xlsx")
		}
	}
	f.SetSheetName(sheet, "Кандидаты")
	return f.WriteToBuffer()
}

func writeApplicantData(f *excelize.File, sheet string, list []dbmodels.ApplicantWithJob, row int) (int, error) {
	if err := applyDataCellStyle(f, sheet, 1, row+1, len(applicantHeaders), len(list)+1); err != nil {
		return row, err
	}
	for _, item := range list {
		row++
		// "ФИО"
		col := 1
		if err := writeColumn(f, sheet, col, row, item.GetFIO()); err != nil {
			return row, err
		}

		// "Контакты"
		col++
		if err := writeColumn(f, sheet, col, row, fmt.Sprintf("%v\r%v", item.Phone, item.Email)); err != nil {
			return row, err
		}

		// "Вакансия"
		col++
		if item.Vacancy != nil {
			if err := writeColumn(f, sheet, col, row, item.Vacancy.VacancyName); err != nil {
				return row, err
			}
		}

		// "Должность"
		col++
		if err := writeColumn(f, sheet, col, row, item.JobTitleName); err != nil {
			return row, err
		}

		// "Источник кандидата"
		col++
		if err := writeColumn(f, sheet, col, row, item.Source); err != nil {
			return row, err
		}

		// "Дата отбора"
		col++
		if !item.NegotiationAcceptDate.IsZero() {
			if err := writeColumn(f, sheet, col, row, item.NegotiationAcceptDate.Format("02.01.2006")); err != nil {
				return row, err
			}
		}

		// "Желаемая ЗП"
		col++
		if err := writeColumn(f, sheet, col, row, item.Salary); err != nil {
			return row, err
		}

		// "Дата выхода"
		col++
		if !item.StartDate.IsZero() {
			if err := writeColumn(f, sheet, col, row, item.StartDate.Format("02.01.2006")); err != nil {
				return row, err
			}
		}

		// "Статус"
		col++
		if err := writeColumn(f, sheet, col, row, item.Status); err != nil {
			return row, err
		}
	}
	return row, nil
}
