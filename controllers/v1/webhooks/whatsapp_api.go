package webhooksapi

import (
	"context"
	"encoding/json"
	"hr-tools-backend/lib/whatsup"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/piusalfred/whatsapp/webhooks"
	"github.com/piusalfred/whatsapp/webhooks/business"
	"github.com/piusalfred/whatsapp/webhooks/message"
	log "github.com/sirupsen/logrus"
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
	)
	app.Route("whatsapp", func(router fiber.Router) {
		router.Post("messages", adaptor.HTTPHandlerFunc(messageListener.HandleNotification))
		router.Post("business", adaptor.HTTPHandlerFunc(businessListener.HandleNotification))
		router.Post("messages/verify", adaptor.HTTPHandlerFunc(messageListener.HandleSubscriptionVerification))
		router.Post("business/verify", adaptor.HTTPHandlerFunc(businessListener.HandleSubscriptionVerification))
	})
}

// @Summary Business messages WebHookApi
// @Tags Webhooks. WhatsApp
// @Description Business messages WebHookApi
// @Success 200
// @Failure 400
// @Failure 403
// @Failure 500
// @router /api/v1/webhooks/whatsapp/business [post]
func HandleBusinessNotification(ctx context.Context, notification *business.Notification) *webhooks.Response {
	for _, entity := range notification.Entry {
		logger := log.WithField("business_account_id", entity.ID)
		body, err := json.Marshal(notification)
		if err == nil {
			logger = logger.
				WithField("request_body", string(body))
		}
		logger.Info("Получен вебхук whatsApp business API")
	}
	return &webhooks.Response{StatusCode: http.StatusOK}
}

// @Summary Messages WebHookApi
// @Tags Webhooks. WhatsApp
// @Description Messages WebHookApi
// @Success 200
// @Failure 400
// @Failure 403
// @Failure 500
// @router /api/v1/webhooks/whatsapp/messages [post]
func HandleMessageNotification(ctx context.Context, notification *message.Notification) *webhooks.Response {
	for _, entity := range notification.Entry {
		logger := log.WithField("business_account_id", entity.ID)
		body, err := json.Marshal(notification)
		if err == nil {
			logger = logger.
				WithField("request_body", string(body))
		}
		logger.Info("Получен вебхук whatsApp messages API")
		whatsup.Instance.HandleWebHook(ctx, entity)
	}
	return &webhooks.Response{StatusCode: http.StatusOK}
}
