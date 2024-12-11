package xlsexport

import "github.com/xuri/excelize/v2"

func writeColumn(f *excelize.File, sheet string, col, row int, value interface{}) error {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, cell, value); err != nil {
		return err
	}
	return nil
}

func writeHeader(f *excelize.File, sheet string, row int, headers []string) (int, error) {
	row++
	style, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Font: &excelize.Font{
			Bold:   true,
			Italic: false,
			Family: "Times New Roman",
			Size:   11,
			// Color:  "777777",
		},
	})
	if err != nil {
		return row, err
	}
	cellFirst, err := excelize.CoordinatesToCellName(1, row)
	if err != nil {
		return row, err
	}
	cellLast, err := excelize.CoordinatesToCellName(len(headers), row)
	if err != nil {
		return row, err
	}

	if err = f.SetCellStyle(sheet, cellFirst, cellLast, style); err != nil {
		return row, err
	}
	lastCol, err := excelize.ColumnNumberToName(len(headers))
	if err != nil {
		return row, err
	}

	if err = f.SetColWidth(sheet, "A", lastCol, 25); err != nil {
		return row, err
	}

	for idx, value := range headers {
		if err = writeColumn(f, sheet, idx+1, row, value); err != nil {
			return row, err
		}
	}
	return row, nil
}

func applyDataCellStyle(f *excelize.File, sheet string, coFrom, rowFrom, colTo, rowTo int) error {
	style, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
		Font: &excelize.Font{
			Bold:   false,
			Italic: false,
			Family: "Times New Roman",
			Size:   11,
			// Color:  "777777",
		},
	})
	if err != nil {
		return err
	}
	cellFirst, err := excelize.CoordinatesToCellName(coFrom, rowFrom)
	if err != nil {
		return err
	}
	cellLast, err := excelize.CoordinatesToCellName(colTo, rowTo)
	if err != nil {
		return err
	}
	if err = f.SetCellStyle(sheet, cellFirst, cellLast, style); err != nil {
		return err
	}
	return nil
}
