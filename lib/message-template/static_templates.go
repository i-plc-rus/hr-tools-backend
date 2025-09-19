package messagetemplate

import (
	"bytes"
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	dbmodels "hr-tools-backend/models/db"
	"html/template"
	"os"
	"strings"
)

const (
	licenceRenewTitle  = "Продление лицензии"
	survaySuggestTitle = "Пройти тестирование"
)

func BuildLicenceRenewMsg(text string, user dbmodels.SpaceUser, space dbmodels.Space) (string, error) {
	tpl, err := getTemplate("static/sales_licence_renew.html", true)
	if err != nil {
		return "", err
	}
	data := models.SalesTemplateData{
		OrganizationName: space.OrganizationName,
		Inn:              space.Inn,
		Kpp:              space.Kpp,
		OGRN:             space.OGRN,
		DirectorName:     space.DirectorName,
		UserFIO:          user.GetFullName(),
		UserPhoneNumber:  user.PhoneNumber,
		TextRequest:      text,
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GetLicenceRenewTitle() string {
	return licenceRenewTitle
}

func GetSurvaySuggestMessage(companyName, link string, isHtml bool) (msg string, err error) {
	var tpl *template.Template
	if isHtml {
		filePath := "static/applicant_survey_suggest.html"
		tpl, err = getTemplate(filePath, isHtml)
	} else {
		filePath := "static/applicant_survey_suggest.txt"
		tpl, err = getTemplate(filePath, isHtml)
	}
	if err != nil {
		return "", err
	}
	data := models.SurvaySuggestTemplateData{
		CompanyName: companyName,
		SurvayLink:  link,
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GetSurvaySuggestTitle() string {
	return survaySuggestTitle
}

func getTemplate(filePath string, isHtml bool) (*template.Template, error) {
	tmplBody, err := getTplFile(filePath)
	if err != nil {
		return nil, err
	}
	var body string
	if isHtml {
		body = strings.Replace(string(tmplBody), "\n", "", -1)
	} else {
		body = string(tmplBody)
	}

	tpl, err := template.New("msg_body").Parse(body)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

func getTplFile(filePath string) ([]byte, error) {
	body, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "ошибка чтения файла шаблона %v", filePath)
	}
	return body, nil
}

func GetSurvayStep0SuggestMessage(companyName, link string, isHtml bool) (msg string, err error) {
	var tpl *template.Template
	if isHtml {
		filePath := "static/applicant_survey_suggest.html" //TODO добавить шаблон для шага 0
		tpl, err = getTemplate(filePath, isHtml)
	} else {
		filePath := "static/applicant_survey_suggest.txt" //TODO добавить шаблон для шага 0
		tpl, err = getTemplate(filePath, isHtml)
	}
	if err != nil {
		return "", err
	}
	data := models.SurvaySuggestTemplateData{
		CompanyName: companyName,
		SurvayLink:  link,
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
