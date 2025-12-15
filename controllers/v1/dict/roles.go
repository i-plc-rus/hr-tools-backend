package dict

import (
	"hr-tools-backend/controllers"
	apimodels "hr-tools-backend/models/api"
	dictapimodels "hr-tools-backend/models/api/dict"

	"github.com/gofiber/fiber/v2"
)

type roleDictApiController struct {
	controllers.BaseAPIController
}

func InitRoleDictApiRouters(app *fiber.App) {
	controller := roleDictApiController{}
	app.Route("role", func(router fiber.Router) {
		router.Get("list", controller.list)
	})
}

// @Summary Список ролей
// @Tags Справочник. Роли
// @Description Список ролей
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.CityData	true	"request body"
// @Success 200 {object} apimodels.Response{data=[]dictapimodels.RoleView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/role/list [get]
func (c *roleDictApiController) list(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(dictapimodels.GetRoles()))
}
