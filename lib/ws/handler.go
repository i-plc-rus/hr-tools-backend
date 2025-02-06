package ws

import (
	wsclient "hr-tools-backend/lib/ws/client"
	connectionhub "hr-tools-backend/lib/ws/hub/connection-hub"
	"hr-tools-backend/middleware"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func InitWs(app *fiber.App) {
	app.Use("", func(ctx *fiber.Ctx) error {
		userID := middleware.GetUserID(ctx)
		ctx.Locals("userID", userID)
		return ctx.Next()
	})
	app.Get("/", websocket.New(supportHandler))
}

// @Summary Системные пуши
// @Tags Websocket Системные пуши
// @Description Системные пуши
// @Param   Authorization		header		string		true		"Authorization token"
// @Success 200 {object} wsmodels.ServerMessage
// @Failure 400
// @Failure 403
// @Failure 500
// @router /ws [get]
func supportHandler(c *websocket.Conn) {

	userID := c.Locals("userID").(string)
	client := wsclient.NewClient(userID, c)
	connectionhub.Instance.AddClient(userID, c)
	defer func() {
		connectionhub.Instance.DeleteClient(userID)
	}()
	client.Dispatch()
}
