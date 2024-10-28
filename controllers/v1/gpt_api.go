package apiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	gpthandler "hr-tools-backend/lib/gpt"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	gptmodels "hr-tools-backend/models/api/gpt"
)

type gptApiController struct {
	controllers.BaseAPIController
}

func InitGptApiRouters(app *fiber.App) {
	controller := gptApiController{}
	app.Route("gpt", func(usersRootRoute fiber.Router) {
		usersRootRoute.Use(middleware.AuthorizationRequired())
		usersRootRoute.Post("generate_vacancy_description", controller.GetVacancyDescription)
	})
}

// @Summary Сгенерировать описание вакансии
// @Tags GPT
// @Description Сгенерировать описание вакансии
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body				body		gptmodels.GenVacancyDescRequest	true	"request body"
// @Success 200 {object} apimodels.Response{data=gptmodels.GenVacancyDescResponse}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/gpt/generate_vacancy_description [post]
func (c *gptApiController) GetVacancyDescription(ctx *fiber.Ctx) error {
	var payload gptmodels.GenVacancyDescRequest
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	resp, err := gpthandler.Instance.GenerateVacancyDescription(spaceID, payload.Text)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusCreated).JSON(apimodels.NewResponse(resp))
}
