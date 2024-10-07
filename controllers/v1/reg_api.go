package apiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	spaceauthhandler "hr-tools-backend/lib/space/auth"
	apimodels "hr-tools-backend/models/api"
	authapimodels "hr-tools-backend/models/api/auth"
)

type regController struct {
	controllers.BaseAPIController
}

func InitRegRouters(app *fiber.App) {
	controller := regController{}
	app.Route("auth", func(router fiber.Router) {
		router.Post("send-email-confirm", controller.SendEmailConfirm) // отправить на почту код с подтверждением
		router.Get("verify-email", controller.VerifyEmail)             // подтвердить почту
		router.Post("check-email", controller.CheckEmail)              // проверить почту (на дубли в системе)
	})
}

// @Summary Отправить письмо с подтверждением на почту
// @Tags Регистрация_организации
// @Description Отправляет письмо с подтверждением на почту
// @Param	body				body		authapimodels.SendEmail	true	"request body"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/auth/register [post]
func (c *regController) SendEmailConfirm(ctx *fiber.Ctx) error {
	var payload authapimodels.SendEmail
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	err := spaceauthhandler.Instance.SendEmailConfirmation(payload.Email)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Подтверждение почты кодом
// @Tags Регистрация_организации
// @Description Подтверждение почты кодом
// @Param	code				query		string	false	"код подтверждения"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/auth/verify-email [get]
func (c *regController) VerifyEmail(ctx *fiber.Ctx) error {
	verifyCode := ctx.Query("code", "")
	err := spaceauthhandler.Instance.VerifyEmail(verifyCode)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Проверить почту
// @Tags Регистрация_организации
// @Description Проверить почту
// @Param	body				body		authapimodels.SendEmail	true	"request body"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/auth/check-email [post]
func (c *regController) CheckEmail(ctx *fiber.Ctx) error {
	var payload authapimodels.SendEmail
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	passed, err := spaceauthhandler.Instance.CheckEmail(payload.Email)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	if !passed {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("данная почта уже существует в системе"))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}
