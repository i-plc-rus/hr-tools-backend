package apiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	messagetemplate "hr-tools-backend/lib/message-template"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	msgtemplateapimodels "hr-tools-backend/models/api/message-template"
)

type msgTemplateApiController struct {
	controllers.BaseAPIController
}

func InitMsgTemplateApiRouters(app *fiber.App) {
	controller := msgTemplateApiController{}
	app.Route("msg-templates", func(router fiber.Router) {
		router.Post("send-email-msg", controller.SendEmailMessage) // отправить сообщение на почту кандидату
		router.Get("list", controller.GetTemplatesList)            // получить список шаблонов сообщений
	})
}

// @Summary Отправить сообщение кандидату на почту
// @Tags Шаблоны сообщений
// @Description Отправить сообщение кандидату на почту
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body				body		msgtemplateapimodels.SendMessage	true	"request body"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/send-email-msg [post]
func (c *msgTemplateApiController) SendEmailMessage(ctx *fiber.Ctx) error {
	var payload msgtemplateapimodels.SendMessage
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	err := messagetemplate.Instance.SendEmailMessage(spaceID, payload.MsgTemplateID, payload.ApplicantID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Список шаблонов сообщений
// @Tags Шаблоны сообщений
// @Description Список шаблонов сообщений
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=[]msgtemplateapimodels.MsgTemplateView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/list [get]
func (c *msgTemplateApiController) GetTemplatesList(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	list, err := messagetemplate.Instance.GetListTemplates(spaceID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}
