package apiv1

import (
	"hr-tools-backend/controllers"
	"hr-tools-backend/lib/applicant"
	applicanthistoryhandler "hr-tools-backend/lib/applicant-history"
	filestorage "hr-tools-backend/lib/file-storage"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	applicantapimodels "hr-tools-backend/models/api/applicant"
	"io"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

type applicantApiController struct {
	controllers.BaseAPIController
}

func InitApplicantApiRouters(app *fiber.App) {
	controller := applicantApiController{}
	app.Route("applicant", func(router fiber.Router) {
		router.Get("doc/:id", controller.GetDoc) // скачать документ по id
		router.Post("list", controller.list)
		router.Post("", controller.create)
		router.Route(":id", func(idRouter fiber.Router) {
			idRouter.Post("upload-resume", controller.UploadResume) // загрузить резюме кандидата
			idRouter.Post("upload-doc", controller.UploadDoc)       // загрузить документ кандидата
			idRouter.Get("doc/list", controller.GetDocList)         // получить список документов кандидата
			idRouter.Get("resume", controller.GetResume)            // скачать резюме кандидата
			idRouter.Get("", controller.get)
			idRouter.Put("", controller.update)
			idRouter.Put("tag", controller.addTag)
			idRouter.Delete("tag", controller.delTag)
			idRouter.Put("change_stage", controller.changeStage)
			idRouter.Put("join", controller.join)
			idRouter.Put("isolate", controller.isolate)
			idRouter.Put("changes", controller.changes)
			idRouter.Put("note", controller.note)
		})
	})
}

// @Summary Загрузить резюме кандидата
// @Tags Кандидат
// @Description Загрузить резюме кандидата
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "ID кандидата"
// @Param   resume		formData	file 	true 	"file to upload"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/upload-resume [post]
func (c *applicantApiController) UploadResume(ctx *fiber.Ctx) error {
	applicantID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	file, err := ctx.FormFile("profile_photo")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	buffer, err := file.Open()
	if err != nil {
		log.WithError(err).Error("Ошибка при получении файла резюме")
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	defer buffer.Close()
	fileBody, err := io.ReadAll(buffer)
	if err != nil {
		log.WithError(err).Error("Ошибка при загрузке файла резюме")
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = filestorage.Instance.UploadResume(ctx.UserContext(), spaceID, applicantID, fileBody, file.Filename)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Загрузить документ кандидата
// @Tags Кандидат
// @Description Загрузить документ кандидата
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "ID кандидата"
// @Param   resume		formData	file 	true 	"file to upload"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/upload-doc [post]
func (c *applicantApiController) UploadDoc(ctx *fiber.Ctx) error {
	applicantID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	file, err := ctx.FormFile("profile_photo")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	buffer, err := file.Open()
	if err != nil {
		log.WithError(err).Error("Ошибка при получении файла документа")
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	defer buffer.Close()
	fileBody, err := io.ReadAll(buffer)
	if err != nil {
		log.WithError(err).Error("Ошибка при загрузке файла документа")
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = filestorage.Instance.UploadDoc(ctx.UserContext(), spaceID, applicantID, fileBody, file.Filename)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Скачать документ кандидата
// @Tags Кандидат
// @Description Скачать документ кандидата
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "ID документа"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/doc/{id} [get]
func (c *applicantApiController) GetDoc(ctx *fiber.Ctx) error {
	docID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	body, err := filestorage.Instance.GetFile(ctx.UserContext(), spaceID, docID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}

	return ctx.Send(body)
}

// @Summary Скачать резюме кандидата
// @Tags Кандидат
// @Description Скачать резюме кандидата
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "ID кандидата"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/resume [get]
func (c *applicantApiController) GetResume(ctx *fiber.Ctx) error {
	applicantID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	body, err := filestorage.Instance.GetResume(ctx.UserContext(), spaceID, applicantID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}

	return ctx.Send(body)
}

// @Summary Получить список документов кандидата
// @Tags Кандидат
// @Description Получить список документов кандидата
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "ID кандидата"
// @Success 200 {object} apimodels.Response{data=[]filesapimodels.FileView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/doc/list [get]
func (c *applicantApiController) GetDocList(ctx *fiber.Ctx) error {
	applicantID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	body, err := filestorage.Instance.GetDocList(ctx.UserContext(), applicantID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}

	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(body))
}

// @Summary Список
// @Tags Кандидат
// @Description Список
// @Param	body body	 applicantapimodels.ApplicantFilter	true	"request filter body"
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.ScrollerResponse{data=[]applicantapimodels.ApplicantView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/list [post]
func (c *applicantApiController) list(ctx *fiber.Ctx) error {
	var payload applicantapimodels.ApplicantFilter
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	list, rowCount, err := applicant.Instance.ListOfApplicant(spaceID, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewScrollerResponse(list, rowCount))
}

// @Summary Создание
// @Tags Кандидат
// @Description Создание
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 applicantapimodels.ApplicantData	true	"request body"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant [post]
func (c *applicantApiController) create(ctx *fiber.Ctx) error {
	var payload applicantapimodels.ApplicantData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	id, err := applicant.Instance.CreateApplicant(spaceID, userID, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Получение по ИД
// @Tags Кандидат
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=applicantapimodels.ApplicantViewExt}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id} [get]
func (c *applicantApiController) get(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := applicant.Instance.GetApplicant(spaceID, id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Обновление
// @Tags Кандидат
// @Description Обновление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 applicantapimodels.ApplicantData	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id} [put]
func (c *applicantApiController) update(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload applicantapimodels.ApplicantData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	err = applicant.Instance.UpdateApplicant(spaceID, id, userID, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Добавить тэг
// @Tags Кандидат
// @Description Добавить тэг
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	tag					query 	string							false		 "добавляемый Тег"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/tag [put]
func (c *applicantApiController) addTag(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	tag := ctx.Query("tag", "")
	if tag == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не указан тэг"))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	err = applicant.Instance.ApplicantAddTag(spaceID, id, userID, tag)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Удалить тэг
// @Tags Кандидат
// @Description Удалить тэг
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	tag					query 	string							false		 "удаляемый Тег"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/tag [delete]
func (c *applicantApiController) delTag(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	tag := ctx.Query("tag", "")
	if tag == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не указан тэг"))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	err = applicant.Instance.ApplicantRemoveTag(spaceID, id, userID, tag)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary (Дубли) Объединение кандидатов
// @Tags Кандидат
// @Description (Дубли) Объединение кандидатов
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	duplicate_id		query 	string							true		 "Идентификатор кандидата - дубликата, который будет перенесен в архив"
// @Param   id          		path    string  				    	true         "Идентификатор основного кандидата"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/join [put]
func (c *applicantApiController) join(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	duplicateID := ctx.Query("duplicate_id", "")
	if duplicateID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не указан идентификатор кандидата - дубликата"))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	err = applicant.Instance.ResolveDuplicate(spaceID, id, duplicateID, userID, true)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary (Дубли) Пометить кандидатов как разных
// @Tags Кандидат
// @Description (Дубли) Пометить кандидатов как разных
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	duplicate_id		query 	string							true		 "Идентификатор кандидата - дубликата, который помечается как не дубликат"
// @Param   id          		path    string  				    	true         "Идентификатор основного кандидата"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/isolate [put]
func (c *applicantApiController) isolate(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	duplicateID := ctx.Query("duplicate_id", "")
	if duplicateID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не указан идентификатор кандидата - дубликата"))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	err = applicant.Instance.ResolveDuplicate(spaceID, id, duplicateID, userID, false)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Перевести на другой этап подбора
// @Tags Кандидат
// @Description Перевести на другой этап подбора
// @Param   Authorization		header	string	true	"Authorization token"
// @Param	stage_id			query 	string							true		 "Идентификатор этапа на который необходимо перевести кандидата"
// @Param   id          		path    string  				    	true         "Идентификатор кандидата"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/change_stage [put]
func (c *applicantApiController) changeStage(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	stageID := ctx.Query("stage_id", "")
	if stageID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("не указан идентификатор этапа"))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	err = applicant.Instance.ChangeStage(spaceID, userID, id, stageID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Лог действий
// @Tags Кандидат
// @Description Лог действий
// @Param   Authorization		header	string	true	"Authorization token"
// @Param   id          		path    string  true    "Идентификатор кандидата"
// @Param	body body	 applicantapimodels.ApplicantHistoryFilter	true	"request filter"
// @Success 200 {object} apimodels.ScrollerResponse{data=[]applicantapimodels.ApplicantHistoryView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/changes [put]
func (c *applicantApiController) changes(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	var payload applicantapimodels.ApplicantHistoryFilter
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	data, rowCount, err := applicanthistoryhandler.Instance.List(spaceID, id, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewScrollerResponse(data, rowCount))
}

// @Summary Добавить заметку о кандидате
// @Tags Кандидат
// @Description Добавить заметку о кандидате
// @Param   Authorization	 header		string	true	"Authorization token"
// @Param   id          	 path    	string  true    "Идентификатор кандидата"
// @Param	body			 body	 	applicantapimodels.ApplicantNote	true	"request data"
// @Success 200 {object} apimodels.ScrollerResponse{data=[]applicantapimodels.ApplicantHistoryView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/note [put]
func (c *applicantApiController) note(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	var payload applicantapimodels.ApplicantNote
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	userID := middleware.GetUserID(ctx)
	spaceID := middleware.GetUserSpace(ctx)
	err = applicanthistoryhandler.Instance.SaveNote(spaceID, id, userID, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}
