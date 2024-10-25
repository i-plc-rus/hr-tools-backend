package external

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	hhhandler "hr-tools-backend/lib/external-services/hh"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
)

type hhApiController struct {
	controllers.BaseAPIController
}

func InitHHApiRouters(app *fiber.App) {
	controller := hhApiController{}
	app.Route("hh", func(router fiber.Router) {
		router.Get("check_connected", controller.isConnect)
		router.Get("connect_uri", controller.connect)
		router.Route(":id", func(vacancyRoute fiber.Router) {
			router.Put("publish", controller.publish)
			router.Put("update", controller.update)
			router.Put("close", controller.close)
			router.Get("negotiations", controller.negotiations) //todo загрузка через воркер?
		})
		router.Get("get_resume", controller.getResume)
	})
}

// @Summary Получение ссылки для авторизации
// @Tags Интеграция HeadHunter
// @Description Получение ссылки для авторизации
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/ext/hh/check_connected [get]
func (c *hhApiController) isConnect(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	connected := hhhandler.Instance.CheckConnected(spaceID)
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(connected))
}

// @Summary Проверка подключения к HH
// @Tags Интеграция HeadHunter
// @Description Проверка подключения к HH
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/ext/hh/connect [get]
func (c *hhApiController) connect(ctx *fiber.Ctx) error {

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := hhhandler.Instance.GetConnectUri(spaceID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Публикация вакансии
// @Tags Интеграция HeadHunter
// @Description Публикация вакансии
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
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
	vacancyUrl, err := hhhandler.Instance.VacancyPublish(ctx.UserContext(), spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(vacancyUrl))
}

// @Summary Редактирование вакансии
// @Tags Интеграция HeadHunter
// @Description Редактирование вакансии
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
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
	err = hhhandler.Instance.VacancyUpdate(ctx.UserContext(), spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Удаление вакансии
// @Tags Интеграция HeadHunter
// @Description Удаление вакансии
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
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
	err = hhhandler.Instance.VacancyClose(ctx.UserContext(), spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @router  /api/v1/space/ext/hh/{id}/negotiations [get]
func (c *hhApiController) negotiations(ctx *fiber.Ctx) error {
	//todo impl
	return nil
}

// @router  /api/v1/space/ext/hh/get_resume [put]
func (c *hhApiController) getResume(ctx *fiber.Ctx) error {
	//todo impl
	return nil
}
