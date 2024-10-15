package apiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	aprovalstageshandler "hr-tools-backend/lib/aproval-stages"
	vacancyreqhandler "hr-tools-backend/lib/vacancy-req"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
)

type vacancyReqApiController struct {
	controllers.BaseAPIController
}

func InitVacancyRequestApiRouters(app *fiber.App) {
	controller := vacancyReqApiController{}
	app.Route("vacancy_request", func(router fiber.Router) {
		router.Post("list", controller.list)
		router.Post("", controller.create)
		router.Route(":id", func(idRoute fiber.Router) {
			idRoute.Put("", controller.update)
			idRoute.Get("", controller.get)
			idRoute.Delete("", controller.delete)
			idRoute.Put("approval_stages", controller.saveStages)
		})
	})
}

// @Summary Создание
// @Tags Заявка
// @Description Создание
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 vacancyapimodels.VacancyRequestData	true	"request body"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request [post]
func (c *vacancyReqApiController) create(ctx *fiber.Ctx) error {
	var payload vacancyapimodels.VacancyRequestData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	id, err := vacancyreqhandler.Instance.Create(spaceID, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Обновление
// @Tags Заявка
// @Description Обновление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 vacancyapimodels.VacancyRequestData	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id} [put]
func (c *vacancyReqApiController) update(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload vacancyapimodels.VacancyRequestData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = vacancyreqhandler.Instance.Update(spaceID, id, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение по ИД
// @Tags Заявка
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=vacancyapimodels.VacancyRequestView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id} [get]
func (c *vacancyReqApiController) get(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := vacancyreqhandler.Instance.GetByID(spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Удаление
// @Tags Заявка
// @Description Удаление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id} [delete]
func (c *vacancyReqApiController) delete(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = vacancyreqhandler.Instance.Delete(spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Список
// @Tags Заявка
// @Description Список
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=[]vacancyapimodels.VacancyRequestView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/list [post]
func (c *vacancyReqApiController) list(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	list, err := vacancyreqhandler.Instance.List(spaceID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}

// @Summary Обновление цепочки согласования
// @Tags Заявка
// @Description Обновление цепочки согласования
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 []vacancyapimodels.ApprovalStages	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/approval_stages [put]
func (c *vacancyReqApiController) saveStages(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload vacancyapimodels.ApprovalStages
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = aprovalstageshandler.Instance.Save(spaceID, id, payload.ApprovalStages)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}
