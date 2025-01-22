package models

import "fmt"

type SpacePushSettingCode string

type PushTpl struct {
	Name  string
	Title string
	Msg   string
}

var PushCodeMap = map[SpacePushSettingCode]PushTpl{
	//Важные уведомления
	PushLicenseExpire: {Name: "Окончание действия лицензии на продукт", Title: "Окончание действия лицензии", Msg: "Ваша лицензия на продукт истекает {{LicEndDate}}. Пожалуйста, продлите действие, чтобы избежать блокировки."}, //TODO для модуля лицензирования

	PushVRClosed:   {Name: "Необходимость в вакансии закончилась", Title: "Необходимость в вакансии закончилась", Msg: "Заявка «%v» завершена. Статус: %v."},
	PushVRApproved: {Name: "Согласование заявки", Title: "Заявка согласована", Msg: "Заявка «%v» была согласована пользователем %v."},
	PushVRRejected: {Name: "Отклонение заявки", Title: "Заявка отклонена", Msg: "Заявка «%v» была отклонена пользователем %v/%v."},

	PushVacancyResponsible: {Name: "Ответственный за вакансию назначен", Title: "Назначен ответственный", Msg: "Теперь за вакансию «%v» отвечает %v."},
	PushVacancyNewStatus:   {Name: "Изменение статуса вакансии", Title: "Изменён статус вакансии", Msg: "Статус вакансии «%v» изменён на %v."},
	PushVacancyPublished:   {Name: "Публикация вакансии на стороннем сайте (HH, Avito)", Title: "Вакансия опубликована на %v", Msg: "Вакансия «%v» успешно опубликована на %v."},

	PushApplicantNegotiation: {Name: "Получение отклика по вакансии", Title: "Новый отклик на вакансию", Msg: "На вакансию «%v» пришёл новый отклик от кандидата %v."},
	PushApplicantNote:        {Name: "Заказчик комментирует кандидата на вакансии, в команде которой вы состоите", Title: "Комментарий от заказчика по кандидату", Msg: "Заказчик %v оставил комментарий к кандидату %v на вакансии «%v»."},
	PushApplicantMsg:         {Name: "Пришло сообщение через Avito/HH", Title: "Новое сообщение от кандидата через %v", Msg: "Получено новое сообщение через %v от кандидата %v по вакансии «%v»."},
	PushApplicantNewStage:    {Name: "Кандидат переведён на этап «Следующий этап»", Title: "Кандидат переведен на следующий этап", Msg: "Кандидат %v переведён на следующий этап «%v» по вакансии «%v»."},
}

const (
	PushLicenseExpire SpacePushSettingCode = "PushLicenseExpire"

	PushVRClosed   SpacePushSettingCode = "PushVRClosed"
	PushVRApproved SpacePushSettingCode = "PushVRApproved"
	PushVRRejected SpacePushSettingCode = "PushVRRejected"

	PushVacancyResponsible SpacePushSettingCode = "PushVacancyResponsible"
	PushVacancyNewStatus   SpacePushSettingCode = "PushVacancyNewStatus"
	PushVacancyPublished   SpacePushSettingCode = "PushVacancyPublished"

	PushApplicantNegotiation SpacePushSettingCode = "PushApplicantNegotiation"
	PushApplicantNote        SpacePushSettingCode = "PushApplicantNote"
	PushApplicantMsg         SpacePushSettingCode = "PushApplicantMsg"//!!
	PushApplicantNewStage    SpacePushSettingCode = "PushApplicantNewStage"
)

type NotificationData struct {
	Code  SpacePushSettingCode
	Msg   string
	Title string
}

func GetPushVRClosed(vacancyName, vrStatus string) NotificationData {
	code:= PushVRClosed
	return NotificationData{
		Code:  code,
		Title: PushCodeMap[code].Title,
		Msg:   fmt.Sprintf(PushCodeMap[code].Msg, vacancyName, vrStatus),
	}
}

func GetPushVRApproved(vacancyName, userName string) NotificationData {
	code:= PushVRApproved
	return NotificationData{
		Code:  code,
		Title: PushCodeMap[code].Title,
		Msg:   fmt.Sprintf(PushCodeMap[code].Msg, vacancyName, userName),
	}
}

func GetPushVRRejected(vacancyName, userName, userRole string) NotificationData {
	code:= PushVRRejected
	return NotificationData{
		Code:  code,
		Title: PushCodeMap[code].Title,
		Msg:   fmt.Sprintf(PushCodeMap[code].Msg, vacancyName, userName, userRole),
	}
}

func GetPushVacancyResponsible(vacancyName, responsibleFullName string) NotificationData {
	code:= PushVacancyResponsible
	return NotificationData{
		Code:  code,
		Title: PushCodeMap[code].Title,
		Msg:   fmt.Sprintf(PushCodeMap[code].Msg, vacancyName, responsibleFullName),
	}
}

func GetPushVacancyNewStatus(vacancyName, vacancyStatus string) NotificationData {
	code:= PushVacancyNewStatus
	return NotificationData{
		Code:  code,
		Title: PushCodeMap[code].Title,
		Msg:   fmt.Sprintf(PushCodeMap[code].Msg, vacancyName, vacancyStatus),
	}
}

func GetPushVacancyPublished(vacancyName, pubService string) NotificationData {
	code:= PushVacancyPublished
	return NotificationData{
		Code:  code,
		Title: fmt.Sprintf(PushCodeMap[code].Title, pubService),
		Msg:   fmt.Sprintf(PushCodeMap[code].Msg, vacancyName, pubService),
	}
}

func GetPushApplicantNegotiation(vacancyName, applicantFullName string) NotificationData {
	code:= PushVacancyPublished
	return NotificationData{
		Code:  code,
		Title: PushCodeMap[code].Title,
		Msg:   fmt.Sprintf(PushCodeMap[code].Msg, vacancyName, applicantFullName),
	}
}

func GetPushApplicantNote(vacancyName, applicantFullName, userFullName string) NotificationData {
	code:= PushApplicantNote
	return NotificationData{
		Code:  code,
		Title: PushCodeMap[code].Title,
		Msg:   fmt.Sprintf(PushCodeMap[code].Msg, userFullName, applicantFullName, vacancyName),
	}
}

func GetPushApplicantMsg(vacancyName, applicantFullName, pubService string) NotificationData {
	code:= PushApplicantMsg
	return NotificationData{
		Code:  code,
		Title: fmt.Sprintf(PushCodeMap[code].Title, pubService),
		Msg:   fmt.Sprintf(PushCodeMap[code].Msg, pubService, applicantFullName, vacancyName),
	}
}

func GetPushApplicantNewStage(vacancyName, userFullName, stageName string) NotificationData {
	code:= PushApplicantNewStage
	return NotificationData{
		Code:  code,
		Title: PushCodeMap[code].Title,
		Msg:   fmt.Sprintf(PushCodeMap[code].Msg, userFullName, stageName, vacancyName),
	}
}