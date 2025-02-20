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

func InitSupersetAdminApiRouters(app *fiber.App) {
	controller := supersetApiController{}
	superset := fiber.New()
	app.Mount("/superset", superset)
	superset.Use(middleware.AdminPanelAuthorizationRequired())
	superset.Put("create_dashboard", controller.createDashboard)
}

// @Summary Получение гостевого токена
// @Tags Superset
// @Description Получение гостевого токена
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/superset/guest_token [get]
func (c *supersetApiController) getGuestToken(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	id, err := supersethandler.Instance.GetGuestToken(ctx.Context(), spaceID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения токена")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Создание дашборда для спейса из шаблона
// @Tags Superset admin
// @Description Создание дашборда для спейса из шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	space_id			query 		string	true	"Идентификатор спейса"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/admin_panel/superset/create_dashboard [put]
func (c *supersetApiController) createDashboard(ctx *fiber.Ctx) error {
	spaceID := ctx.Query("space_id", "")
	if spaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не указан идентификатор спейса"))
	}
	err := supersethandler.Instance.ImportDashboard(ctx.Context(), spaceID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}
