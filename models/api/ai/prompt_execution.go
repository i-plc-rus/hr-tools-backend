package aiapimodels

import dbmodels "hr-tools-backend/models/db"

type ExecutionResult struct {
	SysPromt   string                         `json:"sys_promt"`
	UserPromt  string                         `json:"user_promt"`
	Answer     string                         `json:"answer"`
	ParsedData any                            `json:"parsed_data"`
	ReqestType dbmodels.PromptType            `json:"reqest_type"`
	Status     dbmodels.PromptExecutionStatus `json:"status"`
}
