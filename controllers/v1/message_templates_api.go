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
		router.Post("", controller.create)
		router.Get("variables", controller.variables)
		router.Route(":id", func(idRoute fiber.Router) {
			idRoute.Put("", controller.update)
			idRoute.Get("", controller.get)
			idRoute.Delete("", controller.delete)
		})
	})
}

// @Summary Отправить сообщение кандидату на почту
// @Tags Шаблоны сообщений
// @Description Отправить сообщение кандидату на почту
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body				body		msgtemplateapimodels.SendMessage	true	"request body"
// @Success 200 {object} apimodels.Response
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
	userID := middleware.GetUserID(ctx)
	hMsg, err := messagetemplate.Instance.SendEmailMessage(spaceID, payload.MsgTemplateID, payload.ApplicantID, userID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка отправки сообщения кандидату")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка шаблонов")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}

// @Summary Создание
// @Tags Шаблоны сообщений
// @Description Создание
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 msgtemplateapimodels.MsgTemplateData	true	"request body"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router  /api/v1/space/msg-templates [post]
func (c *msgTemplateApiController) create(ctx *fiber.Ctx) error {
	var payload msgtemplateapimodels.MsgTemplateData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	id, err := messagetemplate.Instance.Create(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка добавления шаблона сообщений")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Переменные шаблона
// @Tags Шаблоны сообщений
// @Description Переменные шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=[]msgtemplateapimodels.TemplateItem}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/variables [get]
func (c *msgTemplateApiController) variables(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(messagetemplate.GetVariables()))
}

// @Summary Обновление
// @Tags Шаблоны сообщений
// @Description Обновление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 msgtemplateapimodels.MsgTemplateData	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id} [put]
func (c *msgTemplateApiController) update(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload msgtemplateapimodels.MsgTemplateData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = messagetemplate.Instance.Update(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка изменения шаблона сообщений")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение по ИД
// @Tags Шаблоны сообщений
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=dictapimodels.CompanyView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id} [get]
func (c *msgTemplateApiController) get(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := messagetemplate.Instance.GetByID(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения шаблона сообщений")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Удаление
// @Tags Шаблоны сообщений
// @Description Удаление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id} [delete]
func (c *msgTemplateApiController) delete(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = messagetemplate.Instance.Delete(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления причины шаблона сообщений")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}
