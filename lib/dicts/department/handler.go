package departmentprovider

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	companystructprovider "hr-tools-backend/lib/dicts/company-struct"
	"hr-tools-backend/lib/dicts/department/store"
	initchecker "hr-tools-backend/lib/utils/init-checker"
	dictapimodels "hr-tools-backend/models/api/dict"
	dbmodels "hr-tools-backend/models/db"
	"strings"
)

type Provider interface {
	Create(spaceID string, request dictapimodels.DepartmentData) (id string, err error)
	Update(spaceID, id string, request dictapimodels.DepartmentData) error
	Get(spaceID, id string) (item dictapimodels.DepartmentView, err error)
	FindByName(spaceID string, request dictapimodels.DepartmentFind) (list []dictapimodels.DepartmentTreeItem, err error)
	Delete(spaceID, id string) error
}

var Instance Provider

func NewHandler() {
	instance := impl{
		store:         store.NewInstance(db.DB),
		companyStruct: companystructprovider.Instance,
	}
	initchecker.CheckInit(
		"store", instance.store,
		"companyStruct", instance.companyStruct,
	)
	Instance = instance
}

type impl struct {
	store         store.Provider
	companyStruct companystructprovider.Provider
}

func (i impl) Create(spaceID string, request dictapimodels.DepartmentData) (id string, err error) {
	logger := log.WithField("space_id", spaceID)
	err = i.isUserCompanyStruct(request.CompanyStructID, spaceID)
	if err != nil {
		return "", err
	}
	rec := dbmodels.Department{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		CompanyStructID: request.CompanyStructID,
		ParentID:        request.ParentID,
		Name:            request.Name,
	}
	id, err = i.store.Create(rec)
	if err != nil {
		return "", err
	}
	logger.
		WithField("department_name", rec.Name).
		WithField("rec_id", rec.ID).
		Info("создано подразделение")
	return id, nil
}

func (i impl) Update(spaceID, id string, request dictapimodels.DepartmentData) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	updMap := map[string]interface{}{
		"name": request.Name,
	}
	err := i.store.Update(spaceID, id, updMap)
	if err != nil {
		return err
	}
	logger.Info("обновлено подразделение")
	return nil
}

func (i impl) Get(spaceID, id string) (item dictapimodels.DepartmentView, err error) {
	rec, err := i.store.GetByID(spaceID, id)
	if err != nil {
		return dictapimodels.DepartmentView{}, err
	}
	if rec == nil {
		return dictapimodels.DepartmentView{}, errors.New("подразделение не найдено")
	}
	return dictapimodels.DepartmentConvert(*rec), nil
}

func (i impl) FindByName(spaceID string, request dictapimodels.DepartmentFind) (list []dictapimodels.DepartmentTreeItem, err error) {
	recList, err := i.store.FindByCompanyStruct(spaceID, request.CompanyStructID)
	if err != nil {
		return nil, err
	}

	tree := []dictapimodels.DepartmentTreeItem{}
	for _, rec := range recList {
		if rec.ParentID != "" {
			continue
		}
		item := dictapimodels.DepartmentTreeItem{
			DepartmentView: dictapimodels.DepartmentConvert(rec),
			SubUnits:       getChildren(rec.ID, recList),
		}
		tree = append(tree, item)
	}

	if request.Name == "" {
		return tree, nil
	}
	result := make([]dictapimodels.DepartmentTreeItem, 0, len(list))
	searchName := strings.ToLower(request.Name)
	for _, treeItem := range tree {
		found, subUnits := filterTree(searchName, treeItem)
		if found {
			treeItem.SubUnits = subUnits
			result = append(result, treeItem)
		}
	}
	return result, nil
}

func (i impl) Delete(spaceID, id string) error {
	logger := log.WithField("space_id", spaceID).
		WithField("rec_id", id)
	err := i.store.Delete(spaceID, id)
	if err != nil {
		return err
	}
	logger.Info("удалено подразделение")
	return nil
}

func (i impl) isUserCompanyStruct(companyStructID, spaceID string) error {
	_, err := i.companyStruct.Get(spaceID, companyStructID)
	if err != nil {
		return err
	}
	return nil
}

func getChildren(rootID string, recList []dbmodels.Department) []dictapimodels.DepartmentTreeItem {
	result := []dictapimodels.DepartmentTreeItem{}
	for _, rec := range recList {
		if rec.ParentID != rootID {
			continue
		}
		item := dictapimodels.DepartmentTreeItem{
			DepartmentView: dictapimodels.DepartmentConvert(rec),
			SubUnits:       getChildren(rec.ID, recList),
		}
		result = append(result, item)
	}
	return result
}

func filterTree(searchName string, item dictapimodels.DepartmentTreeItem) (bool, []dictapimodels.DepartmentTreeItem) {
	foundSubUnits := []dictapimodels.DepartmentTreeItem{}
	for _, subUnit := range item.SubUnits {
		found, foundList := filterTree(searchName, subUnit)
		if found {
			subUnit.SubUnits = foundList
			foundSubUnits = append(foundSubUnits, subUnit)
		}
	}
	if len(foundSubUnits) > 0 {
		return true, foundSubUnits
	}
	if strings.Contains(strings.ToLower(item.Name), searchName) {
		return true, []dictapimodels.DepartmentTreeItem{}
	}
	return false, []dictapimodels.DepartmentTreeItem{}
}
