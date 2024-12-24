package dict

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	companyprovider "hr-tools-backend/lib/dicts/company"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	dictapimodels "hr-tools-backend/models/api/dict"
)

type companyDictApiController struct {
	controllers.BaseAPIController
}

func InitCompanyDictApiRouters(app *fiber.App) {
	controller := companyDictApiController{}
	app.Route("company", func(router fiber.Router) {
		router.Post("find", controller.companyFindByName)
		router.Use(middleware.SpaceAdminRequired())
		router.Post("", controller.companyCreate)
		router.Put(":id", controller.companyUpdate)
		router.Get(":id", controller.companyGet)
		router.Delete(":id", controller.companyDelete)
	})
}

// @Summary Создание
// @Tags Справочник. Компания
// @Description Создание
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.CompanyData	true	"request body"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/company [post]
func (c *companyDictApiController) companyCreate(ctx *fiber.Ctx) error {
	var payload dictapimodels.CompanyData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	id, err := companyprovider.Instance.Create(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка создания записи в справочнике компаний")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Обновление
// @Tags Справочник. Компания
// @Description Обновление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.CompanyData	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/company/{id} [put]
func (c *companyDictApiController) companyUpdate(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload dictapimodels.CompanyData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = companyprovider.Instance.Update(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка обновления данных в справочнике компаний")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение по ИД
// @Tags Справочник. Компания
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=dictapimodels.CompanyView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/company/{id} [get]
func (c *companyDictApiController) companyGet(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := companyprovider.Instance.Get(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения записи из справочника компаний")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Удаление
// @Tags Справочник. Компания
// @Description Удаление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/company/{id} [delete]
func (c *companyDictApiController) companyDelete(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = companyprovider.Instance.Delete(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления данных из справочника компаний")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Поиск по названию
// @Tags Справочник. Компания
// @Description Поиск по названию
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.CompanyData	true	"request body"
// @Success 200 {object} apimodels.Response{data=[]dictapimodels.CompanyView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/company/find [post]
func (c *companyDictApiController) companyFindByName(ctx *fiber.Ctx) error {
	var payload dictapimodels.CompanyData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	list, err := companyprovider.Instance.FindByName(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка из справочника компаний")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}
