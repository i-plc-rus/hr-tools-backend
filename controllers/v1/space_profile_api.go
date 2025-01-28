package apiv1

import (
	"hr-tools-backend/controllers"
	filestorage "hr-tools-backend/lib/file-storage"
	spacehandler "hr-tools-backend/lib/space/handler"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	spaceapimodels "hr-tools-backend/models/api/space"
	dbmodels "hr-tools-backend/models/db"
	"io"

	"github.com/gofiber/fiber/v2"
)

type spaceProfileApiController struct {
	controllers.BaseAPIController
}

func InitSpaceProfileRouters(app *fiber.App) {
	controller := spaceProfileApiController{}
	app.Route("profile", func(route fiber.Router) {
		route.Use(middleware.AuthorizationRequired())
		route.Use(middleware.SpaceAdminRequired())
		route.Get("", controller.getProfile)
		route.Put("", controller.updateProfile)
		route.Post("photo", controller.uploadPhoto)
		route.Get("photo", controller.getPhoto)
	})
}

// @Summary Получение профиля
// @Tags Профиль организации
// @Description Получение профиля
// @Success 200 {object} apimodels.Response{data=spaceapimodels.ProfileData}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/profile [get]
func (c *spaceProfileApiController) getProfile(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	data, err := spacehandler.Instance.GetProfile(spaceID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения профиля организации")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(data))
}

// @Summary Обновление профиля
// @Tags Профиль организации
// @Description Обновление профиля
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/profile [put]
func (c *spaceProfileApiController) updateProfile(ctx *fiber.Ctx) error {
	var payload spaceapimodels.ProfileData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err := spacehandler.Instance.UpdateProfile(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка обновления профиля организации")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Загрузить фото
// @Tags Профиль организации
// @Description Загрузить фото
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   photo				formData	file 	true 	"Фото"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/profile/photo [post]
func (c *spaceProfileApiController) uploadPhoto(ctx *fiber.Ctx) error {
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
	err = filestorage.Instance.Upload(ctx.UserContext(), spaceID, userID, fileBody, file.Filename, dbmodels.CompanyProfilePhoto, contentType)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка сохранения фото профиля")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Скачать фото
// @Tags Профиль организации
// @Description Скачать фото
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/profile/photo [get]
func (c *spaceProfileApiController) getPhoto(ctx *fiber.Ctx) error {
	userID := middleware.GetUserID(ctx)

	spaceID := middleware.GetUserSpace(ctx)
	body, file, err := filestorage.Instance.GetFileByType(ctx.UserContext(), spaceID, userID, dbmodels.CompanyProfilePhoto)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения данных фото профиля")
	}
	if file != nil && file.ContentType != "" {
		ctx.Set(fiber.HeaderContentType, file.ContentType)
		ctx.Set(fiber.HeaderContentDisposition, `inline; filename="`+file.Name+`"`)
	}
	return ctx.Send(body)
}
