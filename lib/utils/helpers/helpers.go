package helpers

import "context"

func IsContextDone(ctx context.Context) bool {
	if ctx == nil {
		return true
	}
	select {
	case <-ctx.Done():
		return true
	default:
	}
	return false
}
