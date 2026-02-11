package dict

import (
	"hr-tools-backend/controllers"
	rejectreasonprovider "hr-tools-backend/lib/dicts/reject-reason"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	dictapimodels "hr-tools-backend/models/api/dict"

	"github.com/gofiber/fiber/v2"
)

type rejectReasonDictApiController struct {
	controllers.BaseAPIController
}

func InitRejectReasonDictApiRouters(app *fiber.App) {
	controller := rejectReasonDictApiController{}
	app.Route("reject_reason", func(router fiber.Router) {
		router.Use(middleware.RbacMiddleware())
		router.Post("find", controller.find)
		router.Post("", controller.create)
		router.Put(":id", controller.update)
		router.Get(":id", controller.get)
		router.Delete(":id", controller.delete)
	})
}

// @Summary Список
// @Tags Справочник. Причины отказа
// @Description Список
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.RejectReasonFind	false	"request body"
// @Success 200 {object} apimodels.Response{data=[]dictapimodels.RejectReasonView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/reject_reason/find [post]
func (c *rejectReasonDictApiController) find(ctx *fiber.Ctx) error {
	var payload dictapimodels.RejectReasonFind
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	list, err := rejectreasonprovider.Instance.List(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка причин отказов")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}

// @Summary Создание
// @Tags Справочник. Причины отказа
// @Description Создание
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.RejectReasonData	true	"request body"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/reject_reason [post]
func (c *rejectReasonDictApiController) create(ctx *fiber.Ctx) error {
	var payload dictapimodels.RejectReasonData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	id, hMsg, err := rejectreasonprovider.Instance.Create(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка добавления причины отказа")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Обновление
// @Tags Справочник. Причины отказа
// @Description Обновление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.RejectReasonData	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/reject_reason/{id} [put]
func (c *rejectReasonDictApiController) update(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload dictapimodels.RejectReasonData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	hMsg, err := rejectreasonprovider.Instance.Update(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка изменения причины отказа")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение по ИД
// @Tags Справочник. Причины отказа
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=dictapimodels.RejectReasonView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/reject_reason/{id} [get]
func (c *rejectReasonDictApiController) get(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := rejectreasonprovider.Instance.Get(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения причины отказа")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Удаление
// @Tags Справочник. Причины отказа
// @Description Удаление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/reject_reason/{id} [delete]
func (c *rejectReasonDictApiController) delete(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = rejectreasonprovider.Instance.Delete(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления причины отказа")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}
