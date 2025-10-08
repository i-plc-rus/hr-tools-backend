package dbmodels

import filesapimodels "hr-tools-backend/models/api/files"

type FileStorage struct {
	BaseSpaceModel
	Name        string
	ApplicantID string
	Type        FileType
	ContentType string
}

func (f FileStorage) ToModel() filesapimodels.FileView {
	return filesapimodels.FileView{
		ID:          f.ID,
		Name:        f.Name,
		ApplicantID: f.ApplicantID,
		SpaceID:     f.SpaceID,
		ContentType: f.ContentType,
	}
}

type FileType string

const (
	ApplicantResume         FileType = "applicant_resume"
	ApplicantDoc            FileType = "applicant_doc"
	ApplicantPhoto          FileType = "applicant_photo"
	UserProfilePhoto        FileType = "user_profile_photo"
	ApplicantVideoInterview FileType = "applicant_video_interview"
	ApplicantEmotions       FileType = "applicant_emotions"
	CompanyProfilePhoto     FileType = "company_profile_photo"
	CompanyLogo             FileType = "company_logo"
	CompanySign             FileType = "company_sign"
	CompanyStamp            FileType = "company_stamp"
)

type UploadFileInfo struct {
	SpaceID        string
	ApplicantID    string
	FileName       string
	FileType       FileType
	ContentType    string
	IsUniqueByName bool
}
