package apiv1

import (
	"hr-tools-backend/controllers"
	"hr-tools-backend/lib/survey"
	vacancyhandler "hr-tools-backend/lib/vacancy"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	surveyapimodels "hr-tools-backend/models/api/survey"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"

	"github.com/gofiber/fiber/v2"
)

type vacancyApiController struct {
	controllers.BaseAPIController
}

func InitVacancyApiRouters(app *fiber.App) {
	controller := vacancyApiController{}
	app.Route("vacancy", func(router fiber.Router) {
		router.Use(middleware.LicenseRequired())
		
		router.Post("list", controller.list)
		router.Post("", controller.create)
		router.Route(":id", func(idRoute fiber.Router) {
			idRoute.Put("", controller.update)
			idRoute.Get("", controller.get)
			idRoute.Delete("", controller.delete)
			idRoute.Put("pin", controller.pin)
			idRoute.Put("favorite", controller.favorite)
			idRoute.Put("change_status", controller.changeStatus)
			idRoute.Post("comment", controller.addComment)
			idRoute.Route("stage", func(stageRoute fiber.Router) {
				stageRoute.Post("list", controller.stageList)
				stageRoute.Post("", controller.stageCreate)
				stageRoute.Delete("", controller.stageDelete)
				stageRoute.Put("change_order", controller.stageChangeOrder)
			})
			idRoute.Route("team", func(teamRoute fiber.Router) {
				teamRoute.Post("list", controller.teamList)
				teamRoute.Post("users_list", controller.usersList)
				teamRoute.Route(":user_id", func(userIDRoute fiber.Router) {
					userIDRoute.Put("invite", controller.inviteToTeam)
					userIDRoute.Put("set_as_responsible", controller.setAsResponsible)
					userIDRoute.Put("exclude", controller.excludeFromTeam)
				})
			})
			idRoute.Route("survey", func(surveyRoute fiber.Router) {
				surveyRoute.Get("", controller.getSurvey)
				surveyRoute.Post("", controller.saveSurvey)
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
	id, hMsg, err := vacancyhandler.Instance.Create(spaceID, userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка создания вакансии")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
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

// @Summary Добавить комментарий к вакансии
// @Tags Вакансия
// @Description Добавить комментарий к вакансии
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 vacancyapimodels.Comment	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/comment [post]
func (c *vacancyApiController) addComment(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload vacancyapimodels.Comment
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = vacancyhandler.Instance.AddComment(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка добавления комментария")
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

// @Summary Команда
// @Tags Вакансия
// @Description Команда
// @Param   Authorization       header      string  true    "Authorization token"
// @Param   id                  path    	string  true    "vacancy ID"
// @Success 200 {object} apimodels.Response{data=[]vacancyapimodels.TeamPerson}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/team/list [post]
func (c *vacancyApiController) teamList(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := vacancyhandler.Instance.GetTeam(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения данных по команде")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Команда. Список пользователей в системе
// @Tags Вакансия
// @Description Команда. Список пользователей в системе
// @Param   Authorization       header      string  						true    "Authorization token"
// @Param	body 				body	 	vacancyapimodels.PersonFilter	true	"request filter body"
// @Param   id                  path    	string                          true     "vacancy ID"
// @Success 200 {object} apimodels.Response{data=[]vacancyapimodels.TeamPerson}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/team/users_list [post]
func (c *vacancyApiController) usersList(ctx *fiber.Ctx) error {
	var payload vacancyapimodels.PersonFilter
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := vacancyhandler.Instance.UsersList(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения список пользователей в спейсе")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Команда. Пригласить участника
// @Tags Вакансия
// @Description Команда. Пригласить участника
// @Param   Authorization       header      string  true    "Authorization token"
// @Param   id                  path    string                          true         "vacancy ID"
// @Param   user_id             path    string                          true         "user ID"
// @Success 200 {object} apimodels.Response{data=[]vacancyapimodels.TeamPerson}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/team/{user_id}/invite [put]
func (c *vacancyApiController) inviteToTeam(ctx *fiber.Ctx) error {
	vacancyID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	userID, err := c.GetIDByKey(ctx, "user_id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := vacancyhandler.Instance.InviteToTeam(nil, spaceID, vacancyID, userID, false)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка приглашения участника в команду")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Команда. Исключить участника
// @Tags Вакансия
// @Description Команда. Исключить участника
// @Param   Authorization       header      string  true    	"Authorization token"
// @Param   id                  path    	string  true         "vacancy ID"
// @Param   user_id             path    	string  true         "user ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/team/{user_id}/exclude [put]
func (c *vacancyApiController) excludeFromTeam(ctx *fiber.Ctx) error {
	vacancyID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	userID, err := c.GetIDByKey(ctx, "user_id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	hMsg, err := vacancyhandler.Instance.ExecuteFromTeam(spaceID, vacancyID, userID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка исключения участника из команды")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Команда. Сделать ответственным
// @Tags Вакансия
// @Description Команда. Сделать ответственным
// @Param   Authorization       header      string  true    "Authorization token"
// @Param   id                  path    	string  true    "vacancy ID"
// @Param   user_id             path   		string  true    "user ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/team/{user_id}/set_as_responsible [put]
func (c *vacancyApiController) setAsResponsible(ctx *fiber.Ctx) error {
	vacancyID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	userID, err := c.GetIDByKey(ctx, "user_id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = vacancyhandler.Instance.SetAsResponsible(spaceID, vacancyID, userID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка назначения участника ответственным по вакансии")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение анкеты HR (генерация, в случае отсутствия)
// @Tags Анкета HR по вакансии
// @Description Получение анкеты HR (генерация, в случае отсутствия)
// @Param   Authorization		header	string	true	"Authorization token"
// @Param   id          		path    string  true    "идентификатор вакансии"
// @Success 200 {object} apimodels.Response{data=surveyapimodels.HRSurveyView}
// @Failure 400 {object} apimodels.Response
// @Failure 404 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/survey [get]
func (c *vacancyApiController) getSurvey(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, hMsg, err := survey.Instance.GetHRSurvey(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения анкеты по вакансии")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	if resp == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(apimodels.NewError("Анкета отсутсвует"))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Сохранение анкеты HR (перегенерация вопросов в случае если вопрос не подходит)
// @Tags Анкета HR по вакансии
// @Description Сохранение анкеты HR (перегенерация вопросов в случае если вопрос не подходит)
// @Param   Authorization		header	string	true	"Authorization token"
// @Param   id          		path    string  true         "rec ID"
// @Param	body body	 surveyapimodels.HRSurvey	true	"request body"
// @Success 200 {object} apimodels.Response{data=surveyapimodels.HRSurveyView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy/{id}/survey [post]
func (c *vacancyApiController) saveSurvey(ctx *fiber.Ctx) error {
	var payload surveyapimodels.HRSurvey
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := survey.Instance.SaveHRSurvey(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка сохранения анкеты по вакансии")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}
