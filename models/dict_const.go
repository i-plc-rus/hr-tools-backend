package models

import (
	"github.com/pkg/errors"
)

type VacancyStatus string

const (
	VacancyStatusOpened    VacancyStatus = "Открыта"
	VacancyStatusCanceled  VacancyStatus = "Отменена"
	VacancyStatusSuspended VacancyStatus = "Приостановлена"
	VacancyStatusClosed    VacancyStatus = "Закрыта"
)

func (v VacancyStatus) IsClosed() bool {
	return v == VacancyStatusCanceled || v == VacancyStatusClosed
}

type VRStatus string

const (
	VRStatusDraft      VRStatus = "DRAFT"
	VRStatusCreated    VRStatus = "CREATED"
	VRStatusInApproval VRStatus = "IN_APPROVAL"
	VRStatusApproved   VRStatus = "APPROVED"
	VRStatusRejected   VRStatus = "REJECTED"
	VRStatusInHr       VRStatus = "IN_HR"
	VRStatusCancelled  VRStatus = "CANCELLED"
	VRStatusDone       VRStatus = "DONE"
)

func (v VRStatus) IsAllowChange(newStatus VRStatus) bool {
	if v == newStatus {
		return true
	}

	switch newStatus {
	case VRStatusCreated:
		return v == VRStatusInApproval || v == VRStatusDraft || v == VRStatusRejected
	case VRStatusInApproval:
		return v == VRStatusCreated || v == VRStatusRejected || v == VRStatusDraft
	case VRStatusApproved:
		return v == VRStatusInApproval
	case VRStatusRejected:
		return v == VRStatusCreated || v == VRStatusInApproval || v == VRStatusApproved
	case VRStatusInHr:
		return v == VRStatusApproved
	case VRStatusCancelled:
		return true
	case VRStatusDone:
		return v == VRStatusInHr
	case VRStatusDraft:
		return v == VRStatusInApproval || v == VRStatusCreated
	}
	return false
}

func (v VRStatus) AllowAccept() bool {
	return v == VRStatusInApproval
}

func (v VRStatus) AllowReject() bool {
	return v == VRStatusInApproval
}

type ApprovalState string

const (
	AStatePending        ApprovalState = "PENDING"
	AStateApproved       ApprovalState = "APPROVED"
	AStateRequestChanges ApprovalState = "REQUEST_CHANGES"
	AStateRejected       ApprovalState = "REJECTED"
	AStateRemoved        ApprovalState = "REMOVED"
)

type VacancyPubStatus string

const (
	VacancyPubStatusNone       VacancyPubStatus = "Не размещена"
	VacancyPubStatusModeration VacancyPubStatus = "Публикуется"
	VacancyPubStatusPublished  VacancyPubStatus = "Опубликована"
	VacancyPubStatusRejected   VacancyPubStatus = "Отклонена"
	VacancyPubStatusClosed     VacancyPubStatus = "Закрыта"
	VacancyPubStatusDraft      VacancyPubStatus = "Черновик"
)

type ApplicantStatus string

const (
	ApplicantStatusInProcess   ApplicantStatus = "В процессе"
	ApplicantStatusRejected    ApplicantStatus = "Отклонен"
	ApplicantStatusNegotiation ApplicantStatus = "Отклик"
	ApplicantStatusArchive     ApplicantStatus = "Архивный"
)

type NegotiationStatus string

const (
	NegotiationStatusWait     NegotiationStatus = "Рассмотреть позже"
	NegotiationStatusRejected NegotiationStatus = "Отклонен"
	NegotiationStatusAccepted NegotiationStatus = "Подходит"
)

type ApplicantSource string

const (
	ApplicantSourceManual ApplicantSource = "Ручной ввод"
	ApplicantSourceAvito  ApplicantSource = "Avito"
	ApplicantSourceHh     ApplicantSource = "HeadHunter"
	ApplicantSourceEmail  ApplicantSource = "Электронная почта"
	ApplicantSourceSoc    ApplicantSource = "Социальные сети"
	ApplicantSite         ApplicantSource = "Карьерный сайт"
)

type RelocationType string

const (
	RelocationTypeNo   RelocationType = "no"       // "не могу переехать"
	RelocationTypeYes  RelocationType = "possible" // "могу переехать"
	RelocationTypeWant RelocationType = "want"     // "хочу переехать"
)

func (r RelocationType) ToString() string {
	switch r {
	case RelocationTypeNo:
		return "не могу переехать"
	case RelocationTypeYes:
		return "могу переехать"
	case RelocationTypeWant:
		return "хочу переехать"
	}
	return ""
}

type EducationType string

const (
	EducationTypeSecondary        EducationType = "secondary"         //"Среднее"
	EducationTypeSpecialSecondary EducationType = "special_secondary" //"Среднее специальное"
	EducationTypeUnfinishedHigher EducationType = "unfinished_higher" //"Неоконченное высшее"
	EducationTypeHigher           EducationType = "higher"            //"Высшее"
	EducationTypeBachelor         EducationType = "bachelor"          //"Бакалавр"
	EducationTypeMaster           EducationType = "master"            //"Магистр"
	EducationTypeCandidate        EducationType = "candidate"         //"Кандидат наук"
	EducationTypeDoctor           EducationType = "doctor"            //"Доктор наук"
)

func (e EducationType) ToString() string {
	switch e {
	case EducationTypeSecondary:
		return "Среднее"
	case EducationTypeSpecialSecondary:
		return "реднее специальное"
	case EducationTypeUnfinishedHigher:
		return "Неоконченное высшее"
	case EducationTypeHigher:
		return "Высшее"
	case EducationTypeBachelor:
		return "Бакалавр"
	case EducationTypeMaster:
		return "Магистр"
	case EducationTypeCandidate:
		return "Кандидат наук"
	case EducationTypeDoctor:
		return "Доктор наук"
	}
	return ""
}

type ExperienceType string

const (
	ExperienceTypeNo           ExperienceType = "No"           //"Нет опыта"
	ExperienceTypeBetween1And3 ExperienceType = "Between1And3" //"От 1 года до 3 лет"
	ExperienceTypeBetween3And6 ExperienceType = "Between3And6" //"От 3 года до 6 лет"
	ExperienceTypeMoreThan6    ExperienceType = "MoreThan6"    //"Более 6 лет"
)

type ResponsePeriodType string

const (
	ResponsePeriodType3days     ResponsePeriodType = "до 3 дней"
	ResponsePeriodType7days     ResponsePeriodType = "до 7 дней"
	ResponsePeriodType7toMonth  ResponsePeriodType = "от 7 дней до 30 дней"
	ResponsePeriodTypeMoreMonth ResponsePeriodType = "более месяца"
)

type LanguageLevelType string

const (
	LanguageLevelA1 LanguageLevelType = "a1"
	LanguageLevelA2 LanguageLevelType = "a2"
	LanguageLevelB1 LanguageLevelType = "b1"
	LanguageLevelB2 LanguageLevelType = "b2"
	LanguageLevelC1 LanguageLevelType = "c1"
	LanguageLevelC2 LanguageLevelType = "c2"
	LanguageLevelL1 LanguageLevelType = "l1"
)

type GenderType string

const (
	GenderTypeM GenderType = "male"   // мужской
	GenderTypeF GenderType = "female" // женский
)

func (g GenderType) ToString() string {
	switch g {
	case GenderTypeM:
		return "мужской"
	case GenderTypeF:
		return "женский"
	}
	return ""
}

type TripReadinessType string

const (
	TripReadinessReady     TripReadinessType = "ready"     //готов к командировкам
	TripReadinessSometimes TripReadinessType = "sometimes" //"готов к редким командировкам
	TripReadinessNever     TripReadinessType = "never"     //"готов к редким командировкам
)

func (t TripReadinessType) ToString() string {
	switch t {
	case TripReadinessReady:
		return "готов к командировкам"
	case TripReadinessSometimes:
		return "готов к редким командировкам"
	case TripReadinessNever:
		return "не готов к командировкам"
	}
	return ""
}

type DriverLicenseType string

const (
	DriverLicenseA  DriverLicenseType = "A"
	DriverLicenseB  DriverLicenseType = "B"
	DriverLicenseC  DriverLicenseType = "C"
	DriverLicenseD  DriverLicenseType = "D"
	DriverLicenseE  DriverLicenseType = "E"
	DriverLicenseBE DriverLicenseType = "BE"
	DriverLicenseCE DriverLicenseType = "CE"
	DriverLicenseDE DriverLicenseType = "DE"
	DriverLicenseTM DriverLicenseType = "TM"
	DriverLicenseTB DriverLicenseType = "TB"
)

type SearchStatusType string

const (
	SearchStatusActive           SearchStatusType = "active_search"       //Активно ищет работу
	SearchStatusLookingForOffers SearchStatusType = "looking_for_offers"  //Рассматривает предложения
	SearchStatusNotLookingForJob SearchStatusType = "not_looking_for_job" //Не ищет работу
	SearchStatusHasJobOffer      SearchStatusType = "has_job_offer"       //Предложили работу, решает
	SearchStatusAcceptedJobOffer SearchStatusType = "accepted_job_offer"  //Вышел на новое место
)

func (s SearchStatusType) ToString() string {
	switch s {
	case SearchStatusActive:
		return "Активно ищет работу"
	case SearchStatusLookingForOffers:
		return "Рассматривает предложения"
	case SearchStatusNotLookingForJob:
		return "Не ищет работу"
	case SearchStatusHasJobOffer:
		return "Предложили работу, решает"
	case SearchStatusAcceptedJobOffer:
		return "Вышел на новое место"
	}
	return ""
}

type SearchLabelType string

const (
	SearchLabelPhoto  SearchLabelType = "only_with_photo"  //Только с фотографией
	SearchLabelSalary SearchLabelType = "only_with_salary" //Не показывать резюме без зарплаты
	SearchLabelAge    SearchLabelType = "only_with_age"    //Не показывать резюме без указания возраста
	SearchLabelGender SearchLabelType = "only_with_gender" //Не показывать резюме без указания пола
)

type LimitType string

const (
	LimitTypeMin  LimitType = "Минут"
	LimitTypeHour LimitType = "Часов"
	LimitTypDay   LimitType = "Дней"
	LimitTypeWeek LimitType = "Недель"
)

type ApAddedPeriodType string

const (
	ApAddedPeriodTypeTDay  ApAddedPeriodType = "За сегодня"
	ApAddedPeriodTypeYDay  ApAddedPeriodType = "За вчера"
	ApAddedPeriodType7days ApAddedPeriodType = "За последние 7 дней"
	ApAddedPeriodTypeMonth ApAddedPeriodType = "За последний месяц"
	ApAddedPeriodTypeYear  ApAddedPeriodType = "За последний год"
)

type AddedType string

const (
	AddedTypeAdded       AddedType = "Добавлен"
	AddedTypeNegotiation AddedType = "Откликнулся"
)

type DuplicateType string

const (
	DuplicateTypeByAuthor   DuplicateType = "ByAuthor"
	DuplicateTypeByContacts DuplicateType = "ByContacts"
)

type RejectInitiator string

const (
	HrReject        RejectInitiator = "Рекрутер"
	HeadReject      RejectInitiator = "Руководитель"
	ApplicantReject RejectInitiator = "Кандидат"
)

func (initiator RejectInitiator) IsValid() error {
	if initiator != HrReject &&
		initiator != HeadReject &&
		initiator != ApplicantReject {
		return errors.New("инициатор отказа указан неверно")
	}
	return nil
}

type TemplateType string

const (
	TplMail          TemplateType = "Письмо"
	TplApplicantNote TemplateType = "Комментарий к кандидату"
	TplRejectNote    TemplateType = "Комментарий к отказу"
	TplReminder      TemplateType = "Напоминание"
	TplRatingNote    TemplateType = "Комментарий к оценке"
	TplSms           TemplateType = "SMS"
	TplOffer         TemplateType = "Оффер"
)

var tplMap = map[TemplateType]bool{
	TplMail:          true,
	TplApplicantNote: true,
	TplRejectNote:    true,
	TplReminder:      true,
	TplRatingNote:    true,
	TplSms:           true,
	TplOffer:         true,
}

func (t TemplateType) IsValid() bool {
	_, ok := tplMap[t]
	return ok
}

type MessengerType string

const (
	MessengerTypeJob     MessengerType = "job"
	MessengerTypeSMS     MessengerType = "sms"
	MessengerTypeWhatsUp MessengerType = "whatsup"
)

type LicenseStatus string

const (
	LicenseStatusActive      LicenseStatus = "ACTIVE"
	LicenseStatusExpiresSoon LicenseStatus = "EXPIRES_SOON"
	LicenseStatusExpired     LicenseStatus = "EXPIRED"
	LicenseStatusGrace       LicenseStatus = "GRACE"
)

func (s LicenseStatus) IdReadOnly() bool {
	return s == LicenseStatusExpired || s == LicenseStatusGrace
}

type LicensePaymentStatus string

const (
	LicensePaymentStatusPending LicensePaymentStatus = "PENDING"
	LicensePaymentStatusPaid    LicensePaymentStatus = "PAID"
	LicensePaymentStatusFailed  LicensePaymentStatus = "FAILED"
)

type VideoInterviewStatus string

const (
	VideoInterviewStatusAbsent     VideoInterviewStatus = "ABSENT"
	VideoInterviewStatusUploading  VideoInterviewStatus = "UPLOADING"
	VideoInterviewStatusProcessing VideoInterviewStatus = "PROCESSING"
	VideoInterviewStatusReady      VideoInterviewStatus = "READY"
	VideoInterviewStatusError      VideoInterviewStatus = "ERROR"
)
