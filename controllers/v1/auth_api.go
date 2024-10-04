package apiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	spaceauthhandler "hr-tools-backend/lib/space/auth"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	authapimodels "hr-tools-backend/models/api/auth"
)

type authApiController struct {
	controllers.BaseAPIController
}

func InitAuthApiRouters(app *fiber.App) {
	controller := authApiController{}
	app.Route("auth", func(router fiber.Router) {
		router.Post("login", controller.login)
		router.Use(middleware.AuthorizationRequired()).Get("me", controller.me)
		router.Post("refresh-token", controller.refreshToken)
	})
}

// @Summary Аутентификация пользователя
// @Tags Аутентификация пользователей
// @Description Аутентификация пользователя
// @Param	body				body		authapimodels.LoginRequest	true	"request body"
// @Success 200 {object} apimodels.Response{data=authapimodels.JWTResponse}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/auth/login [post]
func (c *authApiController) login(ctx *fiber.Ctx) error {
	var payload authapimodels.LoginRequest
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	resp, err := spaceauthhandler.Instance.Login(payload.Email, payload.Password)
	if err != nil {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Получить информацию о текущем пользователе
// @Tags Аутентификация пользователей
// @Description Получить информацию о текущем пользователе
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} spaceapimodels.SpaceUser
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/auth/me [get]
func (c *authApiController) me(ctx *fiber.Ctx) error {
	resp, err := spaceauthhandler.Instance.Me(ctx)
	if err != nil {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Обновить JWT
// @Tags Аутентификация пользователей
// @Description Обновить JWT
// @Param	body				body		authapimodels.JWTRefreshRequest	true	"request body"
// @Success 200 {object} apimodels.Response{data=authapimodels.JWTResponse}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/auth/refresh-token [post]
func (c *authApiController) refreshToken(ctx *fiber.Ctx) error {
	var payload authapimodels.JWTRefreshRequest
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	resp, err := spaceauthhandler.Instance.RefreshToken(ctx, payload.RefreshToken)
	if err != nil {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}
