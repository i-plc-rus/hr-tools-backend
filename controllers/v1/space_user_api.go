package apiv1

import (
	"hr-tools-backend/controllers"
	filestorage "hr-tools-backend/lib/file-storage"
	spaceusershander "hr-tools-backend/lib/space/users/hander"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	spaceapimodels "hr-tools-backend/models/api/space"
	dbmodels "hr-tools-backend/models/db"
	"io"

	"github.com/gofiber/fiber/v2"
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
	app.Route("user_profile", func(userRootRoute fiber.Router) {
		userRootRoute.Use(middleware.AuthorizationRequired())
		userRootRoute.Get("", controller.getProfile)
		userRootRoute.Put("", controller.updateProfile)
		userRootRoute.Put("change_password", controller.changePassword)
		userRootRoute.Post("photo", controller.uploadPhoto) // загрузить фото
		userRootRoute.Get("photo", controller.getPhoto)     // скачать фото
	})
}

// @Summary Создать нового пользователя
// @Tags Пользователи space
// @Description Создать нового пользователя
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body				body		spaceapimodels.CreateUser	true	"request body"
// @Success 200 {object} apimodels.Response{data=string}
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
	id, hMsg, err := spaceusershander.Instance.CreateUser(payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка создания пользователя")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusCreated).JSON(apimodels.NewResponse(id))
}

// @Summary Удалить пользователя
// @Tags Пользователи space
// @Description Удалить пользователя
// @Param   Authorization		header		string	true	"Authorization token"
// @Param 	id 				path 		string  true 	"space user ID"
// @Param	body				body		spaceapimodels.CreateUser	true	"request body"
// @Success 200 {object} apimodels.Response
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления пользователя")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка обновления данных пользователя")
	}
	return ctx.Status(fiber.StatusCreated).JSON(apimodels.NewResponse(nil))
}

// @Summary Получить список пользователей space
// @Tags Пользователи space
// @Description Получить список пользователей space
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body				body		spaceapimodels.SpaceUserFilter	true	"request body"
// @Success 200 {object} apimodels.ScrollerResponse{data=[]spaceapimodels.SpaceUser}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/users/list [post]
func (c *spaceUserController) ListUsers(ctx *fiber.Ctx) error {
	var payload spaceapimodels.SpaceUserFilter
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	users, rowCount, err := spaceusershander.Instance.GetListUsers(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка пользователей")
	}
	return ctx.Status(fiber.StatusCreated).JSON(apimodels.NewScrollerResponse(users, rowCount))
}

// @Summary Получить пользователя space по ID
// @Tags Пользователи space
// @Description Получить пользователя space по ID
// @Param   Authorization		header		string	true	"Authorization token"
// @Param 	id 				path 		string  true 	"space user ID"
// @Success 200 {object} apimodels.Response{data=spaceapimodels.SpaceUser}
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения данных пользователя")
	}
	return ctx.Status(fiber.StatusCreated).JSON(apimodels.NewResponse(user))
}

// @Summary Получить профиль пользователя
// @Tags Профиль пользователя space
// @Description Получить профиль пользователя
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=spaceapimodels.SpaceUserProfileView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/user_profile [get]
func (c *spaceUserController) getProfile(ctx *fiber.Ctx) error {
	userID := middleware.GetUserID(ctx)
	user, err := spaceusershander.Instance.GetProfileByID(userID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения данных профиля")
	}
	return ctx.Status(fiber.StatusCreated).JSON(apimodels.NewResponse(user))
}

// @Summary Обновить профиль пользователя
// @Tags Профиль пользователя space
// @Description Обновить профиль пользователя
// @Param   Authorization	header		string	true	"Authorization token"
// @Param 	id 				path 		string  true 	"space user ID"
// @Param	body			body		spaceapimodels.SpaceUserProfileData	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/user_profile [put]
func (c *spaceUserController) updateProfile(ctx *fiber.Ctx) error {
	userID := middleware.GetUserID(ctx)
	var payload spaceapimodels.SpaceUserProfileData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	err := spaceusershander.Instance.UpdateUserProfile(userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка обновления профиля")
	}
	return ctx.Status(fiber.StatusCreated).JSON(apimodels.NewResponse(nil))
}

// @Summary Изменить пароль
// @Tags Профиль пользователя space
// @Description Изменить пароль
// @Param	body	body		spaceapimodels.PasswordChange	true	"request body"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/user_profile/change_password [put]
func (c *spaceUserController) changePassword(ctx *fiber.Ctx) error {
	userID := middleware.GetUserID(ctx)
	var payload spaceapimodels.PasswordChange
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	hMsg, err := spaceusershander.Instance.СhangePassword(userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка изменения пароля")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Загрузить фото
// @Tags Профиль пользователя space
// @Description Загрузить фото
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   photo				formData	file 	true 	"Фото"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/user_profile/photo [post]
func (c *spaceUserController) uploadPhoto(ctx *fiber.Ctx) error {
	userID := middleware.GetUserID(ctx)

	file, err := ctx.FormFile("photo")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	buffer, err := file.Open()
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка при получении файла с фото")
	}
	defer buffer.Close()
	fileBody, err := io.ReadAll(buffer)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка при загрузке файла с фото")
	}

	spaceID := middleware.GetUserSpace(ctx)
	contentType := helpers.GetFileContentType(file)
	err = filestorage.Instance.Upload(ctx.UserContext(), spaceID, userID, fileBody, file.Filename, dbmodels.UserProfilePhoto, contentType)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка сохранения фото профиля")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Скачать фото
// @Tags Профиль пользователя space
// @Description Скачать фото
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/user_profile/photo [get]
func (c *spaceUserController) getPhoto(ctx *fiber.Ctx) error {
	userID := middleware.GetUserID(ctx)

	spaceID := middleware.GetUserSpace(ctx)
	body, file, err := filestorage.Instance.GetFileByType(ctx.UserContext(), spaceID, userID, dbmodels.UserProfilePhoto)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения данных фото профиля")
	}
	if file != nil && file.ContentType != "" {
		ctx.Set(fiber.HeaderContentType, file.ContentType)
		ctx.Set(fiber.HeaderContentDisposition, `inline; filename="`+file.Name+`"`)
	}
	return ctx.Send(body)
}
