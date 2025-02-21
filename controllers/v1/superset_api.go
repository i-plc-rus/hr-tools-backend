package apiv1

import (
	"hr-tools-backend/controllers"
	supersethandler "hr-tools-backend/lib/superset"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	supersetapimodels "hr-tools-backend/models/api/superset"

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
// @Param	dashboard_code		query 		string	false	"Код дашборда"
// @Success 200 {object} apimodels.Response{data=supersetapimodels.GuestTokenResponse}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/superset/guest_token [get]
func (c *supersetApiController) getGuestToken(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	dashboardCode := ctx.Query("dashboard_code", "")
	if dashboardCode == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не указан код дашборда"))
	}
	token, id, hMsg, err := supersethandler.Instance.GetGuestToken(ctx.Context(), spaceID, dashboardCode)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения токена")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	resp := supersetapimodels.GuestTokenResponse{
		Token:       token,
		DashboardID: id,
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}
