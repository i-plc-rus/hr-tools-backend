package externalapiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	externalservices "hr-tools-backend/lib/external-services"
	hhhandler "hr-tools-backend/lib/external-services/hh"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	hhapimodels "hr-tools-backend/models/api/hh"
)

type hhApiController struct {
	controllers.BaseAPIController
	handler externalservices.JobSiteProvider
}

func InitHHApiRouters(app *fiber.App) {
	controller := hhApiController{
		handler: hhhandler.Instance,
	}
	app.Route("hh", func(router fiber.Router) {
		router.Get("check_connected", controller.isConnect)
		router.Get("connect_uri", controller.connect)
		router.Route(":id", func(vacancyRoute fiber.Router) {
			vacancyRoute.Put("publish", controller.publish)
			vacancyRoute.Put("update", controller.update)
			vacancyRoute.Put("close", controller.close)
			vacancyRoute.Put("attach", controller.attach)
			vacancyRoute.Get("status", controller.status)
		})
	})
}

// @Summary Проверка подключения к HH
// @Tags Интеграция HeadHunter
// @Description Проверка подключения к HH
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/ext/hh/check_connected [get]
func (c *hhApiController) isConnect(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	connected := c.handler.CheckConnected(spaceID)
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(connected))
}

// @Summary Получение ссылки для авторизации
// @Tags Интеграция HeadHunter
// @Description Получение ссылки для авторизации
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/ext/hh/connect_uri [get]
func (c *hhApiController) connect(ctx *fiber.Ctx) error {

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := c.handler.GetConnectUri(spaceID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Публикация вакансии
// @Tags Интеграция HeadHunter
// @Description Публикация вакансии
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "идентификатор вакансии"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router  /api/v1/space/ext/hh/{id}/publish [put]
func (c *hhApiController) publish(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	err = c.handler.VacancyPublish(ctx.UserContext(), spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Публикация обновления по вакансии
// @Tags Интеграция HeadHunter
// @Description Публикация обновления по вакансии
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "идентификатор вакансии"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router  /api/v1/space/ext/hh/{id}/update [put]
func (c *hhApiController) update(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	err = c.handler.VacancyUpdate(ctx.UserContext(), spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Удаление вакансии
// @Tags Интеграция HeadHunter
// @Description Удаление вакансии
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "идентификатор вакансии"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router  /api/v1/space/ext/hh/{id}/close [put]
func (c *hhApiController) close(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	err = c.handler.VacancyClose(ctx.UserContext(), spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Привязать существующую вакансию
// @Tags Интеграция HeadHunter
// @Description Привязать существующую вакансию
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "идентификатор вакансии"
// @Param	body body	 hhapimodels.VacancyAttach	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/ext/hh/{id}/attach [put]
func (c *hhApiController) attach(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	var payload hhapimodels.VacancyAttach
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	hhID, err := payload.GetID()
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	err = c.handler.VacancyAttach(ctx.UserContext(), spaceID, id, hhID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Статус размещения
// @Tags Интеграция HeadHunter
// @Description Статус размещения
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "идентификатор вакансии"
// @Success 200 {object} apimodels.Response{data=vacancyapimodels.ExtVacancyInfo}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/ext/hh/{id}/status [put]
func (c *hhApiController) status(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	info, err := c.handler.GetVacancyInfo(ctx.UserContext(), spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(info))
}
