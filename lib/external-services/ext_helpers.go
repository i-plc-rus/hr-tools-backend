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
		SpaceID:   getStringCtxValue(ctx, spaceIDKey),
		Request:   getStringCtxValue(ctx, requestKey),
		Uri:       getStringCtxValue(ctx, uriKey),
		RecID:     getStringCtxValue(ctx, recIDKey),
		WithAudit: getBoolCtxValue(ctx, withAuditKey),
	}
	return data
}

func getStringCtxValue(ctx context.Context, key string) string {
	value := ctx.Value(key)
	if value == nil {
		return ""
	}
	return value.(string)
}

func getBoolCtxValue(ctx context.Context, key string) bool {
	value := ctx.Value(key)
	if value == nil {
		return false
	}
	return value.(bool)
}
