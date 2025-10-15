package publicapi

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/controllers"
	"hr-tools-backend/lib/survey"
	"hr-tools-backend/lib/vk"
	apimodels "hr-tools-backend/models/api"
	surveyapimodels "hr-tools-backend/models/api/survey"
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
		router.Route("/step0/:id", func(idRoute fiber.Router) {
			idRoute.Get("", controller.getVkStep0Survey)
			idRoute.Put("", controller.setVkStep0Survey)
		})
		router.Get("/video-interview/:id", controller.getVideoSurveyData)
		router.Post("/upload-answer/:id/:questionID", controller.uploadAnswer)
	})
}

// @Summary Получение анкеты
// @Tags Анкета кандидата
// @Description Получение анкеты
// @Param   id          		path    string  true         "Идентификатор анкеты"
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
// @Param   id          		path    string  true         "Идентификатор анкеты"
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

// @Summary ВК. Шаг 0. Получение анкеты c типовыми вопросами
// @Tags ВК
// @Description ВК. Шаг 0. Получение анкеты c типовыми вопросами
// @Param   id          		path    string  true         "Идентификатор анкеты"
// @Success 200 {object} apimodels.Response{data=surveyapimodels.VkStep0SurveyView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/public/survey/step0/{id} [get]
func (c *publicsurveyApiController) getVkStep0Survey(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	resp, err := vk.Instance.GetSurveyStep0(id)
	if err != nil {
		logger := log.WithField("survey_id", id)
		return c.SendError(ctx, logger, err, "Ошибка получения анкеты")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary ВК. Шаг 0. cохранение ответов по анкете
// @Tags ВК
// @Description ВК. Шаг 0. cохранение ответов по анкете
// @Param   id          		path    string  true         "Идентификатор анкеты"
// @Param	body body	 surveyapimodels.VkStep0SurveyAnswers	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/public/survey/step0/{id} [put]
func (c *publicsurveyApiController) setVkStep0Survey(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	var payload surveyapimodels.VkStep0SurveyAnswers
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	err = payload.Validate()
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	result, err := vk.Instance.HandleSurveyStep0(id, payload)
	if err != nil {
		logger := log.WithField("survey_id", id)
		return c.SendError(ctx, logger, err, "Ошибка сохранения ответов по анкете")
	}

	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(result))
}

// @Summary ВК. Шаг 8. Прохождение видео-интервью (Данные для интервью)
// @Tags ВК
// @Description ВК. Шаг 8. Прохождение видео-интервью (Данные для интервью)
// @Param   id          		path    string  true         "Идентификатор анкеты"
// @Success 200 {object} apimodels.Response{data=surveyapimodels.VkStep1SurveyView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/public/survey/video-interview/{id} [get]
func (c *publicsurveyApiController) getVideoSurveyData(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	resp, err := vk.Instance.GetVideoSurvey(id)
	if err != nil {
		logger := log.WithField("survey_id", id)
		return c.SendError(ctx, logger, err, "Ошибка получения данных для интервью")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary ВК. Шаг 8. Прохождение видео-интервью (загрузка видео ответа на сервер)
// @Tags ВК
// @Description ВК. Шаг 8. Прохождение видео-интервью (загрузка видео ответа на сервер)
// @Param   id          path    	string  true         "Идентификатор анкеты"
// @Param   question_id path    	string  true         "Идентификатор вопроса"
// @Param   file		formData	file 	true 	"Видео"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/public/upload-answer/{id}/{question_id} [post]
func (c *publicsurveyApiController) uploadAnswer(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("Невозможно сохранить ответ, не указан идентификатор кандидата"))
	}

	questionID := ctx.Params("questionID")
	if questionID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("Невозможно сохранить ответ, не указан идентификатор вопроса"))
	}
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("Не удалось получить файл"))
	}
	logger := log.WithField("survey_id", id)
	err = vk.Instance.UploadVideoAnswer(ctx.UserContext(), id, questionID, fileHeader)
	if err != nil {
		return c.SendError(ctx, logger, err, "Ошибка сохранения видео файла")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}
