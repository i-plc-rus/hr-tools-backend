package applicanthistoryhandler

import (
	"fmt"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"reflect"
	"time"
)

func GetStageChange(stageName string) dbmodels.ApplicantChanges {
	return dbmodels.ApplicantChanges{
		Description: fmt.Sprintf("Перевод на этап %v", stageName),
	}
}

func GetDuplicateMark(minorRec dbmodels.Applicant) dbmodels.ApplicantChanges {
	fio := fmt.Sprintf("%v %v %v", minorRec.LastName, minorRec.FirstName, minorRec.MiddleName)
	return dbmodels.ApplicantChanges{
		Description: fmt.Sprintf("Кандидат (%v) отмечен как дубликат этого профиля", fio),
		Data: []dbmodels.ApplicantChange{
			{
				Field:    "duplicate_id",
				OldValue: nil,
				NewValue: minorRec.ID,
			},
		},
	}
}

func GetNotDuplicateMark(minorRec dbmodels.Applicant) dbmodels.ApplicantChanges {
	fio := fmt.Sprintf("%v %v %v", minorRec.LastName, minorRec.FirstName, minorRec.MiddleName)
	return dbmodels.ApplicantChanges{
		Description: fmt.Sprintf("Кандидат (%v) отмечен что не является дубликатом этого профиля", fio),
		Data: []dbmodels.ApplicantChange{
			{
				Field:    "not_duplicate_id",
				OldValue: nil,
				NewValue: minorRec.ID,
			},
		},
	}
}

func GetCreateChanges(descr string, rec dbmodels.Applicant) dbmodels.ApplicantChanges {
	result := dbmodels.ApplicantChanges{
		Description: descr,
		Data:        make([]dbmodels.ApplicantChange, 0),
	}
	rType := reflect.TypeOf(rec)
	vType := reflect.ValueOf(rec)
	for k := 0; k < rType.NumField(); k++ {
		field := rType.Field(k)
		fieldName := helpers.ToSnakeCase(field.Name)
		if ignoreFields[fieldName] {
			// пропускаем не нужные поля
			continue
		}
		if vType.Field(k).IsZero() {
			// пропускаем пустые поля
			continue
		}
		comment := field.Tag.Get(dbmodels.CommentTag)
		value := getValue(vType.Field(k).Interface())
		change := dbmodels.ApplicantChange{
			Field:    fieldName,
			OldValue: "",
			NewValue: value,
		}
		if comment != "" {
			change.Field = comment
		}
		result.Data = append(result.Data, change)

	}
	paramChanges := getParamChanges(dbmodels.ApplicantParams{}, rec.Params)
	if len(paramChanges) != 0 {
		result.Data = append(result.Data, paramChanges...)
	}
	return result
}

func GetUpdateChanges(descr string, rec dbmodels.Applicant, updMap map[string]interface{}) dbmodels.ApplicantChanges {
	result := dbmodels.ApplicantChanges{
		Description: descr,
		Data:        make([]dbmodels.ApplicantChange, 0, len(updMap)),
	}
	if len(updMap) == 0 {
		return result
	}
	recMap := map[string]interface{}{}
	recCommentMap := map[string]string{}
	rType := reflect.TypeOf(rec)
	vType := reflect.ValueOf(rec)
	for k := 0; k < rType.NumField(); k++ {
		field := rType.Field(k)
		fieldName := helpers.ToSnakeCase(field.Name)
		recCommentMap[fieldName] = field.Tag.Get(dbmodels.CommentTag)
		recMap[fieldName] = getValue(vType.Field(k).Interface())
	}

	for key, value := range updMap {
		fieldName := helpers.ToSnakeCase(key)
		newValue := getValue(value)
		if ignoreFields[fieldName] {
			if fieldName == "params" {
				paramChanges := getParamChanges(rec.Params, value.(dbmodels.ApplicantParams))
				if len(paramChanges) != 0 {
					result.Data = append(result.Data, paramChanges...)
				}
			}
			continue
		}
		change := dbmodels.ApplicantChange{
			Field:    fieldName,
			OldValue: "",
			NewValue: newValue,
		}
		oldValue, ok := recMap[fieldName]
		if ok {
			change.OldValue = oldValue
		}
		if change.OldValue == change.NewValue {
			// пропускаем поля без изменений
			continue
		}

		comment, ok := recCommentMap[fieldName]
		if ok && comment != "" {
			change.Field = comment
		}
		result.Data = append(result.Data, change)
	}
	return result
}

func GetRejectChange(reason string, rec dbmodels.Applicant, updMap map[string]interface{}) dbmodels.ApplicantChanges {
	return GetUpdateChanges(fmt.Sprintf("Кандидат отклонен по причине: %v", reason), rec, updMap)
}

func GetArchiveChange(reason string) dbmodels.ApplicantChanges {
	return dbmodels.ApplicantChanges{
		Description: fmt.Sprintf("Кандидат перемещен в архив по причине: %v", reason),
	}
}

func getParamChanges(oldParams, newParams dbmodels.ApplicantParams) []dbmodels.ApplicantChange {
	result := []dbmodels.ApplicantChange{}
	rType := reflect.TypeOf(oldParams)
	vOldType := reflect.ValueOf(oldParams)
	vNewType := reflect.ValueOf(newParams)
	for k := 0; k < rType.NumField(); k++ {
		field := rType.Field(k)
		fieldName := helpers.ToSnakeCase(field.Name)

		newValue := getValue(vNewType.Field(k).Interface())
		oldValue := getValue(vOldType.Field(k).Interface())

		comment := field.Tag.Get(dbmodels.CommentTag)
		change := dbmodels.ApplicantChange{
			Field:    fieldName,
			OldValue: oldValue,
			NewValue: newValue,
		}
		if change.NewValue == change.OldValue {
			continue
		}
		if comment != "" {
			change.Field = comment
		}
		result = append(result, change)
	}
	return result
}

var ignoreFields = map[string]bool{"base_space_model": true, "not_duplicates": true, "params": true, "vacancy": true,
	"selection_stage": true, "duplicates": true, "space_id": true, "vacancy_id": true}

func getValue(value interface{}) interface{} {
	xType := fmt.Sprintf("%T", value)
	switch xType {
	case "models.RelocationType":
		return value.(models.RelocationType).ToString()
	case "models.GenderType":
		return value.(models.GenderType).ToString()
	case "[]models.Employment":
		result := []string{}
		for _, item := range value.([]models.Employment) {
			result = append(result, item.ToString())
		}
		return fmt.Sprintf("%+v", result)
	case "[]models.Schedule":
		result := []string{}
		for _, item := range value.([]models.Schedule) {
			result = append(result, item.ToString())
		}
		return fmt.Sprintf("%+v", result)
	case "models.EducationType":
		return value.(models.EducationType).ToString()
	case "models.TripReadinessType":
		return value.(models.TripReadinessType).ToString()
	case "[]dbmodels.Language":
		result := []string{}
		for _, item := range value.([]dbmodels.Language) {
			result = append(result, fmt.Sprintf("%v - %v", item.Name, item.LanguageLevel))
		}
		return fmt.Sprintf("%+v", result)
	case "bool":
		if value.(bool) {
			return "да"
		}
		return "нет"
	case "time.Time":
		return value.(time.Time).Format("02.01.2006")
	}
	return fmt.Sprintf("%+v", value)
}
