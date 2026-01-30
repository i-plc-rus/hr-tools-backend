package dbmodels

type PromptExecution struct {
	BaseModel
	SysPromt   string                `comment:"System промт"`
	UserPromt  string                `comment:"User промт"`
	Answer     string                `comment:"Ответ ИИ"`
	ReqestType PromptType            `gorm:"type:varchar(255)" comment:"Тип запроса к ИИ"`
	Status     PromptExecutionStatus `gorm:"type:varchar(255)" comment:"Статус запроса"`
}

type PromptExecutionStatus string

const (
	PromptExecutionSent     PromptExecutionStatus = "sent"
	PromptExecutionResponse PromptExecutionStatus = "response"
	PromptExecutionError    PromptExecutionStatus = "error"
)

type PromptType string

const (
	PromptTypeQuestions PromptType = "Questions"
	PromptTypeCheck     PromptType = "PromptCheck"
)
