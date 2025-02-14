package apiv1

import (
	"hr-tools-backend/controllers"
	negotiationchathandler "hr-tools-backend/lib/external-services/negotiation-chat"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	negotiationapimodels "hr-tools-backend/models/api/negotiation"

	"github.com/gofiber/fiber/v2"
)

type messengerApiController struct {
	controllers.BaseAPIController
}

func InitMessengerApiRouters(app *fiber.App) {
	controller := messengerApiController{}
	app.Route("messenger", func(router fiber.Router) {
		router.Route("job", func(jobMessengerRoute fiber.Router) {
			jobMessengerRoute.Get("is_available", controller.isJobMessengerAvailable)
			jobMessengerRoute.Post("send_message", controller.sendJobMessage)
			jobMessengerRoute.Post("list", controller.jobMessagesList)
		})
		router.Route("whatsup", func(jobMessengerRoute fiber.Router) {
		})
		router.Route("sms", func(jobMessengerRoute fiber.Router) {
		})
	})
}

// @Summary Проверка доступности чата через работный сайт
// @Tags Мессенджер
// @Description Проверка доступности чата через работный сайт
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body				body		negotiationapimodels.MessengerAvailableRequest	true	"request body"
// @Success 200 {object} apimodels.Response{data=negotiationapimodels.MessengerAvailableResponse}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/messenger/job/is_available [get]
func (c *messengerApiController) isJobMessengerAvailable(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	var payload negotiationapimodels.MessengerAvailableRequest
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	resp, err := negotiationchathandler.Instance.IsVailable(spaceID, payload.ApplicantID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка проверки доступности чата через работный сайт")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Отправка нового сообщение кандидату через работный сайт
// @Tags Мессенджер
// @Description Отправка нового сообщение кандидату через работный сайт
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body				body		negotiationapimodels.NewMessageRequest	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/messenger/job/send_message [post]
func (c *messengerApiController) sendJobMessage(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	var payload negotiationapimodels.NewMessageRequest
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	err := negotiationchathandler.Instance.SendMessage(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка отправки нового сообщение кандидату через работный сайт")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение списка сообщений чата с кандидатом через работный сайт
// @Tags Мессенджер
// @Description Получение списка сообщений чата с кандидатом через работный сайт
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body				body		negotiationapimodels.MessageListRequest	true	"request body"
// @Success 200 {object} apimodels.Response{data=negotiationapimodels.MessageItem}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/messenger/job/list [post]
func (c *messengerApiController) jobMessagesList(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	var payload negotiationapimodels.MessageListRequest
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	userID := middleware.GetUserID(ctx)

	list, err := negotiationchathandler.Instance.GetMessages(spaceID, userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка сообщений чата с кандидатом через работный сайт")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}
