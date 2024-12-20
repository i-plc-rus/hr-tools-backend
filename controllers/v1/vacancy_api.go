package apiv1

import (
	"hr-tools-backend/controllers"
	vacancyhandler "hr-tools-backend/lib/vacancy"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"

	"github.com/gofiber/fiber/v2"
)

type vacancyApiController struct {
	controllers.BaseAPIController
}

func InitVacancyApiRouters(app *fiber.App) {
	controller := vacancyApiController{}
	app.Route("vacancy", func(router fiber.Router) {
		router.Post("list", controller.list)
		router.Post("", controller.create)
		router.Route(":id", func(idRoute fiber.Router) {
			idRoute.Put("", controller.update)
			idRoute.Get("", controller.get)
			idRoute.Delete("", controller.delete)
			idRoute.Put("pin", controller.pin)
			idRoute.Put("favorite", controller.favorite)
			idRoute.Put("change_status", controller.changeStatus)
			idRoute.Route("stage", func(stageRoute fiber.Router) {
				stageRoute.Post("list", controller.stageList)
				stageRoute.Post("", controller.stageCreate)
				stageRoute.Delete("", controller.stageDelete)
				stageRoute.Put("change_order", controller.stageChangeOrder)
			})
		})
	})
}

// @Summary Создание
// @Tags Вакансия
// @Description Создание
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 vacancyapimodels.VacancyData	true	"request body"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy [post]
func (c *vacancyApiController) create(ctx *fiber.Ctx) error {
	var payload vacancyapimodels.VacancyData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(false); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	id, err := vacancyhandler.Instance.Create(spaceID, userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка создания вакансии")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Обновление
// @Tags Вакансия
// @Description Обновление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 vacancyapimodels.VacancyData	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id} [put]
func (c *vacancyApiController) update(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload vacancyapimodels.VacancyData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(false); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = vacancyhandler.Instance.Update(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка изменения вакансии")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение по ИД
// @Tags Вакансия
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=vacancyapimodels.VacancyView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id} [get]
func (c *vacancyApiController) get(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := vacancyhandler.Instance.GetByID(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения вакансии")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Удаление
// @Tags Вакансия
// @Description Удаление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id} [delete]
func (c *vacancyApiController) delete(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = vacancyhandler.Instance.Delete(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления вакансии")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Список
// @Tags Вакансия
// @Description Список
// @Param	body body	 vacancyapimodels.VacancyFilter	true	"request filter body"
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.ScrollerResponse{data=[]vacancyapimodels.VacancyView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/list [post]
func (c *vacancyApiController) list(ctx *fiber.Ctx) error {
	var payload vacancyapimodels.VacancyFilter
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	list, rowCount, err := vacancyhandler.Instance.List(spaceID, userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка вакансий")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewScrollerResponse(list, rowCount))
}

// @Summary Закрепить
// @Tags Вакансия
// @Description Закрепить
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	set					query 	bool							false		 "выбрано/не выбрано"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/pin [put]
func (c *vacancyApiController) pin(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	isSet := ctx.QueryBool("set", false)

	userID := middleware.GetUserID(ctx)
	err = vacancyhandler.Instance.ToPin(id, userID, isSet)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка закрепления вакансии")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary В избранное
// @Tags Вакансия
// @Description В избранное
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	set					query 	bool							false		 "выбрано/не выбрано"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/favorite [put]
func (c *vacancyApiController) favorite(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	isSet := ctx.QueryBool("set", false)
	userID := middleware.GetUserID(ctx)
	err = vacancyhandler.Instance.ToFavorite(id, userID, isSet)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка добавления вакансии в избранное")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Изменение статуса вакансии
// @Tags Вакансия
// @Description Изменение статуса вакансии
// @Param   Authorization		header		string					true	"Authorization token"
// @Param   id          		path    string  				    true    "rec ID"
// @Param	body body	 vacancyapimodels.StatusChangeRequest	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/change_status [put]
func (c *vacancyApiController) changeStatus(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	var payload vacancyapimodels.StatusChangeRequest
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	err = vacancyhandler.Instance.StatusChange(spaceID, id, userID, payload.Status)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка изменения статуса вакансии")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Список этапов подбора
// @Tags Вакансия
// @Description Список этапов подбора
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.ScrollerResponse{data=[]vacancyapimodels.SelectionStageView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/stage/list [post]
func (c *vacancyApiController) stageList(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	list, err := vacancyhandler.Instance.StageList(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения этапов подбора")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}

// @Summary Изменение порядка этапов подбора
// @Tags Вакансия
// @Description Изменение порядка этапов подбора
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    true         "rec ID"
// @Param	stage_id			query 	string						false		 "идентификатор этапа"
// @Param	stage_order			query 	int							false		 "новый порядковый номер"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/stage/change_order [put]
func (c *vacancyApiController) stageChangeOrder(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	stageID := ctx.Query("stage_id", "")
	if stageID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не указан этап подбора вакансии"))
	}
	newOrder := ctx.QueryInt("stage_order", -1)
	if newOrder <= 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не указан новый порядковый номер этапа подбора вакансии"))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = vacancyhandler.Instance.StageChangeOrder(spaceID, id, stageID, newOrder)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка изменения порядка этапов подбора")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Добавить этап подбора
// @Tags Вакансия
// @Description Добавить этап подбора
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    true         "rec ID"
// @Param	body body	 vacancyapimodels.SelectionStageAdd	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/stage [post]
func (c *vacancyApiController) stageCreate(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload vacancyapimodels.SelectionStageAdd
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = vacancyhandler.Instance.StageCreate(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка добавления этапа подбора")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Удаление этап подбора
// @Tags Вакансия
// @Description Удаление этап подбора
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Param	stage_id			query 	string						false		 "идентификатор этапа"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/stage [delete]
func (c *vacancyApiController) stageDelete(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	stageID := ctx.Query("stage_id", "")
	if stageID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не указан этап подбора вакансии"))
	}
	hMsg, err := vacancyhandler.Instance.StageDelete(spaceID, id, stageID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления этапа подбора")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}
