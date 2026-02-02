package apiv1

import (
	"hr-tools-backend/controllers"
	promptcheckhandler "hr-tools-backend/lib/ai/prompt-check"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"

	"github.com/gofiber/fiber/v2"
)

type aiController struct {
	controllers.BaseAPIController
}

func InitAIRouters(app *fiber.App) {
	controller := aiController{}
	app.Route("ai", func(aiRoute fiber.Router) {
		aiRoute.Use(middleware.AuthorizationRequired())
		aiRoute.Use(middleware.LicenseRequired())
		aiRoute.Use(middleware.RbacMiddleware())

		aiRoute.Post("prompt", controller.RunPrompt)
		aiRoute.Get(":id/result", controller.GetResult)
		aiRoute.Get("status", controller.GetStatus)

		aiRoute.Route("questions_prompt", func(qRouter fiber.Router) {
			qRouter.Post("check", controller.QPromptCheck)
			qRouter.Post("validate_template", controller.QPromptValidateTpl)
			qRouter.Post("check_on_applicant", controller.QPromptApplicantCheck)
		})
	})
}

// @Summary Отправка промпта в ИИ
// @Tags ИИ
// @Description Отправка чистого промпта, без проверок и преобразований
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   prompt		formData	string 	true 	"Промпт"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/ai/prompt [post]
func (c *aiController) RunPrompt(ctx *fiber.Ctx) error {
	prompt := ctx.FormValue("prompt", "")
	if prompt == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не заполнен параметр prompt"))
	}

	id, err := promptcheckhandler.Instance.RunPrompt(prompt)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, err.Error())
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Результат отправки промпта в ИИ
// @Tags ИИ
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "ID"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/ai/{id}/result [get]
func (c *aiController) GetResult(ctx *fiber.Ctx) error {
	recID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	data, err := promptcheckhandler.Instance.ExecutionInfo(recID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, err.Error())
	}
	ctx.Set(helpers.HeaderLogIgnore, "true")
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(data))
}

// @Summary Проверка наличия выполняемых запросов
// @Tags ИИ
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "ID"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/ai/status [get]
func (c *aiController) GetStatus(ctx *fiber.Ctx) error {
	ctx.Set(helpers.HeaderLogIgnore, "true")
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(promptcheckhandler.Instance.Status()))
}

// @Summary Отправка промпта в ИИ (генерация вопросов шаг 1)
// @Tags ИИ
// @Description Отправка промпта в ИИ с проверкой ответа на соответствие ожидаемой структуре
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   prompt		formData	string 	true 	"Промпт"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/ai/questions_prompt/check [post]
func (c *aiController) QPromptCheck(ctx *fiber.Ctx) error {
	prompt := ctx.FormValue("prompt", "")
	if prompt == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не заполнен параметр prompt"))
	}
	id, err := promptcheckhandler.Instance.RunQuestionsPrompt(prompt)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, err.Error())
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Проверка шаблона промпта (генерация вопросов шаг 1 (для кандидата))
// @Tags ИИ
// @Description Проверка шаблона промпта, формирование конечного промпта из шаблона с данными по вакансии, кандидату и тп
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   prompt_template		formData	string 	true 	"Шиблон промпта, содержащий тэги {{}}"
// @Param   applicant_id		formData	string 	true 	"Ид кандидата"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/ai/questions_prompt/validate_template [post]
func (c *aiController) QPromptValidateTpl(ctx *fiber.Ctx) error {
	promptTpl := ctx.FormValue("prompt_template", "")
	if promptTpl == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не заполнен параметр prompt_template"))
	}
	applicantID := ctx.FormValue("applicant_id", "")
	if applicantID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не заполнен параметр applicant_id"))
	}
	spaceID := middleware.GetUserSpace(ctx)
	id, err := promptcheckhandler.Instance.ValidateQPromptTemplate(promptTpl, spaceID, applicantID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, err.Error())
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Отправка промпта в ИИ (генерация вопросов шаг 1 (для кандидата))
// @Tags ИИ
// @Description Заполение шаблона промпта данными кандидата/вакансии, отправка результата в ИИ, проверка ответа на соответствие ожидаемой структуре
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   prompt_template		formData	string 	true 	"Шиблон промпта, содержащий тэги {{}}"
// @Param   applicant_id		formData	string 	true 	"Ид кандидата"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/ai/questions_prompt/check_on_applicant [post]
func (c *aiController) QPromptApplicantCheck(ctx *fiber.Ctx) error {
	promptTpl := ctx.FormValue("prompt_template", "")
	if promptTpl == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не заполнен параметр prompt_template"))
	}
	applicantID := ctx.FormValue("applicant_id", "")
	if applicantID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не заполнен параметр applicant_id"))
	}
	spaceID := middleware.GetUserSpace(ctx)
	id, err := promptcheckhandler.Instance.RunQuestionsPromptOnApplicant(promptTpl, spaceID, applicantID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, err.Error())
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}
