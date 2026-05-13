package dict

import (
	"hr-tools-backend/controllers"
	"hr-tools-backend/middleware"
	"hr-tools-backend/models"
	apimodels "hr-tools-backend/models/api"

	"github.com/gofiber/fiber/v2"
)

type commonDictApiController struct {
	controllers.BaseAPIController
}

func InitCommonDictApiRouters(app *fiber.App) {
	controller := commonDictApiController{}
	app.Route("common", func(router fiber.Router) {
		router.Use(middleware.RbacMiddleware())
		router.Get("", controller.get)
	})
}

// @Summary Получение справочников
// @Tags Справочник. Статические справочники
// @Description Получение справочников
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=models.CommonDict}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/common [get]
func (c *commonDictApiController) get(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(models.GetCommonDicts()))
}
