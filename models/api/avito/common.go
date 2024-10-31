package avitoapimodels

type ErrorData struct {
	Error ErrorItem `json:"error"`
}

type ErrorItem struct {
	Err400
	Err401
	Err402X
	Err5XX
}

type Err400 struct {
	Reason string `json:"reason"`
	Type   string `json:"type"`
	Value  string `json:"value"`
}

type Err401 struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Err402X struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Err5XX struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
