package companystructload

import (
	"crypto/md5"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	dbmodels "hr-tools-backend/models/db"
	"log"
	"strconv"
	"strings"
)

func PreloadCompanyStruct(tx *gorm.DB, spaceID string) error {
	// Open Excel file
	f, err := excelize.OpenFile("./static_preload/company_structs.xlsx")
	if err != nil {
		return errors.Wrap(err, "Unable to open Excel file with preload data")
	}
	// Read and process data
	companyStructMap := make(map[string]string) // Map to store CompanyStruct UUIDs
	departmentMap := make(map[string]string)    // Map to store Department UUIDs

	rows, err := f.GetRows("hh")
	if err != nil {
		log.Fatalf("Unable to read sheet: %v\n", err)
	}

	for i, row := range rows {
		if i == 0 || i == 1 {
			continue // Skip header row
		}

		// Extract columns: Adjust indexes based on your actual data structure
		companyName := row[0]
		departmentName := row[1]
		avitoAreaID, _ := strconv.Atoi(row[2])
		jobTitleName := row[7]
		hhRoleID := row[4]

		if len(strings.TrimSpace(companyName)) == 0 {
			continue
		}

		// Insert into CompanyStruct if not exists
		companyStructKey := string(md5.New().Sum([]byte(companyName)))
		companyStructID, exists := companyStructMap[companyStructKey]
		if !exists {
			companyStructID = uuid.New().String()
			companyStructMap[companyStructKey] = companyStructID

			compStructRec := dbmodels.CompanyStruct{
				BaseSpaceModel: dbmodels.BaseSpaceModel{
					BaseModel: dbmodels.BaseModel{ID: companyStructID},
					SpaceID:   spaceID,
				},
				Name: companyName,
			}
			err := tx.Save(&compStructRec).Error
			/*err = tx.Exec("INSERT INTO company_structs (id, name, space_id) VALUES ($1, $2, $3)",
			companyStructID, companyName, spaceID).Error*/
			if err != nil {
				return errors.Wrap(err, "Unable to insert into CompanyStruct")
			}
		}

		// Insert into Department if not exists
		departmentKey := string(md5.New().Sum([]byte(fmt.Sprintf("%s-%s", companyName, departmentName))))
		departmentID, exists := departmentMap[departmentKey]
		if !exists {
			departmentID = uuid.New().String()
			departmentMap[departmentKey] = departmentID

			depRec := dbmodels.Department{
				BaseSpaceModel: dbmodels.BaseSpaceModel{
					BaseModel: dbmodels.BaseModel{
						ID: departmentID,
					},
					SpaceID: spaceID,
				},
				CompanyStructID: companyStructID,
				Name:            departmentName,
				BusinessAreaID:  avitoAreaID,
			}
			err = tx.Save(&depRec).Error

			/*err = tx.Exec("INSERT INTO departments (id, name, company_struct_id, business_area_id, space_id) VALUES ($1, $2, $3, $4, $5)",
			departmentID, departmentName, companyStructID, avitoAreaID, spaceID).Error*/
			if err != nil {
				return errors.Wrap(err, "Unable to insert into Department")
			}
		}

		// Insert into JobTitle

		jobTitleRec := dbmodels.JobTitle{
			BaseSpaceModel: dbmodels.BaseSpaceModel{
				SpaceID: spaceID,
			},
			DepartmentID: departmentID,
			Name:         jobTitleName,
			HhRoleID:     hhRoleID,
		}
		err = tx.Save(&jobTitleRec).Error
		/*err = tx.Exec("INSERT INTO job_titles (id, department_id, name, hh_role_id, space_id) VALUES ($1, $2, $3, $4, $5)",
		jobTitleID, departmentID, jobTitleName, hhRoleID, spaceID).Error*/
		if err != nil {
			return errors.Wrap(err, "Unable to insert into JobTitle")
		}
	}
	return nil
}
