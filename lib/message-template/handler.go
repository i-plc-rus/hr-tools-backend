package messagetemplate

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"hr-tools-backend/db"
	applicanthistoryhandler "hr-tools-backend/lib/applicant-history"
	applicantstore "hr-tools-backend/lib/applicant/store"
	pdfexport "hr-tools-backend/lib/export/pdf"
	filestorage "hr-tools-backend/lib/file-storage"
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
)

type Provider interface {
	SendEmailMessage(ctx context.Context, spaceID, templateID, applicantID, userID string) (hMsg string, err error)
	GetListTemplates(spaceID string) (list []msgtemplateapimodels.MsgTemplateView, err error)
	MultiSendEmail(ctx context.Context, spaceID, userID string, data applicantapimodels.MultiEmailRequest) (failMails []string, hMsg string, err error)
	Create(spaceID string, request msgtemplateapimodels.MsgTemplateData) (string, error)
	GetByID(spaceID, id string) (msgtemplateapimodels.MsgTemplateView, error)
	Update(spaceID, id string, request msgtemplateapimodels.MsgTemplateData) error
	Delete(spaceID, id string) error
	PdfPreview(ctx context.Context, spaceID, tplID, userID string) (body []byte, hMsg string, err error)
	GetSenderEmail(spaceID string) (string, error)
}

var Instance Provider

func NewHandler() {
	Instance = &impl{
		msgTemplateStore:   messagetemplatestore.NewInstance(db.DB),
		applicantStore:     applicantstore.NewInstance(db.DB),
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
		spaceUsersStore:    spaceusersstore.NewInstance(db.DB),
		vacancyStore:       vacancystore.NewInstance(db.DB),
		fileStorage:        filestorage.Instance,
		applicantHistory:   applicanthistoryhandler.Instance,
	}
}

type impl struct {
	msgTemplateStore   messagetemplatestore.Provider
	applicantStore     applicantstore.Provider
	spaceSettingsStore spacesettingsstore.Provider
	spaceUsersStore    spaceusersstore.Provider
	vacancyStore       vacancystore.Provider
	fileStorage        filestorage.Provider
	applicantHistory   applicanthistoryhandler.Provider
}

func (i impl) SendEmailMessage(ctx context.Context, spaceID, templateID, applicantID, userID string) (hMsg string, err error) {
	logger := log.WithFields(log.Fields{
		"space_id":     spaceID,
		"template_id":  templateID,
		"applicant_id": applicantID,
	})

	email, err := i.GetSenderEmail(spaceID)
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
		return "", errors.Wrap(err, "ошибка получения профиля отправителя")
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
	var attachment *models.File
	if msgTemplate.TemplateType == models.TplOffer {
		body, hMsg, err := i.buildPdf(ctx, spaceID, tlpData, msgTemplate)
		if err != nil {
			return "", err
		}
		if hMsg != "" {
			return hMsg, nil
		}
		attachment = &models.File{
			FileName:    "offer.pdf",
			Body:        body,
			ContentType: "application/pdf",
		}
	}
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		applicantHistory := applicanthistoryhandler.NewTxHandler(tx)
		// История изменений
		changes := applicanthistoryhandler.GetMailSentChange(title)
		applicantHistory.SaveWithUser(spaceID, applicantID, applicant.VacancyID, userID, user.GetFullName(), dbmodels.HistoryTypeReject, changes)
		//Отправка письма
		err = smtp.Instance.SendHtmlEMail(email, applicant.Email, msg, title, attachment)
		if err != nil {
			return errors.Wrap(err, "ошибка отправки почты кандидату")
		}
		return nil
	})
	if err != nil {
		return "", err
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

func (i impl) MultiSendEmail(ctx context.Context, spaceID, userID string, data applicantapimodels.MultiEmailRequest) (failMails []string, hMsg string, err error) {
	logger := log.WithFields(log.Fields{
		"space_id":    spaceID,
		"template_id": data.MsgTemplateID,
		"user_id":     userID,
	})
	if !smtp.Instance.IsConfigured() {
		return nil, "", errors.New("smtp клиент не настроен")
	}

	email, err := i.GetSenderEmail(spaceID)
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
		var attachment *models.File
		if msgTemplate.TemplateType == models.TplOffer {
			body, hMsg, err := i.buildPdf(ctx, spaceID, tlpData, msgTemplate)
			if err != nil {
				return nil, "", err
			}
			if hMsg != "" {
				return nil, hMsg, nil
			}
			attachment = &models.File{
				FileName:    "offer.pdf",
				Body:        body,
				ContentType: "application/pdf",
			}
		}
		err = smtp.Instance.SendHtmlEMail(email, applicant.Email, msg, title, attachment)
		if err != nil {
			logger.
				WithError(err).
				Error("ошибка отправки почты кандидату")
			failMails = append(failMails, fio)
			continue
		}
		// История изменений
		changes := applicanthistoryhandler.GetMailSentChange(title)
		i.applicantHistory.SaveWithUser(spaceID, applicant.ID, applicant.VacancyID, userID, user.GetFullName(), dbmodels.HistoryTypeReject, changes)
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
		PdfMessage:   request.PdfMessage,
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
		"pdf_message":   request.PdfMessage,
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
		{
			Name:  "ФИО руководителя",
			Value: "{{.CompanyDirectorName}}",
		},
		{
			Name:  "Адрес компании",
			Value: "{{.CompanyAddress}}",
		},
		{
			Name:  "[Контактные данные компании",
			Value: "{{.CompanyContact}}",
		},
	}
}

func (i impl) PdfPreview(ctx context.Context, spaceID, tplID, userID string) (body []byte, hMsg string, err error) {
	msgTemplate, err := i.msgTemplateStore.GetByID(spaceID, tplID)
	if err != nil {
		return nil, "", err
	}
	if msgTemplate.TemplateType != models.TplOffer {
		return nil, "шаблон не предусматривает генерацию pdf", nil
	}
	user, err := i.spaceUsersStore.GetByID(userID)
	if err != nil {
		return nil, "", errors.Wrap(err, "ошибка получения профиля")
	}
	if user == nil {
		return nil, "пользователь не найден", nil
	}
	tplData := models.TemplateData{
		JobTitle:            "[Название должности]",
		ApplicantFIO:        "[ФИО кандидата]",
		ApplicantName:       "[Имя кандидата]",
		ApplicantLastName:   "[Фимилия кандидата]",
		ApplicantMiddleName: "[Отчество кандидата]",
		VacancyName:         "[Название вакансии]",
		ApplicantSource:     "[Источник кандидата]",
		VacancyLink:         "[Ссылка на вакансию]",
		SelfName:            user.GetFullName(),
		CompanyAddress:      "[Адрес компании]",
		CompanyContact:      "[Контактные данные компании]",
		CompanyName:         "[Название компании]",
		CompanyDirectorName: "[ФИО директора компании]",
		Files:               models.TemplateFiles{},
	}
	return i.buildPdf(context.TODO(), spaceID, tplData, msgTemplate)

}

func buildTitle(tmpl string, data models.TemplateData) (string, error) {
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

func buildMsg(tmpl string, emailTextSign string, data models.TemplateData) (string, error) {
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
	msg = fmt.Sprintf("<div>%v</div>", msg)
	if emailTextSign == "" {
		return msg, nil
	}
	return fmt.Sprintf("%v<div>%v</div>", msg, emailTextSign), nil
}

func (i impl) buildPdf(ctx context.Context, spaceID string, tplData models.TemplateData, tplRec *dbmodels.MessageTemplate) (body []byte, hMsg string, err error) {
	if tplRec.PdfMessage == "" {
		return nil, "не указан текст шаблона для pdf", nil
	}
	tplData.Files.Logo, err = i.getFile(ctx, spaceID, tplRec.ID, dbmodels.CompanyLogo)
	if err != nil {
		return nil, "", errors.Wrap(err, "ошибка получения изображения логотипа")
	}

	tplData.Files.Sign, err = i.getFile(ctx, spaceID, tplRec.ID, dbmodels.CompanySign)
	if err != nil {
		return nil, "", errors.Wrap(err, "ошибка получения изображения с подписью")
	}

	tplData.Files.Stamp, err = i.getFile(ctx, spaceID, tplRec.ID, dbmodels.CompanyStamp)
	if err != nil {
		return nil, "", errors.Wrap(err, "ошибка получения изображения со штампом")
	}
	body, err = pdfexport.GenerateOffer(tplRec.PdfMessage, tplData)
	return body, "", err
}

func (i impl) GetSenderEmail(spaceID string) (string, error) {
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

func (i impl) getTlpData(applicant dbmodels.Applicant, vacancyID string, user *dbmodels.SpaceUser) (models.TemplateData, error) {
	result := models.TemplateData{
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
		CompanyDirectorName: "",
		CompanyAddress:      "",
		CompanyContact:      "",
	}
	vacancy, err := i.vacancyStore.GetByID(applicant.SpaceID, vacancyID)
	if err != nil {
		return models.TemplateData{}, err
	}
	if vacancy == nil {
		return models.TemplateData{}, errors.New("вакансия не найдена")
	}
	result.VacancyName = vacancy.VacancyName
	if vacancy.JobTitle != nil {
		result.JobTitle = vacancy.JobTitle.Name
	}
	if vacancy.Company != nil {
		result.CompanyName = vacancy.Company.Name
		result.CompanyDirectorName = vacancy.Space.DirectorName
		result.CompanyAddress = vacancy.Space.CompanyAddress
		result.CompanyContact = vacancy.Space.CompanyContact
	}
	if applicant.Source == models.ApplicantSourceAvito {
		result.VacancyLink = vacancy.AvitoUri
	} else if applicant.Source == models.ApplicantSourceHh {
		result.VacancyLink = vacancy.HhUri
	}

	return result, nil
}

func (i impl) getFile(ctx context.Context, spaceID, tplID string, fileType dbmodels.FileType) (*models.File, error) {
	body, fileData, err := i.fileStorage.GetFileByType(ctx, spaceID, tplID, fileType)
	if err != nil {
		return nil, err
	}
	if fileData != nil && body != nil {
		return &models.File{
			FileName:    fileData.Name,
			ContentType: fileData.ContentType,
			Body:        body,
		}, nil
	}
	return nil, nil
}
