package dbmodels

type AiLog struct {
	BaseSpaceModel
	SysPromt   string       `comment:"System промт"`
	UserPromt  string       `comment:"User промт"`
	Answer     string       `comment:"Ответ ИИ"`
	VacancyID  string       `gorm:"type:varchar(36)" comment:"Идентификатор вакансии"`
	ReqestType AiReqestType `gorm:"type:varchar(255)" comment:"Тип запроса к ИИ"`
	AiName     AiName       `gorm:"type:varchar(255)" comment:"Название ИИ"`
}

type AiName string

const (
	AiYaGptType AiName = "yandexgpt"
)

type AiReqestType string

const (
	AiVacancyDescriptionType AiReqestType = "VacancyDescription"
	AiHRSurveyType           AiReqestType = "HRSurvey"
	AiRegenHRSurveyType      AiReqestType = "RegenHRSurvey"
	AiApplicantSurveyType    AiReqestType = "ApplicantSurvey"
	AiScoreApplicantType     AiReqestType = "ScoreApplicant"
)
