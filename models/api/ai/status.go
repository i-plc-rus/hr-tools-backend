package aiapimodels

type StatusResponse struct {
	IsFree             bool   `json:"is_free"`
	ExecutingRequestID string `json:"executing_request_id"`
}
