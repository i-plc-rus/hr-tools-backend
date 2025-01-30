package analytics

import (
	"bytes"
	"hr-tools-backend/lib/applicant"
	xlsexport "hr-tools-backend/lib/export/xls"
	applicantapimodels "hr-tools-backend/models/api/applicant"
)

type Provider interface {
	Source(spaceID string, filter applicantapimodels.ApplicantFilter) (applicantapimodels.ApplicantSourceData, error)
	SourceExportToXls(spaceID string, filter applicantapimodels.ApplicantFilter) (*bytes.Buffer, error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		applicantProvider: applicant.Instance,
	}
}

type impl struct {
	applicantProvider applicant.Provider
}

func (i impl) Source(spaceID string, filter applicantapimodels.ApplicantFilter) (applicantapimodels.ApplicantSourceData, error) {
	return i.applicantProvider.ListOfSource(spaceID, filter)
}

func (i impl) SourceExportToXls(spaceID string, filter applicantapimodels.ApplicantFilter) (*bytes.Buffer, error) {
	data, err := i.applicantProvider.ListOfSource(spaceID, filter)
	if err != nil {
		return nil, err
	}
	return xlsexport.Instance.ExportSource(data)
}
