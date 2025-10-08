package surveyapimodels

import (
	"io"
)

type VkAiInterviewProvider interface {
	AnalyzeAnswer(vkStepID, applicantID, questionID string, reader io.Reader) (result VkAiInterviewResponse, err error)
}

type VkAiInterviewResponse struct {
	RecognizedText string
	VoiceAmplitude *VkResponseFileData
	Frames         *VkResponseFileData
	Emotion        *VkResponseFileData
	Sentiment      *VkResponseFileData
}

type VkResponseFileData struct {
	Body        []byte
	ContentType string
}
