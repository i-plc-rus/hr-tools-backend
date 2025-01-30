package dict

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	cityprovider "hr-tools-backend/lib/dicts/city"
	apimodels "hr-tools-backend/models/api"
	dictapimodels "hr-tools-backend/models/api/dict"
)

type cityDictApiController struct {
	controllers.BaseAPIController
}

func InitCityDictApiRouters(app *fiber.App) {
	controller := cityDictApiController{}
	app.Route("city", func(router fiber.Router) {
		router.Post("find", controller.cityFindByName)
		router.Get(":id", controller.cityGet)
	})
}

// @Summary Получение по ИД
// @Tags Справочник. Города
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=dictapimodels.CityView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/city/{id} [get]
func (c *cityDictApiController) cityGet(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	resp, err := cityprovider.Instance.Get(id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения данных по городу")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Поиск по названию
// @Tags Справочник. Города
// @Description Поиск по названию
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.CityData	true	"request body"
// @Success 200 {object} apimodels.Response{data=[]dictapimodels.CityView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/city/find [post]
func (c *cityDictApiController) cityFindByName(ctx *fiber.Ctx) error {
	var payload dictapimodels.CityData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	list, err := cityprovider.Instance.FindByName(payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка городов")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}
