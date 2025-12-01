package apiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	licencehandler "hr-tools-backend/lib/licence"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	licenseapimodels "hr-tools-backend/models/api/license"
)

type billingApiController struct {
	controllers.BaseAPIController
}

func InitBillingApiRouters(app *fiber.App) {
	controller := billingApiController{}
	app.Route("billing", func(router fiber.Router) {
		router.Get("license", controller.getLicense)
		router.Get("license/renew", controller.getRenew)
		router.Post("license/renew", controller.renew)
		// router.Post("payment-intent", controller.paymentIntent)//TODO impl
		// router.Post("payment/webhook", controller.paymentWebhook)//TODO impl
	})
}

// @Summary Данные лицензии
// @Tags Лицензия
// @Description Данные лицензии
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=licenseapimodels.License}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/billing/license [get]
func (c *billingApiController) getLicense(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	resp, err := licencehandler.Instance.GetSpaceLicense(spaceID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения данных по лицензии")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Информация о продлении
// @Tags Лицензия
// @Description Информация о продлении
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=licenseapimodels.LicenseRenewInfo}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/billing/license/renew [get]
func (c *billingApiController) getRenew(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)

	resp, err := licencehandler.Instance.GetRenewSpaceLicense(spaceID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения данных по продлению лицензии")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Создать черновик продления
// @Tags Лицензия
// @Description Создать черновик продления
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 licenseapimodels.LicenseRenew	true	"request body"
// @Success 200 {object} apimodels.Response{data=licenseapimodels.LicenseRenewResponse}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/billing/license/renew [post]
func (c *billingApiController) renew(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)

	var payload licenseapimodels.LicenseRenew
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	resp, err := licencehandler.Instance.RenewSpaceLicense(spaceID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения данных по лицензии")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}
