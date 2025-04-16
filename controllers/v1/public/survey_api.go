package publicapi

import (
	"hr-tools-backend/controllers"
	"hr-tools-backend/lib/survey"
	apimodels "hr-tools-backend/models/api"
	surveyapimodels "hr-tools-backend/models/api/survey"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

type publicsurveyApiController struct {
	controllers.BaseAPIController
}

func InitPublicSurveyApiRouters(app *fiber.App) {
	controller := publicsurveyApiController{}
	app.Route("survey", func(router fiber.Router) {
		router.Route(":id", func(idRoute fiber.Router) {
			idRoute.Get("", controller.getSurvey)
			idRoute.Put("", controller.setSurvey)
		})
	})
}

// @Summary Получение анкеты
// @Tags Анкета кандидата
// @Description Получение анкеты
// @Param   id          		path    string  true         "Идентификатор ID"
// @Success 200 {object} apimodels.Response{data=surveyapimodels.ApplicantSurveyView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/public/survey/{id} [get]
func (c *publicsurveyApiController) getSurvey(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	resp, err := survey.Instance.GetPublicApplicantSurvey(id)
	if err != nil {
		logger := log.WithField("survey_id", id)
		return c.SendError(ctx, logger, err, "Ошибка получения анкеты")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Сохранение ответов
// @Tags Анкета кандидата
// @Description Сохранение ответов
// @Param   id          		path    string  true         "Идентификатор ID"
// @Param	body body	 surveyapimodels.ApplicantSurveyResponses	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/public/survey/{id} [put]
func (c *publicsurveyApiController) setSurvey(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	var payload surveyapimodels.ApplicantSurveyResponses
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	hMsg, err := survey.Instance.AnswerPublicApplicantSurvey(id, payload.Responses)
	if err != nil {
		logger := log.WithField("survey_id", id)
		return c.SendError(ctx, logger, err, "Ошибка сохранения ответов по анкете")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}

	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}
