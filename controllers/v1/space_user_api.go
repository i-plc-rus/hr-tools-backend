package apiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	spaceusershander "hr-tools-backend/lib/space/users/hander"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	spaceapimodels "hr-tools-backend/models/api/space"
)

type spaceUserController struct {
	controllers.BaseAPIController
}

func InitSpaceUserRouters(app *fiber.App) {
	controller := spaceUserController{}
	app.Route("users", func(usersRootRoute fiber.Router) {
		usersRootRoute.Use(middleware.AuthorizationRequired())
		usersRootRoute.Use(middleware.SpaceAdminRequired())
		usersRootRoute.Post("", controller.CreateUser)
		usersRootRoute.Post("list", controller.ListUsers)
		usersRootRoute.Route(":id", func(usersIDRoute fiber.Router) {
			usersIDRoute.Delete("", controller.DeleteUser)
			usersIDRoute.Put("", controller.UpdateUser)
			usersIDRoute.Get("", controller.GetUserByID)
		})

	})
}

// @Summary Создать нового пользователя
// @Tags Пользователи space
// @Description Создать нового пользователя
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body				body		spaceapimodels.CreateUser	true	"request body"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/users [post]
func (c *spaceUserController) CreateUser(ctx *fiber.Ctx) error {
	var payload spaceapimodels.CreateUser
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	err := spaceusershander.Instance.CreateUser(payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusCreated).JSON(apimodels.NewResponse(nil))
}

// @Summary Удалить пользователя
// @Tags Пользователи space
// @Description Удалить пользователя
// @Param   Authorization		header		string	true	"Authorization token"
// @Param 	id 				path 		string  true 	"space user ID"
// @Param	body				body		spaceapimodels.CreateUser	true	"request body"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/users/{id} [delete]
func (c *spaceUserController) DeleteUser(ctx *fiber.Ctx) error {
	userID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	err = spaceusershander.Instance.DeleteUser(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Обновить пользователя
// @Tags Пользователи space
// @Description Обновить пользователя
// @Param   Authorization		header		string	true	"Authorization token"
// @Param 	id 				path 		string  true 	"space user ID"
// @Param	body				body		spaceapimodels.UpdateUser	true	"request body"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/users/{id} [put]
func (c *spaceUserController) UpdateUser(ctx *fiber.Ctx) error {
	userID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	var payload spaceapimodels.UpdateUser
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	err = spaceusershander.Instance.UpdateUser(userID, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusCreated).JSON(apimodels.NewResponse(nil))
}

// @Summary Получить список пользователей space
// @Tags Пользователи space
// @Description Получить список пользователей space
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body				body		apimodels.Pagination	true	"request body"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/users/list [post]
func (c *spaceUserController) ListUsers(ctx *fiber.Ctx) error {
	var payload apimodels.Pagination
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	users, err := spaceusershander.Instance.GetListUsers(spaceID, payload.Page, payload.Limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusCreated).JSON(apimodels.NewResponse(users))
}

// @Summary Получить пользователя space по ID
// @Tags Пользователи space
// @Description Получить пользователя space по ID
// @Param   Authorization		header		string	true	"Authorization token"
// @Param 	id 				path 		string  true 	"space user ID"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/users/{id} [get]
func (c *spaceUserController) GetUserByID(ctx *fiber.Ctx) error {
	userID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	user, err := spaceusershander.Instance.GetByID(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusCreated).JSON(apimodels.NewResponse(user))
}
