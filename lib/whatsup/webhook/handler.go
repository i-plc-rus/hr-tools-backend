package whatsappwebhook

import (
	"context"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/webhooks"
	"github.com/piusalfred/whatsapp/webhooks/business"
	"github.com/piusalfred/whatsapp/webhooks/message"
)

func HandleBusinessNotification(ctx context.Context, notification *business.Notification) *webhooks.Response {
	fmt.Printf("Business notification received: %+v\n", notification)
	return &webhooks.Response{StatusCode: http.StatusOK}
}

func HandleMessageNotification(ctx context.Context, notification *message.Notification) *webhooks.Response {
	fmt.Printf("Message notification received: %+v\n", notification)
	return &webhooks.Response{StatusCode: http.StatusOK}
}

// LoggingMiddleware logs the start and end of request processing
func LoggingMiddleware[T any](next webhooks.NotificationHandlerFunc[T]) webhooks.NotificationHandlerFunc[T] {
	return func(ctx context.Context, notification *T) *webhooks.Response {
		fmt.Println("Logging: Before handling notification")
		response := next(ctx, notification)
		fmt.Println("Logging: After handling notification")
		return response
	}
}

// AddMetadataMiddleware adds some metadata to the context
func AddMetadataMiddleware[T any](next webhooks.NotificationHandlerFunc[T]) webhooks.NotificationHandlerFunc[T] {
	return func(ctx context.Context, notification *T) *webhooks.Response {
		fmt.Println("Adding metadata to the context")
		ctx = context.WithValue(ctx, "metadata", "some value")
		return next(ctx, notification)
	}
}