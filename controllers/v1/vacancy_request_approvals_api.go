package apiv1

import (
	aprovaltaskhandler "hr-tools-backend/lib/aproval-task"
	vacancyreqhandler "hr-tools-backend/lib/vacancy-req"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	vacancyapimodels "hr-tools-backend/models/api/vacancy"

	"github.com/gofiber/fiber/v2"
)

// @Summary Обновление цепочки согласования
// @Tags Согласование заявок
// @Description Обновление цепочки согласования
// @Param   Authorization		header	string								true	"Authorization token"
// @Param	body 				body	[]vacancyapimodels.ApprovalTaskView	true	"request body"
// @Param   id          		path    string  				    		true    "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/approvals [get]
func (c *vacancyReqApiController) getApprovals(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	result, err := aprovaltaskhandler.Instance.Get(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка задач согласования")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(result))
}

// @Summary Обновление цепочки согласования
// @Tags Согласование заявок
// @Description Обновление цепочки согласования
// @Param   Authorization		header	string								true	"Authorization token"
// @Param	body 				body	vacancyapimodels.ApprovalTasks	true	"request body"
// @Param   id          		path    string  				    		true    "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/approvals [put]
func (c *vacancyReqApiController) saveApprovals(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload vacancyapimodels.ApprovalTasks
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	hMsg, err := aprovaltaskhandler.Instance.Save(spaceID, id, payload.ApprovalTasks)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка обновления цепочки согласования")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Согласовать
// @Tags Согласование заявок
// @Description Согласовать
// @Param   Authorization		header	string								true	"Authorization token"
// @Param   id          		path    string  				    		true    "rec ID"
// @Param   taskId          	path    string  				    		true    "task rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/approvals/{taskId}/approve [post]
func (c *vacancyReqApiController) approve(ctx *fiber.Ctx) error {
	requestID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	taskID, err := c.GetIDByKey(ctx, "taskId")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	hMsg, err := vacancyreqhandler.Instance.Approve(spaceID, requestID, taskID, userID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка согласования заявки")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary На доработку
// @Tags Согласование заявок
// @Description На доработку
// @Param   Authorization		header	string										true	"Authorization token"
// @Param	body 				body	[]vacancyapimodels.ApprovalRequestChanges	true	"request body"
// @Param   id          		path    string  				    				true    "rec ID"
// @Param   taskId          	path    string  				    				true    "task rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/approvals/{taskId}/request_changes [post]
func (c *vacancyReqApiController) requestChanges(ctx *fiber.Ctx) error {
	requestID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	taskID, err := c.GetIDByKey(ctx, "taskId")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	var payload vacancyapimodels.ApprovalRequestChanges
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	hMsg, err := vacancyreqhandler.Instance.RequestChanges(spaceID, requestID, taskID, userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка отправки заявки на доработку")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Отклонить
// @Tags Согласование заявок
// @Description Отклонить
// @Param   Authorization		header	string										true	"Authorization token"
// @Param	body 				body	vacancyapimodels.ApprovalReject			true	"request body"
// @Param   id          		path    string  				    				true    "rec ID"
// @Param   taskId          	path    string  				    				true    "task rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/approvals/{taskId}/reject [post]
func (c *vacancyReqApiController) reject(ctx *fiber.Ctx) error {
	requestID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	taskID, err := c.GetIDByKey(ctx, "taskId")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	var payload vacancyapimodels.ApprovalReject
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	hMsg, err := vacancyreqhandler.Instance.Reject(spaceID, requestID, taskID, userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка отклонения заявки")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary История согласования
// @Tags Согласование заявок
// @Description История согласования
// @Param   Authorization		header	string								true	"Authorization token"
// @Param   id          		path    string  				    		true    "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/vacancy_request/{id}/approval_history [get]
func (c *vacancyReqApiController) getApprovalHistory(ctx *fiber.Ctx) error {
	requestID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	result, err := aprovaltaskhandler.Instance.History(spaceID, requestID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения история согласования")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(result))
}
