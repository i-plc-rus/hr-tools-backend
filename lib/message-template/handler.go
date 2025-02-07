package messagetemplate

import (
	"bytes"
	"fmt"
	"hr-tools-backend/db"
	applicantstore "hr-tools-backend/lib/applicant/store"
	messagetemplatestore "hr-tools-backend/lib/message-template/store"
	"hr-tools-backend/lib/smtp"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	"hr-tools-backend/models"
	applicantapimodels "hr-tools-backend/models/api/applicant"
	msgtemplateapimodels "hr-tools-backend/models/api/message-template"
	dbmodels "hr-tools-backend/models/db"
	"html/template"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	SendEmailMessage(spaceID, templateID, applicantID, userID string) (hMsg string, err error)
	GetListTemplates(spaceID string) (list []msgtemplateapimodels.MsgTemplateView, err error)
	MultiSendEmail(spaceID, userID string, data applicantapimodels.MultiEmailRequest) (failMails []string, hMsg string, err error)
	Create(spaceID string, request msgtemplateapimodels.MsgTemplateData) (string, error)
	GetByID(spaceID, id string) (msgtemplateapimodels.MsgTemplateView, error)
	Update(spaceID, id string, request msgtemplateapimodels.MsgTemplateData) error
	Delete(spaceID, id string) error
}

var Instance Provider

func NewHandler() {
	Instance = &impl{
		msgTemplateStore:   messagetemplatestore.NewInstance(db.DB),
		applicantStore:     applicantstore.NewInstance(db.DB),
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
		spaceUsersStore:    spaceusersstore.NewInstance(db.DB),
		vacancyStore:       vacancystore.NewInstance(db.DB),
	}
}

type impl struct {
	msgTemplateStore   messagetemplatestore.Provider
	applicantStore     applicantstore.Provider
	spaceSettingsStore spacesettingsstore.Provider
	spaceUsersStore    spaceusersstore.Provider
	vacancyStore       vacancystore.Provider
}

func (i impl) SendEmailMessage(spaceID, templateID, applicantID, userID string) (hMsg string, err error) {
	logger := log.WithFields(log.Fields{
		"space_id":     spaceID,
		"template_id":  templateID,
		"applicant_id": applicantID,
	})

	email, err := i.getSenderEmail(spaceID, logger)
	if err != nil {
		return "", err
	}
	if email == "" {
		return "в настройках пространства не указана почта для отправки", nil
	}

	msgTemplate, err := i.getMsgTemplate(spaceID, templateID, logger)
	if err != nil {
		return "", err
	}
	if msgTemplate == nil {
		return "шаблон сообщения не найден", nil
	}

	applicant, err := i.applicantStore.GetByID(spaceID, applicantID)
	if err != nil {
		return "", err
	}
	if applicant == nil {
		return "", errors.New("не найден кандидат по указанномму ID")
	}

	if applicant.Email == "" {
		return "у кандидата не указана почта", nil
	}
	user, err := i.spaceUsersStore.GetByID(userID)
	if err != nil {
		return "", errors.New("ошибка получения профиля отправителя")
	}
	textSign := ""
	if user != nil {
		textSign = user.TextSign
	}
	tlpData, err := i.getTlpData(applicant.Applicant, applicant.VacancyID, user)
	if err != nil {
		return "", err
	}
	msg, err := buildMsg(msgTemplate.Message, textSign, tlpData)
	if err != nil {
		return "", err
	}
	title, err := buildTitle(msgTemplate.Title, tlpData)
	if err != nil {
		return "", err
	}

	err = smtp.Instance.SendEMail(email, applicant.Email, msg, title)
	if err != nil {
		return "", errors.Wrap(err, "ошибка отправки почты кандидату")
	}
	return "", nil
}

func (i impl) GetListTemplates(spaceID string) (list []msgtemplateapimodels.MsgTemplateView, err error) {
	recList, err := i.msgTemplateStore.List(spaceID)
	if err != nil {
		return nil, err
	}
	for _, template := range recList {
		list = append(list, template.ToModel())
	}
	return list, nil
}

func (i impl) MultiSendEmail(spaceID, userID string, data applicantapimodels.MultiEmailRequest) (failMails []string, hMsg string, err error) {
	logger := log.WithFields(log.Fields{
		"space_id":    spaceID,
		"template_id": data.MsgTemplateID,
		"user_id":     userID,
	})
	if !smtp.Instance.IsConfigured() {
		return nil, "", errors.New("smtp клиент не настроен")
	}

	email, err := i.getSenderEmail(spaceID, logger)
	if err != nil {
		return nil, "", err
	}
	if email == "" {
		return nil, "в настройках пространства не указана почта для отправки", nil
	}

	msgTemplate, err := i.getMsgTemplate(spaceID, data.MsgTemplateID, logger)
	if err != nil {
		return nil, "", err
	}
	if msgTemplate == nil {
		return nil, "шаблон сообщения не найден", nil
	}

	user, err := i.spaceUsersStore.GetByID(userID)
	if err != nil {
		return nil, "", errors.Wrap(err, "ошибка получения профиля отправителя")
	}
	textSign := ""
	if user != nil && user.UsePersonalSign {
		textSign = user.TextSign
	}

	applicantList, err := i.applicantStore.ListOfApplicantByIDs(spaceID, data.IDs, nil)
	if err != nil {
		return nil, "", errors.Wrap(err, "ошибка получения списка кандидатов")
	}
	failMails = []string{}
	for _, applicant := range applicantList {
		fio := applicant.GetFIO()
		if applicant.Email == "" {
			logger.
				WithField("applicant_id", applicant.ID).
				Warn("у кандидата не указана почта")
			failMails = append(failMails, fio)
			continue
		}

		tlpData, err := i.getTlpData(applicant.Applicant, applicant.VacancyID, user)
		if err != nil {
			return nil, "", err
		}
		msg, err := buildMsg(msgTemplate.Message, textSign, tlpData)
		if err != nil {
			return nil, "", err
		}
		title, err := buildTitle(msgTemplate.Title, tlpData)
		if err != nil {
			return nil, "", err
		}
		err = smtp.Instance.SendEMail(email, applicant.Email, msg, title)
		if err != nil {
			logger.
				WithError(err).
				Error("ошибка отправки почты кандидату")
			failMails = append(failMails, fio)
			continue
		}
	}

	return failMails, "", nil
}

func (i impl) Create(spaceID string, request msgtemplateapimodels.MsgTemplateData) (string, error) {
	rec := dbmodels.MessageTemplate{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		Name:         request.Name,
		Title:        request.Title,
		Message:      request.Message,
		TemplateType: request.TemplateType,
	}
	return i.msgTemplateStore.Create(rec)
}

func (i impl) GetByID(spaceID, id string) (msgtemplateapimodels.MsgTemplateView, error) {
	rec, err := i.msgTemplateStore.GetByID(spaceID, id)
	if err != nil {
		return msgtemplateapimodels.MsgTemplateView{}, err
	}
	return rec.ToModel(), nil
}

func (i impl) Update(spaceID, id string, request msgtemplateapimodels.MsgTemplateData) error {
	updMap := map[string]interface{}{
		"name":          request.Name,
		"title":         request.Title,
		"message":       request.Message,
		"template_type": request.TemplateType,
	}
	err := i.msgTemplateStore.Update(id, updMap)
	if err != nil {
		return err
	}
	return nil
}

func (i impl) Delete(spaceID, id string) error {
	return i.msgTemplateStore.Delete(spaceID, id)
}

func GetVariables() []msgtemplateapimodels.TemplateItem {
	return []msgtemplateapimodels.TemplateItem{
		{
			Name:  "Название должности",
			Value: "{{.JobTitle}}",
		},
		{
			Name:  "ФИО кандидата",
			Value: "{{.ApplicantFIO}}",
		},
		{
			Name:  "Имя кандидата",
			Value: "{{.ApplicantName}}",
		},
		{
			Name:  "Фамилия кандидата",
			Value: "{{.ApplicantLastName}}",
		},
		{
			Name:  "Отчество кандидата",
			Value: "{{.ApplicantMiddleName}}",
		},
		{
			Name:  "Название вакансии",
			Value: "{{.VacancyName}}",
		},
		{
			Name:  "Название компании",
			Value: "{{.CompanyName}}",
		},
		{
			Name:  "Источник кандидата",
			Value: "{{.ApplicantSource}}",
		},
		{
			Name:  "Ссылка на вакансию",
			Value: "{{.VacancyLink}}",
		},
		{
			Name:  "Мое имя",
			Value: "{{.SelfName}}",
		},
	}
}

func buildTitle(tmpl string, data tplData) (string, error) {
	tpl, err := template.New("msg_title").Parse(tmpl)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func buildMsg(tmpl string, emailTextSign string, data tplData) (string, error) {
	// возможно в будущем будем искать место вставки подписи и заполнение шаблона,
	// пока подпись добавляем в конец
	// если будеи использвать не только text/plain но и text/html, надо будет менять перенос на <br> для последнего
	tpl, err := template.New("msg_body").Parse(tmpl)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	msg := buf.String()
	if emailTextSign == "" {
		return msg, nil
	}
	return fmt.Sprintf("%v\r\n%v", msg, emailTextSign), nil
}

func (i impl) getSenderEmail(spaceID string, logger *log.Entry) (string, error) {
	email, err := i.spaceSettingsStore.GetValueByCode(spaceID, models.SpaceSenderEmail)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения почты для отправки из настроек пространства")
	}
	return email, nil
}

func (i impl) getMsgTemplate(spaceID, templateID string, logger *log.Entry) (*dbmodels.MessageTemplate, error) {
	msgTemplate, err := i.msgTemplateStore.GetByID(spaceID, templateID)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения шаблона сообщения")
	}
	return msgTemplate, nil
}

func (i impl) getTlpData(applicant dbmodels.Applicant, vacancyID string, user *dbmodels.SpaceUser) (tplData, error) {
	resilt := tplData{
		JobTitle:            "",
		ApplicantFIO:        applicant.GetFIO(),
		ApplicantName:       applicant.FirstName,
		ApplicantLastName:   applicant.LastName,
		ApplicantMiddleName: applicant.MiddleName,
		VacancyName:         "",
		CompanyName:         "",
		ApplicantSource:     string(applicant.Source),
		VacancyLink:         "",
		SelfName:            user.GetFullName(),
	}
	if vacancy, err := i.vacancyStore.GetByID(applicant.SpaceID, vacancyID); err == nil && vacancy != nil {
		resilt.VacancyName = vacancy.VacancyName
		if vacancy.JobTitle != nil {
			resilt.JobTitle = vacancy.JobTitle.Name
		}
		if vacancy.Company != nil {
			resilt.CompanyName = vacancy.Company.Name
		}
		if applicant.Source == models.ApplicantSourceAvito {
			resilt.VacancyLink = vacancy.AvitoUri
		} else if applicant.Source == models.ApplicantSourceHh {
			resilt.VacancyLink = vacancy.HhUri
		}
	}
	return resilt, nil
}

type tplData struct {
	JobTitle            string
	ApplicantFIO        string
	ApplicantName       string
	ApplicantLastName   string
	ApplicantMiddleName string
	VacancyName         string
	CompanyName         string
	ApplicantSource     string
	VacancyLink         string
	SelfName            string
}

//TODO замена переменных шаблона
//TODO построение шаблона
