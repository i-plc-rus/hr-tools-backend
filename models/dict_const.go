package models

import (
	"github.com/pkg/errors"
)

type VRUrgency string

const (
	VRTypeUrgent    VRUrgency = "Срочно"
	VRTypeNonUrgent VRUrgency = "В плановом порядке"
)

func (v VRUrgency) Validate() error {
	if v == "" {
		return errors.New("параметр срочности не указан")
	}
	if v != VRTypeUrgent && v != VRTypeNonUrgent {
		return errors.New("параметр срочности указан некорректно")
	}
	return nil
}

type VRType string

const (
	VRTypeNew     VRType = "Новая позиция"
	VRTypeReplace VRType = "Замена"
)

func (v VRType) Validate() error {
	if v == "" {
		return errors.New("тип вакансии не указан")
	}
	if v != VRTypeNew && v != VRTypeReplace {
		return errors.New("типа вакансии указан некорректно")
	}
	return nil
}

type VRSelectionType string

const (
	VRSelectionTypeMass     VRSelectionType = "Массовый"
	VRSelectionTypePersonal VRSelectionType = "Индивидуальный"
)

func (v VRSelectionType) Validate() error {
	if v == "" {
		return errors.New("вид подбора не указан")
	}
	if v != VRSelectionTypeMass && v != VRSelectionTypePersonal {
		return errors.New("вид подбора указан некорректно")
	}
	return nil
}

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
	VRStatusCreated       VRStatus = "Создана"
	VRStatusCanceled      VRStatus = "Отменена"
	VRStatusNotAccepted   VRStatus = "Не согласована"
	VRStatusAccepted      VRStatus = "Согласована"
	VRStatusUnderRevision VRStatus = "На доработке"
	VRStatusUnderAccepted VRStatus = "На согласовании"
	VRStatusTemplate      VRStatus = "Шаблон"
)

func (v VRStatus) IsAllowChange(newStatus VRStatus) bool {
	if v == newStatus {
		return true
	}
	switch newStatus {
	case VRStatusCanceled:
		return true
	case VRStatusNotAccepted:
		return v == VRStatusCreated || v == VRStatusUnderRevision || v == VRStatusUnderAccepted
	case VRStatusAccepted:
		return v == VRStatusUnderAccepted
	case VRStatusUnderRevision:
		return v == VRStatusUnderAccepted || v == VRStatusNotAccepted
	case VRStatusUnderAccepted:
		return v == VRStatusCreated || v == VRStatusUnderRevision
	case VRStatusCreated:
		return v == VRStatusTemplate
	}
	return false
}

func (v VRStatus) AllowAccept() bool {
	return v == VRStatusUnderAccepted || v == VRStatusAccepted
}

func (v VRStatus) AllowReject() bool {
	return v == VRStatusUnderAccepted
}

type ApprovalStatus string

const (
	AStatusApproved ApprovalStatus = "Согласованно"
	AStatusRejected ApprovalStatus = "Не согласованно"
	AStatusAwaiting ApprovalStatus = "Ждет согласования"
)

type VacancyPubStatus string

const (
	VacancyPubStatusNone       VacancyPubStatus = "Не размещена"
	VacancyPubStatusModeration VacancyPubStatus = "Публикуется"
	VacancyPubStatusPublished  VacancyPubStatus = "Опубликована"
	VacancyPubStatusRejected   VacancyPubStatus = "Отклонена"
	VacancyPubStatusClosed     VacancyPubStatus = "Закрыта"
)

type Employment string

const (
	EmploymentTemporary  Employment = "temporary"  //Временная
	EmploymentFull       Employment = "full"       //Полная
	EmploymentInternship Employment = "internship" //Стажировка
	EmploymentPartial    Employment = "partial"    //Частичная
	EmploymentVolunteer  Employment = "volunteer"  //Волонтерство
	EmploymentProbation  Employment = "probation"  //Стажировка
)

func (e Employment) ToString() string {
	switch e {
	case EmploymentTemporary:
		return "Временная"
	case EmploymentFull:
		return "Полная"
	case EmploymentInternship:
		return "Стажировка"
	case EmploymentPartial:
		return "Частичная"
	case EmploymentVolunteer:
		return "Волонтерство"
	case EmploymentProbation:
		return "Стажировка"
	}
	return ""
}

func (e Employment) ToHHEmploymentForm() string {
	switch e {
	case EmploymentTemporary:
		return "PROJECT"
	case EmploymentFull:
		return "FULL"
	case EmploymentPartial:
		return "PART"
	}
	return ""
}

func EmploymentSlice() []string {
	return []string{"Временная", "Полная", "Стажировка", "Частичная", "Волонтерство"}
}

type Experience string

const (
	ExperienceNoMatter   Experience = "noMatter"   // Без опыта
	ExperienceMoreThan1  Experience = "moreThan1"  // Более 1 года
	ExperienceMoreThan3  Experience = "moreThan3"  // Более 3 лет
	ExperienceMoreThan5  Experience = "moreThan5"  // Более 5 лет
	ExperienceMoreThan10 Experience = "moreThan10" // Более 10 лет
)

func (s Experience) ToString() string {
	switch s {
	case ExperienceNoMatter:
		return "Без опыта"
	case ExperienceMoreThan1:
		return "Более 1 года"
	case ExperienceMoreThan3:
		return "Более 3 лет"
	case ExperienceMoreThan5:
		return "Более 5 лет"
	case ExperienceMoreThan10:
		return "Более 10 лет"
	}
	return ""
}
func ExperienceFromDescr(s string) Experience {
	switch s {
	case "Без опыта":
		return ExperienceNoMatter
	case "Более 1 года":
		return ExperienceMoreThan1
	case "Более 3 лет":
		return ExperienceMoreThan3
	case "Более 5 лет":
		return ExperienceMoreThan5
	case "Более 10 лет":
		return ExperienceMoreThan10
	}
	return ""
}

func (s Experience) ToPoint() int {
	switch s {
	case ExperienceNoMatter:
		return 1
	case ExperienceMoreThan1:
		return 2
	case ExperienceMoreThan3:
		return 3
	case ExperienceMoreThan5:
		return 4
	case ExperienceMoreThan10:
		return 5
	}
	return 0
}

func ExperienceSlice() []string {
	return []string{"Без опыта", "Более 1 года", "Более 3 лет", "Более 5 лет", "Более 10 лет"}
}

type Schedule string

const (
	ScheduleFlyInFlyOut Schedule = "flyInFlyOut" // Вахта
	SchedulePartTime    Schedule = "partTime"    // Неполный день
	ScheduleFullDay     Schedule = "fullDay"     // Полный день
	ScheduleFlexible    Schedule = "flexible"    // Гибкий
	ScheduleShift       Schedule = "shift"       // Сменный
)

func (s Schedule) ToString() string {
	switch s {
	case ScheduleFlyInFlyOut:
		return "Вахта"
	case SchedulePartTime:
		return "Неполный день"
	case ScheduleFullDay:
		return "Полный день"
	case ScheduleFlexible:
		return "Гибкий"
	case ScheduleShift:
		return "Сменный"
	}
	return ""
}

func ScheduleSlice() []string {
	return []string{"Вахта", "Неполный день", "Полный день", "Гибкий", "Сменный"}
}

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
