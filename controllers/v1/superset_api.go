package apiv1

import (
	"hr-tools-backend/controllers"
	supersethandler "hr-tools-backend/lib/superset"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"

	"github.com/gofiber/fiber/v2"
)

type supersetApiController struct {
	controllers.BaseAPIController
}

func InitSupersetApiRouters(app *fiber.App) {
	controller := supersetApiController{}
	app.Route("superset", func(router fiber.Router) {
		router.Get("guest_token", controller.getGuestToken)
	})
}

// @Summary Получение гостевого токена
// @Tags Superset
// @Description Получение гостевого токена
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/superset/guest_token [post]
func (c *supersetApiController) getGuestToken(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	id, err := supersethandler.Instance.GetGuestToken(ctx.Context(), spaceID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения токена")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}
