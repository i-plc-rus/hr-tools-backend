package dict

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	departmentprovider "hr-tools-backend/lib/dicts/department"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	dictapimodels "hr-tools-backend/models/api/dict"
)

type departmentDictApiController struct {
	controllers.BaseAPIController
}

func InitDepartmentDictApiRouters(app *fiber.App) {
	controller := departmentDictApiController{}
	app.Route("department", func(router fiber.Router) {
		router.Post("find", controller.departmentFindByName)
		router.Use(middleware.SpaceAdminUser())
		router.Post("", controller.departmentCreate)
		router.Put(":id", controller.departmentUpdate)
		router.Get(":id", controller.departmentGet)
		router.Delete(":id", controller.departmentDelete)
	})
}

// @Summary Создание
// @Tags Справочник. Подразделение
// @Description Создание
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.DepartmentData	true	"request body"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/department [post]
func (c *departmentDictApiController) departmentCreate(ctx *fiber.Ctx) error {
	var payload dictapimodels.DepartmentData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	id, err := departmentprovider.Instance.Create(spaceID, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Обновление
// @Tags Справочник. Подразделение
// @Description Обновление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.DepartmentData	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/department/{id} [put]
func (c *departmentDictApiController) departmentUpdate(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload dictapimodels.DepartmentData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = departmentprovider.Instance.Update(spaceID, id, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение по ИД
// @Tags Справочник. Подразделение
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=dictapimodels.DepartmentView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/department/{id} [get]
func (c *departmentDictApiController) departmentGet(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := departmentprovider.Instance.Get(spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Удаление
// @Tags Справочник. Подразделение
// @Description Удаление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/department/{id} [delete]
func (c *departmentDictApiController) departmentDelete(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = departmentprovider.Instance.Delete(spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Поиск по названию
// @Tags Справочник. Подразделение
// @Description Поиск по названию
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.DepartmentFind	true	"request body"
// @Success 200 {object} apimodels.Response{data=[]dictapimodels.DepartmentView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/department/find [post]
func (c *departmentDictApiController) departmentFindByName(ctx *fiber.Ctx) error {
	var payload dictapimodels.DepartmentFind
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	list, err := departmentprovider.Instance.FindByName(spaceID, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}
