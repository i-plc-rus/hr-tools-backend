package dict

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	companystructprovider "hr-tools-backend/lib/dicts/company-struct"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	dictapimodels "hr-tools-backend/models/api/dict"
)

type companyStructDictApiController struct {
	controllers.BaseAPIController
}

func InitCompanyStructDictApiRouters(app *fiber.App) {
	controller := companyStructDictApiController{}
	app.Route("company_struct", func(router fiber.Router) {
		router.Use(middleware.RbacMiddleware())
		router.Post("find", controller.findByName)
		router.Post("", controller.create)
		router.Put(":id", controller.update)
		router.Get(":id", controller.get)
		router.Delete(":id", controller.delete)
	})
}

// @Summary Создание
// @Tags Справочник. Структура компании
// @Description Создание
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.CompanyStructData	true	"request body"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/company_struct [post]
func (c *companyStructDictApiController) create(ctx *fiber.Ctx) error {
	var payload dictapimodels.CompanyStructData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	id, err := companystructprovider.Instance.Create(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка создания записи в справочнике структур компаний")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Обновление
// @Tags Справочник. Структура компании
// @Description Обновление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.CompanyStructData	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/company_struct/{id} [put]
func (c *companyStructDictApiController) update(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload dictapimodels.CompanyStructData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = companystructprovider.Instance.Update(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка обновления данных в справочнике структур компаний")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение по ИД
// @Tags Справочник. Структура компании
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=dictapimodels.CompanyStructView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/company_struct/{id} [get]
func (c *companyStructDictApiController) get(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := companystructprovider.Instance.Get(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения записи из справочника структур компаний")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Удаление
// @Tags Справочник. Структура компании
// @Description Удаление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/company_struct/{id} [delete]
func (c *companyStructDictApiController) delete(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = companystructprovider.Instance.Delete(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления данных из справочника структур компаний")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Поиск по названию
// @Tags Справочник. Структура компании
// @Description Поиск по названию
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.CompanyStructData	true	"request body"
// @Success 200 {object} apimodels.Response{data=[]dictapimodels.CompanyStructView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/company_struct/find [post]
func (c *companyStructDictApiController) findByName(ctx *fiber.Ctx) error {
	var payload dictapimodels.CompanyStructData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	list, err := companystructprovider.Instance.FindByName(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка из справочника структур компаний")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}
