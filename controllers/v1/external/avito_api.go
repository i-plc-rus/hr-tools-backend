package externalapiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	externalservices "hr-tools-backend/lib/external-services"
	avitohandler "hr-tools-backend/lib/external-services/avito"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	avitoapimodels "hr-tools-backend/models/api/avito"
	"strconv"
)

type avitoApiController struct {
	controllers.BaseAPIController
	handler externalservices.JobSiteProvider
}

func InitAvitoApiRouters(app *fiber.App) {
	controller := avitoApiController{
		handler: avitohandler.Instance,
	}
	app.Route("avito", func(router fiber.Router) {
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

// @Summary Проверка подключения к Avito
// @Tags Интеграция Avito
// @Description Проверка подключения к Avito
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/ext/avito/check_connected [get]
func (c *avitoApiController) isConnect(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	connected := c.handler.CheckConnected(ctx.Context(), spaceID)
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(connected))
}

// @Summary Получение ссылки для авторизации
// @Tags Интеграция Avito
// @Description Получение ссылки для авторизации
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/ext/avito/connect_uri [get]
func (c *avitoApiController) connect(ctx *fiber.Ctx) error {

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := c.handler.GetConnectUri(spaceID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения ссылки для авторизации на Avito")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Публикация вакансии
// @Tags Интеграция Avito
// @Description Публикация вакансии
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "идентификатор вакансии"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router  /api/v1/space/ext/avito/{id}/publish [put]
func (c *avitoApiController) publish(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	hMsg, err := c.handler.VacancyPublish(ctx.UserContext(), spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка публикация вакансии на Avito")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Публикация обновления по вакансии
// @Tags Интеграция Avito
// @Description Публикация обновления по вакансии
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "идентификатор вакансии"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router  /api/v1/space/ext/avito/{id}/update [put]
func (c *avitoApiController) update(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	hMsg, err := c.handler.VacancyUpdate(ctx.UserContext(), spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка обновления вакансии на Avito")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Удаление вакансии
// @Tags Интеграция Avito
// @Description Удаление вакансии
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "идентификатор вакансии"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router  /api/v1/space/ext/avito/{id}/close [put]
func (c *avitoApiController) close(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	hMsg, err := c.handler.VacancyClose(ctx.UserContext(), spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка снятия вакансии с публикации на Avito")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Привязать существующую вакансию
// @Tags Интеграция Avito
// @Description Привязать существующую вакансию
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "идентификатор вакансии"
// @Param	body body	 avitoapimodels.VacancyAttach	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/ext/avito/{id}/attach [put]
func (c *avitoApiController) attach(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	var payload avitoapimodels.VacancyAttach
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	hMsg, err := c.handler.VacancyAttach(ctx.UserContext(), spaceID, id, strconv.Itoa(payload.ID))
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка привязки существующей вакансии на Avito")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Статус размещения
// @Tags Интеграция Avito
// @Description Статус размещения
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "идентификатор вакансии"
// @Success 200 {object} apimodels.Response{data=vacancyapimodels.ExtVacancyInfo}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/ext/avito/{id}/status [get]
func (c *avitoApiController) status(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	info, err := c.handler.GetVacancyInfo(ctx.UserContext(), spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения статуса размещения объявления на Avito")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(info))
}
