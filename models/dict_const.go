package models

import "github.com/pkg/errors"

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

type VRStatus string

const (
	VRStatusCreated       VRStatus = "Создана"
	VRStatusCanceled      VRStatus = "Отменена"
	VRStatusNotAccepted   VRStatus = "Не согласована"
	VRStatusAccepted      VRStatus = "Согласована"
	VRStatusUnderRevision VRStatus = "На доработке"
	VRStatusUnderAccepted VRStatus = "На согласовании"
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

type Experience string

const (
	ExperienceNoMatter   Experience = "noMatter"   // Без опыта
	ExperienceMoreThan1  Experience = "moreThan1"  // Более 1 года
	ExperienceMoreThan3  Experience = "moreThan3"  // Более 3 лет
	ExperienceMoreThan5  Experience = "moreThan5"  // Более 5 лет
	ExperienceMoreThan10 Experience = "moreThan10" // Более 10 лет
)

type Schedule string

const (
	ScheduleFlyInFlyOut Schedule = "flyInFlyOut" // Вахта
	SchedulePartTime    Schedule = "partTime"    // Неполный день
	ScheduleFullDay     Schedule = "fullDay"     // Полный день
	ScheduleFlexible    Schedule = "flexible"    // Гибкий
	ScheduleShift       Schedule = "shift"       // Сменный
)

type ApplicantStatus string

const (
	ApplicantStatusInProcess   ApplicantStatus = "В процессе"
	ApplicantStatusRejected    ApplicantStatus = "Отклонен"
	ApplicantStatusNegotiation ApplicantStatus = "Отклик"
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
)

type RelocationType string

const (
	RelocationTypeNo   RelocationType = "no"       // "не могу переехать"
	RelocationTypeYes  RelocationType = "possible" // "могу переехать"
	RelocationTypeWant RelocationType = "want"     // "хочу переехать"
)

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

type TripReadinessType string

const (
	TripReadinessReady     TripReadinessType = "ready"     //готов к командировкам
	TripReadinessSometimes TripReadinessType = "sometimes" //"готов к редким командировкам
	TripReadinessNever     TripReadinessType = "never"     //"готов к редким командировкам
)

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
