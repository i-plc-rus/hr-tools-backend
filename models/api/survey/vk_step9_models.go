package surveyapimodels

type SemanticData struct {
	Question string
	Comment  string
	Answer   string
}

type VkStep9ScoreResult struct {
	Similarity int
	Comment    string
}
