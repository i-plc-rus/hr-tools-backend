package dbmodels

import filesapimodels "hr-tools-backend/models/api/files"

type FileStorage struct {
	BaseSpaceModel
	Name        string
	ApplicantID string
	Type        FileType
}

func (f FileStorage) ToModel() filesapimodels.FileView {
	return filesapimodels.FileView{
		ID:          f.ID,
		Name:        f.Name,
		ApplicantID: f.ApplicantID,
		SpaceID:     f.SpaceID,
	}
}

type FileType string

const (
	ApplicantResume FileType = "applicant_resume"
	ApplicantDoc    FileType = "applicant_doc"
)
