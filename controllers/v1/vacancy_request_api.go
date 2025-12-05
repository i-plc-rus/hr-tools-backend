package apiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	vacancyreqhandler "hr-tools-backend/lib/vacancy-req"
	"hr-tools-backend/middleware"
	"hr-tools-backend/models"
	apimodels "hr-tools-backend/models/api"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"
)

type vacancyReqApiController struct {
	controllers.BaseAPIController
}

func InitVacancyRequestApiRouters(app *fiber.App) {
	controller := vacancyReqApiController{}
	app.Route("vacancy_request", func(router fiber.Router) {
		router.Use(middleware.LicenseRequired())

		router.Post("list", controller.list)
		router.Post("", controller.create)
		router.Route(":id", func(idRoute fiber.Router) {
			idRoute.Put("", controller.update)
			idRoute.Get("", controller.get)
			idRoute.Delete("", controller.delete)
			idRoute.Put("pin", controller.pin)
			idRoute.Put("favorite", controller.favorite)
			idRoute.Put("on_create", controller.onCreate)     // перевести шаблон на статус заявка создана
			idRoute.Put("on_approval", controller.onApproval) // на согласование
			idRoute.Put("publish", controller.publish)        // создать вакансию
			idRoute.Put("cancel", controller.cancel)          // отменить
			idRoute.Post("comment", controller.addComment)
			idRoute.Route("approvals", func(approvals fiber.Router) {
				approvals.Get("", controller.getApprovals)
				approvals.Put("", controller.saveApprovals)
				approvals.Route(":taskId", func(taskRoute fiber.Router) {
					taskRoute.Post("approve", controller.approve)                // согласовать
					taskRoute.Post("request_changes", controller.requestChanges) // на доработку
					taskRoute.Post("reject", controller.reject)                  // отклонить
				})
			})
			idRoute.Get("approval_history", controller.getApprovalHistory)
		})
	})
}

// @Summary Создание
// @Tags Заявка
// @Description Создание
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 vacancyapimodels.VacancyRequestCreateData	true	"request body"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request [post]
func (c *vacancyReqApiController) create(ctx *fiber.Ctx) error {
	var payload vacancyapimodels.VacancyRequestCreateData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	id, hMsg, err := vacancyreqhandler.Instance.Create(spaceID, userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка создания заявки")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Обновление
// @Tags Заявка
// @Description Обновление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 vacancyapimodels.VacancyRequestEditData	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id} [put]
func (c *vacancyReqApiController) update(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload vacancyapimodels.VacancyRequestEditData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	hMsg, err := vacancyreqhandler.Instance.Update(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка обновления заявки")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение по ИД
// @Tags Заявка
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=vacancyapimodels.VacancyRequestView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id} [get]
func (c *vacancyReqApiController) get(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := vacancyreqhandler.Instance.GetByID(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения заявки")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Удаление
// @Tags Заявка
// @Description Удаление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id} [delete]
func (c *vacancyReqApiController) delete(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = vacancyreqhandler.Instance.Delete(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления заявки")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Закрепить
// @Tags Заявка
// @Description Закрепить
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	set					query 	bool							false		 "выбрано/не выбрано"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/pin [put]
func (c *vacancyReqApiController) pin(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	isSet := ctx.QueryBool("set", false)

	userID := middleware.GetUserID(ctx)
	err = vacancyreqhandler.Instance.ToPin(id, userID, isSet)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка закрепления заявки")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary В избранное
// @Tags Заявка
// @Description В избранное
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	set					query 	bool							false		 "выбрано/не выбрано"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/favorite [put]
func (c *vacancyReqApiController) favorite(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	isSet := ctx.QueryBool("set", false)
	userID := middleware.GetUserID(ctx)
	err = vacancyreqhandler.Instance.ToFavorite(id, userID, isSet)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка добавления заявки в избранное")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Список
// @Tags Заявка
// @Description Список
// @Param	body body	 vacancyapimodels.VrFilter	true	"request filter body"
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.ScrollerResponse{data=[]vacancyapimodels.VacancyRequestView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/list [post]
func (c *vacancyReqApiController) list(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	var payload vacancyapimodels.VrFilter
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	userID := middleware.GetUserID(ctx)
	list, rowCount, err := vacancyreqhandler.Instance.List(spaceID, userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка заявок")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewScrollerResponse(list, rowCount))
}

// @Summary Перевести шаблон на статус заявка создана
// @Tags Заявка
// @Description Перевести шаблон на статус заявка создана
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/on_create [put]
func (c *vacancyReqApiController) onCreate(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	hMsg, err := vacancyreqhandler.Instance.ChangeStatus(spaceID, id, userID, models.VRStatusCreated)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка перевода шаблока в заявку")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Отправить на согласование
// @Tags Заявка
// @Description Отправить на согласование
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/on_approval [put]
func (c *vacancyReqApiController) onApproval(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	hMsg, err := vacancyreqhandler.Instance.ChangeStatus(spaceID, id, userID, models.VRStatusInApproval)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка отравки заявки на согласование")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Отменить заявку
// @Tags Заявка
// @Description Отменить заявку
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/cancel [put]
func (c *vacancyReqApiController) cancel(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	hMsg, err := vacancyreqhandler.Instance.ChangeStatus(spaceID, id, userID, models.VRStatusCancelled)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка отмены заявки")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Создать вакансию
// @Tags Заявка
// @Description Создать вакансию
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/publish [put]
func (c *vacancyReqApiController) publish(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	hMsg, err := vacancyreqhandler.Instance.CreateVacancy(spaceID, id, userID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка создания вакансии")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Добавить комментарий к заявке
// @Tags Заявка
// @Description Добавить комментарий к заявке
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 vacancyapimodels.Comment	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/comment [post]
func (c *vacancyReqApiController) addComment(ctx *fiber.Ctx) error {
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
	err = vacancyreqhandler.Instance.AddComment(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка добавления комментария")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}
