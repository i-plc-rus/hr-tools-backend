package xlsexport

import (
	"bytes"
	"fmt"
	applicantapimodels "hr-tools-backend/models/api/applicant"
	dbmodels "hr-tools-backend/models/db"
	"math"
	"strconv"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type Provider interface {
	ExportApplicantList(list []dbmodels.ApplicantWithJob) (*bytes.Buffer, error)
	ExportSource(data applicantapimodels.ApplicantSourceData) (*bytes.Buffer, error)
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

func (i impl) ExportSource(source applicantapimodels.ApplicantSourceData) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.WithError(err).Error("ошибка закрытия файла")
		}
	}()
	err := addSoureChart(f, "Источники_кандидатов", 0, source)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка формирования файла с источниками кандидатов")
	}
	f.DeleteSheet("Sheet1")
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

func putSoureChart(f *excelize.File, sheet string, row int, chartName string, sourceData []applicantapimodels.SourceItem, sourceTotal int) (int, error) {
	// Заголовок
	row++
	сellFrom, err := excelize.CoordinatesToCellName(1, row)
	if err != nil {
		return row, err
	}
	сellTo, err := excelize.CoordinatesToCellName(3, row)
	if err != nil {
		return row, err
	}
	err = f.MergeCell(sheet, сellFrom, сellTo)
	if err != nil {
		return row, err
	}
	if err := writeColumn(f, sheet, 1, row, chartName); err != nil {
		return row, err
	}
	// стиль для заголовка
	style, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Font: &excelize.Font{
			Bold:   true,
			Italic: false,
			Family: "Times New Roman",
			Size:   11,
		},
	})
	if err != nil {
		return row, err
	}
	if err = f.SetCellStyle(sheet, сellFrom, сellFrom, style); err != nil {
		return row, err
	}
	// хедер таблицы
	row++
	cell, err := excelize.CoordinatesToCellName(1, row)
	if err != nil {
		return row, err
	}
	rowValue := []interface{}{
		"Источник", "Количество", "Процент",
	}
	if err := f.SetSheetRow(sheet, cell, &rowValue); err != nil {
		return row, err
	}
	// стиль для хедера таблицы
	style, err = f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Font: &excelize.Font{
			Bold:   true,
			Italic: false,
			Family: "Times New Roman",
			Size:   11,
		},
	})
	сellFrom, err = excelize.CoordinatesToCellName(1, row)
	if err != nil {
		return row, err
	}
	сellTo, err = excelize.CoordinatesToCellName(3, row)
	if err != nil {
		return row, err
	}
	if err = f.SetCellStyle(sheet, сellFrom, сellTo, style); err != nil {
		return row, err
	}

	// Данные
	rowStart := row + 1
	totalPercent := 0
	for k, data := range sourceData {
		row++
		cell, err = excelize.CoordinatesToCellName(1, row)
		if err != nil {
			return row, err
		}
		percent := 0
		if k != len(sourceData)-1 {
			percent = int(math.Round(float64(data.Count) / float64(sourceTotal) * 100))
			totalPercent += percent
		} else {
			percent = 100 - totalPercent
		}
		rowValue = []interface{}{
			data.Name, data.Count, percent,
		}
		if err := f.SetSheetRow(sheet, cell, &rowValue); err != nil {
			return row, err
		}
	}
	rowEnd := row
	// стиль для данных
	style, err = f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Font: &excelize.Font{
			Bold:   false,
			Italic: false,
			Family: "Times New Roman",
			Size:   11,
		},
	})
	сellFrom, err = excelize.CoordinatesToCellName(1, rowStart)
	if err != nil {
		return row, err
	}
	сellTo, err = excelize.CoordinatesToCellName(3, rowEnd)
	if err != nil {
		return row, err
	}
	if err = f.SetCellStyle(sheet, сellFrom, сellTo, style); err != nil {
		return row, err
	}

	// График
	chart := excelize.Chart{
		Type: excelize.Doughnut,
		Dimension: excelize.ChartDimension{
			Width:  400,
			Height: 300,
		},
		Format: excelize.GraphicOptions{
			OffsetX: 60,
		},
		Series: []excelize.ChartSeries{
			{
				Name:       "Amount",
				Categories: sheet + "!$A$" + strconv.Itoa(rowStart) + ":$A$" + strconv.Itoa(rowEnd),
				Values:     sheet + "!$B$" + strconv.Itoa(rowStart) + ":$B$" + strconv.Itoa(rowEnd),
			},
		},
		Legend: excelize.ChartLegend{
			Position: "bottom",
		},
		Title: []excelize.RichTextRun{
			{
				Text: chartName,
			},
		},
		PlotArea: excelize.ChartPlotArea{
			ShowCatName:     false,
			ShowLeaderLines: false,
			ShowPercent:     false,
			ShowSerName:     false,
			ShowVal:         false,
		},
		ShowBlanksAs: "zero",
	}
	row += 2
	cell, err = excelize.CoordinatesToCellName(1, row)
	if err != nil {
		return row, err
	}
	if err := f.AddChart(sheet, cell, &chart); err != nil {
		return row, err
	}
	row += 16
	return row, nil
}

func addSoureChart(f *excelize.File, sheet string, row2 int, source applicantapimodels.ApplicantSourceData) error {
	index, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	f.SetActiveSheet(index)
	if err = f.SetColWidth(sheet, "A", "C", 25); err != nil {
		return err
	}

	row2, err = putSoureChart(f, sheet, row2, "Общая статистика", source.TotalSource.Data, source.TotalSource.Total)
	if err != nil {
		return err
	}

	row2, err = putSoureChart(f, sheet, row2, "Откликнулись", source.NegotiationSource.Data, source.NegotiationSource.Total)
	if err != nil {
		return err
	}
	row2, err = putSoureChart(f, sheet, row2, "Добавлены", source.AddingSource.Data, source.AddingSource.Total)
	if err != nil {
		return err
	}

	return nil
}
