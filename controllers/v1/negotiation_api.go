package apiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	"hr-tools-backend/lib/applicant"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"
	dbmodels "hr-tools-backend/models/db"
)

type negotiationApiController struct {
	controllers.BaseAPIController
}

func InitNegotiationApiRouters(app *fiber.App) {
	controller := negotiationApiController{}
	app.Route("negotiation", func(router fiber.Router) {
		router.Post("list", controller.list)
		router.Route(":id", func(idRoute fiber.Router) {
			idRoute.Put("status_change", controller.statusChange)
			idRoute.Get("", controller.get)
			idRoute.Put("comment", controller.updateComment)
		})
	})
}

// @Summary Список
// @Tags Отклики
// @Description Список
// @Param	body body	 dbmodels.NegotiationFilter	true	"request filter body"
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=[]negotiationapimodels.NegotiationView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/negotiation/list [post]
func (c *negotiationApiController) list(ctx *fiber.Ctx) error {
	var payload dbmodels.NegotiationFilter
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	list, err := applicant.Instance.ListOfNegotiation(spaceID, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}

// @Summary Смена статуса
// @Tags Отклики
// @Description Смена статуса
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 negotiationapimodels.StatusData	true	"new status"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/negotiation/{id}/status_change [put]
func (c *negotiationApiController) statusChange(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload negotiationapimodels.StatusData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	err = applicant.Instance.UpdateStatus(spaceID, id, payload.Status)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение по ИД
// @Tags Отклики
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=negotiationapimodels.NegotiationView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/negotiation/{id} [get]
func (c *negotiationApiController) get(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := applicant.Instance.GetByID(spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Сохранить комментарий
// @Tags Отклики
// @Description Сохранить комментарий
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 negotiationapimodels.CommentData	true	"Comment data"
// @Param   id   path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/negotiation/{id}/comment [put]
func (c *negotiationApiController) updateComment(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload negotiationapimodels.CommentData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	err = applicant.Instance.UpdateComment(id, payload.Comment)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}
