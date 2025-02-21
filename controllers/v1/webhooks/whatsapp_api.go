package webhooksapi

import (
	"github.com/gofiber/fiber/v2"
	"context"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/webhooks"
	"github.com/piusalfred/whatsapp/webhooks/business"
	"github.com/piusalfred/whatsapp/webhooks/message"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	// log "github.com/sirupsen/logrus"
)

func InitwhatsAppWebhookApiRouters(app *fiber.App) {
	messageListener := webhooks.NewListener(
		HandleMessageNotification,
		func(ctx context.Context) (string, error) {
			return "", nil
		},
		&webhooks.ValidateOptions{
			Validate:  false,
			AppSecret: "",
		},
		// LoggingMiddleware[message.Notification],
		// AddMetadataMiddleware[message.Notification],
	)
	businessListener := webhooks.NewListener(
		HandleBusinessNotification,
		func(ctx context.Context) (string, error) {
			return "", nil
		},
		&webhooks.ValidateOptions{
			Validate:  false,
			AppSecret: "",
		},
		// LoggingMiddleware[business.Notification],
		// AddMetadataMiddleware[business.Notification],
	)
	app.Route("whatsapp", func(router fiber.Router) {
		router.Post("messages", adaptor.HTTPHandlerFunc(messageListener.HandleNotification))
		router.Post("business", adaptor.HTTPHandlerFunc(businessListener.HandleNotification))
		router.Post("messages/verify", adaptor.HTTPHandlerFunc(messageListener.HandleSubscriptionVerification))
		router.Post("business/verify", adaptor.HTTPHandlerFunc(businessListener.HandleSubscriptionVerification))
	})
}


func HandleBusinessNotification(ctx context.Context, notification *business.Notification) *webhooks.Response {
	fmt.Printf("Business notification received: %+v\n", notification)
	return &webhooks.Response{StatusCode: http.StatusOK}
}

// @Summary Создание пользователя
// @Tags Webhooks. WhatsApp
// @Description Создание пользователя
// @Success 200
// @Failure 400
// @Failure 403
// @Failure 500
// @router /api/v1/webhooks/whatsapp/messages [post]
func HandleMessageNotification(ctx context.Context, notification *message.Notification) *webhooks.Response {
	fmt.Printf("Message notification received: %+v\n", notification)
	return &webhooks.Response{StatusCode: http.StatusOK}
}

// LoggingMiddleware logs the start and end of request processing
// func LoggingMiddleware[T any](next webhooks.NotificationHandlerFunc[T]) webhooks.NotificationHandlerFunc[T] {
// 	return func(ctx context.Context, notification *T) *webhooks.Response {
// 		fmt.Println("Logging: Before handling notification")
// 		response := next(ctx, notification)
// 		fmt.Println("Logging: After handling notification")
// 		return response
// 	}
// }

// // AddMetadataMiddleware adds some metadata to the context
// func AddMetadataMiddleware[T any](next webhooks.NotificationHandlerFunc[T]) webhooks.NotificationHandlerFunc[T] {
// 	return func(ctx context.Context, notification *T) *webhooks.Response {
// 		fmt.Println("Adding metadata to the context")
// 		ctx = context.WithValue(ctx, "metadata", "some value")
// 		return next(ctx, notification)
// 	}
// }
