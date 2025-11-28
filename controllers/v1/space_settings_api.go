package apiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	spacesettingshandler "hr-tools-backend/lib/space/settings/handler"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	spaceapimodels "hr-tools-backend/models/api/space"
)

type spaceSettingsApiController struct {
	controllers.BaseAPIController
}

func InitSpaceSettingRouters(app *fiber.App) {
	controller := spaceSettingsApiController{}
	app.Route("settings", func(usersRootRoute fiber.Router) {
		usersRootRoute.Use(middleware.AuthorizationRequired())
		usersRootRoute.Use(middleware.SpaceAdminRequired())
		
		usersRootRoute.Get("list", controller.ListSettings)
		usersRootRoute.Route(":code", func(usersIDRoute fiber.Router) {
			usersIDRoute.Put("", controller.UpdateSetting)
		})

	})
}

// @Summary Обновить значение настройки пространства
// @Tags Настройки space
// @Description Обновить значение настройки пространства
// @Param   Authorization		header		string	true	"Authorization token"
// @Param 	code 				path 		string  true 	"space setting code"
// @Param	body				body		spaceapimodels.UpdateSpaceSettingValue	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/settings/{code} [put]
func (c *spaceSettingsApiController) UpdateSetting(ctx *fiber.Ctx) error {
	settingCode, err := c.GetIDByKey(ctx, "code")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	var payload spaceapimodels.UpdateSpaceSettingValue
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	err = spacesettingshandler.Instance.UpdateSettingValue(spaceID, settingCode, payload.Value)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка обновления настройки")
	}
	return ctx.Status(fiber.StatusCreated).JSON(apimodels.NewResponse(nil))
}

// @Summary Список настроек
// @Tags Настройки space
// @Description Список настроек
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=[]spaceapimodels.SpaceSettingView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/settings/list [get]
func (c *spaceSettingsApiController) ListSettings(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	list, err := spacesettingshandler.Instance.GetList(spaceID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка настроек")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}
