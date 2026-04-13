package aimodels

import surveyapimodels "hr-tools-backend/models/api/survey"

type Vk1QuestionResult struct {
	Questions []surveyapimodels.VkStep1Question
	Comments  map[string]string
}

type Vk1IntroResult struct {
	Intro string
	Outro string
}
