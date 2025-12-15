package apiv1

import (
	"fmt"
	"hr-tools-backend/controllers"
	"hr-tools-backend/lib/analytics"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	applicantapimodels "hr-tools-backend/models/api/applicant"
	"time"

	"github.com/gofiber/fiber/v2"
)

type analyticsApiController struct {
	controllers.BaseAPIController
}

func InitAnalyticsApiRouters(app *fiber.App) {
	controller := analyticsApiController{}
	app.Route("analytics", func(router fiber.Router) {
		router.Use(middleware.LicenseRequired())
		router.Use(middleware.RbacMiddleware())
		router.Put("source", controller.source)
		router.Put("source_export", controller.sourceExport)
	})
}

// @Summary Источники кандидатов
// @Tags Аналитика
// @Description Источники кандидатов
// @Param   Authorization		header	string	true	"Authorization token"
// @Param	body body	 applicantapimodels.ApplicantFilter	true	"request body"
// @Success 200 {object} apimodels.Response{data=applicantapimodels.ApplicantSourceData}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/analytics/source [put]
func (c *analyticsApiController) source(ctx *fiber.Ctx) error {
	var payload applicantapimodels.ApplicantFilter
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	payload.Limit = 1000
	data, err := analytics.Instance.Source(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения аналитики по источникам кандидатов")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(data))
}

// @Summary Источники кандидатов. Выгрузить в Excel
// @Tags Аналитика
// @Description Источники кандидатов. Выгрузить в Excel
// @Param   Authorization		header	string	true	"Authorization token"
// @Param	body body	applicantapimodels.ApplicantFilter	true	"request body"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/analytics/source_export [put]
func (c *analyticsApiController) sourceExport(ctx *fiber.Ctx) error {
	var payload applicantapimodels.ApplicantFilter
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	payload.Limit = 1000
	data, err := analytics.Instance.SourceExportToXls(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения аналитики по источникам кандидатов для выгрузки в Excel")
	}
	fileName := fmt.Sprintf("applicants-%v.xlsx", time.Now().Format("20060102-150405"))
	ctx.Set("Content-Type", "application/vnd.ms-excel")
	ctx.Set(fiber.HeaderContentDisposition, `attachment; filename="`+fileName+`"`)
	return ctx.SendStream(data)
}
