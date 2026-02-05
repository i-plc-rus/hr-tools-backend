package apiv1

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	"hr-tools-backend/lib/dadata"
	spacehandler "hr-tools-backend/lib/space/handler"
	apimodels "hr-tools-backend/models/api"
	spaceapimodels "hr-tools-backend/models/api/space"
)

type orgApiController struct {
	controllers.BaseAPIController
}

func InitOrgApiRouters(app *fiber.App) {
	controller := orgApiController{}
	app.Route("organizations", func(router fiber.Router) {
		router.Get("suggest", controller.orgSuggest)
		router.Post("retrieve", controller.orgDetails)
		router.Post("", controller.createOrg)
	})
}

// @Summary Поиск по ИНН через Дадата
// @Tags Организации
// @Description Поиск по ИНН через Дадата
// @Param	query				query		string	false	"параметры запроса в дадату"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/organizations/suggest [get]
func (c *orgApiController) orgSuggest(ctx *fiber.Ctx) error {
	daDataQuery := ctx.Query("query", "")
	response, errs := dadataproxy.ProxySuggestRequest(daDataQuery)
	if len(errs) != 0 {
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"errs": errs,
		})
	}
	return ctx.Status(fiber.StatusOK).Send(response)
}

// @Summary Запрос детальной информации через Дадата
// @Tags Организации
// @Description Запрос детальной информации через Дадата
// @Param	query				query		string	false	"параметры запроса в дадату"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/organizations/retrieve [post]
func (c *orgApiController) orgDetails(ctx *fiber.Ctx) error {
	daDataQuery := ctx.Query("query", "")
	// проверить по ИНН базе и вернуть ошибку если такой уже есть

	// отправить запрос в дадату и вернуть ответ
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(daDataQuery))
}

// @Summary Создание организации
// @Tags Организации
// @Description Создание организации
// @Param	body				body		spaceapimodels.CreateOrganization	true	"request body"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/organizations [post]
func (c *orgApiController) createOrg(ctx *fiber.Ctx) error {
	var payload spaceapimodels.CreateOrganization
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	err := spacehandler.Instance.CreateOrganizationSpace(payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, fmt.Sprintf("Ошибка создания организации: %v", err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}
