package dbmodels

type ExtApiAudit struct {
	BaseSpaceModel
	RecID    string
	Service  string
	Uri      string
	Request  string
	Response string
	Status   int
}
