package dict

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	jobtitleprovider "hr-tools-backend/lib/dicts/job-title"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	dictapimodels "hr-tools-backend/models/api/dict"
)

type jobTitleDictApiController struct {
	controllers.BaseAPIController
}

func InitJobTitleDictApiRouters(app *fiber.App) {
	controller := jobTitleDictApiController{}
	app.Route("job_title", func(router fiber.Router) {
		router.Post("find", controller.jobTitleFindByName)
		router.Use(middleware.SpaceAdminRequired())
		router.Post("", controller.jobTitleCreate)
		router.Put(":id", controller.jobTitleUpdate)
		router.Get(":id", controller.jobTitleGet)
		router.Delete(":id", controller.jobTitleDelete)
	})
}

// @Summary Создание
// @Tags Справочник. Штатные должности
// @Description Создание
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.JobTitleData	true	"request body"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/job_title [post]
func (c *jobTitleDictApiController) jobTitleCreate(ctx *fiber.Ctx) error {
	var payload dictapimodels.JobTitleData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	id, err := jobtitleprovider.Instance.Create(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка создания записи в справочнике штатных должностей")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Обновление
// @Tags Справочник. Штатные должности
// @Description Обновление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.JobTitleData	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/job_title/{id} [put]
func (c *jobTitleDictApiController) jobTitleUpdate(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload dictapimodels.JobTitleData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = jobtitleprovider.Instance.Update(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка обновления данных в справочнике штатных должностей")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение по ИД
// @Tags Справочник. Штатные должности
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=dictapimodels.CompanyView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/job_title/{id} [get]
func (c *jobTitleDictApiController) jobTitleGet(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := jobtitleprovider.Instance.Get(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения записи из справочника штатных должностей")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Удаление
// @Tags Справочник. Штатные должности
// @Description Удаление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/job_title/{id} [delete]
func (c *jobTitleDictApiController) jobTitleDelete(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = jobtitleprovider.Instance.Delete(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления данных из справочника штатных должностей")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Поиск по названию
// @Tags Справочник. Штатные должности
// @Description Поиск по названию
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 dictapimodels.JobTitleData	true	"request body"
// @Success 200 {object} apimodels.Response{data=[]dictapimodels.CompanyView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/dict/job_title/find [post]
func (c *jobTitleDictApiController) jobTitleFindByName(ctx *fiber.Ctx) error {
	var payload dictapimodels.JobTitleData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	list, err := jobtitleprovider.Instance.FindByName(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка из справочника штатных должностей")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}
