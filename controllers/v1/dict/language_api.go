package dict

import (
	"hr-tools-backend/controllers"
	languagesprovider "hr-tools-backend/lib/dicts/languages"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	dictapimodels "hr-tools-backend/models/api/dict"

	"github.com/gofiber/fiber/v2"
)

type langDictApiController struct {
	controllers.BaseAPIController
}

func InitLangDictApiRouters(app *fiber.App) {
	controller := langDictApiController{}
	app.Route("lang", func(router fiber.Router) {
		router.Use(middleware.RbacMiddleware())
		router.Post("find", controller.langFindByName)
	})
}

// @Summary Поиск по названию
// @Tags Справочник. Языки
// @Description Поиск по названию
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.LangData	true	"request body"
// @Success 200 {object} apimodels.Response{data=[]dictapimodels.LangView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/lang/find [post]
func (c *langDictApiController) langFindByName(ctx *fiber.Ctx) error {
	var payload dictapimodels.LangData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	list, err := languagesprovider.Instance.FindByName(payload.Name)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка из справочника компаний")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}
