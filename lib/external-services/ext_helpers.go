package externalservices

import "context"

const (
	spaceIDKey   string = "spaceID"
	recIDKey     string = "recID"
	withAuditKey string = "withAudit"
	uriKey       string = "uri"
	requestKey   string = "request"
)

type AuditData struct {
	SpaceID   string
	Request   string
	Uri       string
	RecID     string
	WithAudit bool
}

func GetAuditContext(ctx context.Context, uri string, request []byte) context.Context {
	rCtx := context.WithValue(ctx, withAuditKey, true)
	rCtx = context.WithValue(rCtx, uriKey, uri)
	if len(request) != 0 {
		rCtx = context.WithValue(rCtx, requestKey, string(request))
	}
	return rCtx
}

func GetContextWithRecID(ctx context.Context, spaceID, recID string) context.Context {
	ctx = context.WithValue(ctx, spaceIDKey, spaceID)
	return context.WithValue(ctx, recIDKey, recID)
}

func ExtractAuditData(ctx context.Context) AuditData {
	data := AuditData{
		SpaceID:   ctx.Value(spaceIDKey).(string),
		Request:   ctx.Value(requestKey).(string),
		Uri:       ctx.Value(uriKey).(string),
		RecID:     ctx.Value(recIDKey).(string),
		WithAudit: ctx.Value(withAuditKey).(bool),
	}
	return data
}
