package db

import (
	log "github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	dbmodels "hr-tools-backend/models/db"
)

func correctHhRoles() {
	log.Info("корректировка ролей HH")

	// Open Excel file
	f, err := excelize.OpenFile("./static_preload/company_structs.xlsx")
	if err != nil {
		log.WithError(err).Error("ошибка чтения файла с преднастройками ролей")
		return
	}
	rows, err := f.GetRows("hh")
	if err != nil {
		log.Fatalf("Unable to read sheet: %v\n", err)
		return
	}
	for i, row := range rows {
		if i == 0 || i == 1 {
			continue // Skip header row
		}
		jobTitleName := row[7]
		hhRoleID := row[6]
		updTx := DB.Model(&dbmodels.JobTitle{}).Where("Name = ?", jobTitleName).Update("HhRoleID", hhRoleID)
		if updTx.Error != nil {
			log.WithError(updTx.Error).Errorf("ошибка корректировки записей с HhRoleID = %v", hhRoleID)
			return
		}
	}
}
