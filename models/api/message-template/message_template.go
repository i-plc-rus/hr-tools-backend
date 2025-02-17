package msgtemplateapimodels

import (
	"hr-tools-backend/models"
	"html/template"

	"strings"

	"github.com/pkg/errors"
)

type MsgTemplateData struct {
	Name         string              `json:"name"`          // Название шаблона
	Title        string              `json:"title"`         // Тема/заголовок письма с переменными шаблона (Пример шаблона: "Информация от {{.CompanyName}}")
	Message      string              `json:"message"`       // Текст шаблона с переменными шаблона (Пример шаблона: "Вакансия {{.VacancyName}} более не актуальна, приходи еще, {{.SelfName}}")
	TemplateType models.TemplateType `json:"template_type"` // Тип шаблона
	PdfMessage   string              `json:"pdf_message"`   // Текст шаблона оффера для генерации pdf с переменными шаблона (Пример шаблона: "<center>Ваш оффер!</center>Тут кокой то текст оффера с разными стилями: <b>bold</b>, <i>italic</i>, <u>underlined</u>, or <b><i><u>all at once</u></i></b>!<br><br><right>С уважением</right><right>Директор {{.CompanyName}}</right><right>{{.CompanyDirectorName}}</right>")
}

type TemplateItem struct {
	Name  string `json:"name"`  // Значение для отображения пользователю
	Value string `json:"value"` // Переменная шаблона
}

func (t MsgTemplateData) Validate() error {
	if t.Name == "" {
		return errors.New("не указано название шаблона")
	}
	if t.TemplateType == "" {
		return errors.New("не указан тип шаблона")
	}
	if !t.TemplateType.IsValid() {
		return errors.New("тип шаблона не поддерживается")
	}
	if t.Message == "" {
		return errors.New("не указан текст шаблона")
	}
	_, err := template.New("validate").Parse(t.Message)
	if err != nil {
		return errors.New("текст шаблона содержит ошибки")
	}
	if t.Title != "" {
		_, err := template.New("validate").Parse(t.Title)
		if err != nil {
			return errors.New("Тема/заголовок шаблона содержит ошибки")
		}
	}
	if t.TemplateType == models.TplOffer {
		if t.PdfMessage == "" {
			return errors.New("не указан текст pdf шаблона для оффера")
		}
	}
	return nil
}

type MsgTemplateView struct {
	MsgTemplateData
	ID string `json:"id"` // Идентификатор шаблона
}

type SendMessage struct {
	ApplicantID   string `json:"applicant_id"`    // ID кандидата/отклика кому отправить сообщение
	MsgTemplateID string `json:"msg_template_id"` // ID шаблона сообщения, которое нужно отправить
}

func (r SendMessage) Validate() error {
	if len(strings.TrimSpace(r.ApplicantID)) == 0 {
		return errors.New("не указан кандидат")
	}
	if len(strings.TrimSpace(r.MsgTemplateID)) == 0 {
		return errors.New("не указан шаблон сообщения")
	}
	return nil
}

type OfferTemplateData struct {
	Title      string `json:"title"`       // Тема/заголовок письма с переменными шаблона (Пример шаблона: "Информация от {{.CompanyName}}")
	PdfMessage string `json:"pdf_message"` // Текст шаблона оффера с переменными шаблона (Пример шаблона: "<center>Ваш оффер!</center>Тут кокой то текст оффера с разными стилями: <b>bold</b>, <i>italic</i>, <u>underlined</u>, or <b><i><u>all at once</u></i></b>!<br><br><right>С уважением</right><right>Директор {{.CompanyName}}</right><right>{{.CompanyDirectorName}}</right>")
	Message    string `json:"message"`     // Текст шаблона для сопроводительного письма с переменными шаблона (Пример шаблона: "Вам направлен оффер по вакансии {{.VacancyName}}, <br><p align="right">С уважением {{.SelfName}}</p>")
}
