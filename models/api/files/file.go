package filesapimodels

type FileView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ApplicantID string `json:"applicant_id"`
	SpaceID     string `json:"space_id"`
}
