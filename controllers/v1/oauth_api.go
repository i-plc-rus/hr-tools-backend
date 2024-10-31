package apiv1

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/controllers"
	hhclient "hr-tools-backend/lib/external-services/hh"
)

type oAuthApiController struct {
	controllers.BaseAPIController
}

func InitOAuthApiRouters(app *fiber.App) {
	controller := oAuthApiController{}
	app.Route("oauth", func(router fiber.Router) {
		router.Route("callback", func(callback fiber.Router) {
			callback.Get("hh", controller.hhCallBack)
			callback.Get("avito", controller.avitoCallBack)
		})
	})
}

// @Summary Аутентификация HH
// @Tags Аутентификация OAuth
// @Description Аутентификация HH
// @Param   state     query     string  				true   "space ID"
// @Param   code      query    string  				    true       "authorization_code"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/oauth/callback/hh [get]
func (c *oAuthApiController) hhCallBack(ctx *fiber.Ctx) error {

	spaceID := ctx.Query("state", "")
	code := ctx.Query("code", "")
	log.Infof("hhCallBack (code: %v) (state: %v)", code, spaceID)

	if code != "" && spaceID != "" {
		go hhclient.Instance.RequestToken(spaceID, code)
	}

	_, err := ctx.Status(fiber.StatusOK).Write([]byte("ok"))
	return err
}

// @Summary Аутентификация Avito
// @Tags Аутентификация OAuth
// @Description Аутентификация Avito
// @Param   state     query     string  				true   "space ID"
// @Param   code      query     string  				true       "authorization_code"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/oauth/callback/avito [get]
func (c *oAuthApiController) avitoCallBack(ctx *fiber.Ctx) error {
	spaceID := ctx.Query("state", "")
	code := ctx.Query("code", "")
	log.Infof("avitoCallBack (code: %v) (state: %v)", code, spaceID)

	if code != "" && spaceID != "" {
		go hhclient.Instance.RequestToken(spaceID, code)
	}

	_, err := ctx.Status(fiber.StatusOK).Write([]byte("ok"))
	return err
}
