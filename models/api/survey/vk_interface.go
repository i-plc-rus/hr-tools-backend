package surveyapimodels

import ()

type VkAiProvider interface {
	VkStep1(spaceID, vacancyID string, aiData AiData) (resp VkStep1, err error)
	VkStep1Regen(spaceID, vacancyID string, aiData AiData) (newQuestions []VkStep1Question, comments map[string]string, err error)
	VkStep9Score(aiData SemanticData) (scoreResult VkStep9ScoreResult, err error)
	VkStep11Report(spaceID, vacancyID string, aiData ReportRequestData) (reportResult ReportResult, err error)
}
