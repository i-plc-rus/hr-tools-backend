package extservicestore

type Provider interface {
	Set(spaceID, code string, value []byte) error
	Get(spaceID, code string) (value []byte, err error)
}
