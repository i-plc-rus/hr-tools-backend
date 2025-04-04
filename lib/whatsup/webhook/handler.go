package whatsappwebhook

import (
	"context"
	"github.com/piusalfred/whatsapp/webhooks"
)

// LoggingMiddleware logs the start and end of request processing
func LoggingMiddleware[T any](next webhooks.NotificationHandlerFunc[T]) webhooks.NotificationHandlerFunc[T] {
	return func(ctx context.Context, notification *T) *webhooks.Response {
		// Before handling notification
		response := next(ctx, notification)
		// After handling notification
		return response
	}
}

// AddMetadataMiddleware adds some metadata to the context
func AddMetadataMiddleware[T any](next webhooks.NotificationHandlerFunc[T]) webhooks.NotificationHandlerFunc[T] {
	return func(ctx context.Context, notification *T) *webhooks.Response {
		// Can add metadata to the context")
		// ctx = context.WithValue(ctx, "metadata", "some value")
		return next(ctx, notification)
	}
}
