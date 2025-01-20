package apiv1

import (
	"fmt"
	"hr-tools-backend/controllers"
	"hr-tools-backend/lib/applicant"
	applicanthistoryhandler "hr-tools-backend/lib/applicant-history"
	applicantdict "hr-tools-backend/lib/dicts/applicant"
	filestorage "hr-tools-backend/lib/file-storage"
	messagetemplate "hr-tools-backend/lib/message-template"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	applicantapimodels "hr-tools-backend/models/api/applicant"
	dbmodels "hr-tools-backend/models/db"
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
)

type applicantApiController struct {
	controllers.BaseAPIController
}

func InitApplicantApiRouters(app *fiber.App) {
	controller := applicantApiController{}
	app.Route("applicant", func(router fiber.Router) {
		router.Get("doc/:id", controller.GetDoc)       // скачать документ по id
		router.Delete("doc/:id", controller.deleteDoc) // удлить документ по id
		router.Post("list", controller.list)
		router.Post("reject_list", controller.rejectList)
		router.Post("", controller.create)
		router.Route("multi-actions", func(mRouter fiber.Router) {
			mRouter.Put("reject", controller.multiReject)
			mRouter.Put("change_stage", controller.multiChangeStage)
			mRouter.Put("export_xls", controller.multiExportXls)
			mRouter.Put("send_email", controller.multiSendMail)
		})
		router.Route(":id", func(idRouter fiber.Router) {
			idRouter.Post("upload-resume", controller.UploadResume) // загрузить резюме кандидата
			idRouter.Post("upload-doc", controller.UploadDoc)       // загрузить документ кандидата
			idRouter.Post("upload-photo", controller.uploadPhoto)   // загрузить фото кандидата
			idRouter.Delete("photo", controller.deletePhoto)
			idRouter.Delete("resume", controller.deleteResume)
			idRouter.Get("doc/list", controller.GetDocList) // получить список документов кандидата
			idRouter.Get("resume", controller.GetResume)    // скачать резюме кандидата
			idRouter.Get("photo", controller.getPhoto)      // скачать фото кандидата
			idRouter.Get("", controller.get)
			idRouter.Put("", controller.update)
			idRouter.Put("tag", controller.addTag)
			idRouter.Delete("tag", controller.delTag)
			idRouter.Put("change_stage", controller.changeStage)
			idRouter.Put("join", controller.join)
			idRouter.Put("isolate", controller.isolate)
			idRouter.Put("changes", controller.changes)
			idRouter.Put("note", controller.note)
			idRouter.Put("reject", controller.reject)
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

	file, err := ctx.FormFile("resume")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	logger := c.GetLogger(ctx)
	buffer, err := file.Open()
	if err != nil {
		return c.SendError(ctx, logger, err, "Ошибка при получении файла резюме")
	}
	defer buffer.Close()
	fileBody, err := io.ReadAll(buffer)
	if err != nil {
		return c.SendError(ctx, logger, err, "Ошибка при загрузке файла резюме")
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = filestorage.Instance.Upload(ctx.UserContext(), spaceID, applicantID, fileBody, file.Filename, dbmodels.ApplicantResume)
	if err != nil {
		return c.SendError(ctx, logger, err, "Ошибка сохранения файла резюме")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Загрузить документ кандидата
// @Tags Кандидат
// @Description Загрузить документ кандидата
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    	string  true    "ID кандидата"
// @Param   document			formData	file 	true 	"file to upload"
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

	file, err := ctx.FormFile("document")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	logger := c.GetLogger(ctx)
	buffer, err := file.Open()
	if err != nil {
		return c.SendError(ctx, logger, err, "Ошибка при получении файла документа")
	}
	defer buffer.Close()
	fileBody, err := io.ReadAll(buffer)
	if err != nil {
		return c.SendError(ctx, logger, err, "Ошибка при загрузке файла документа")
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = filestorage.Instance.Upload(ctx.UserContext(), spaceID, applicantID, fileBody, file.Filename, dbmodels.ApplicantDoc)
	if err != nil {
		return c.SendError(ctx, logger, err, "Ошибка сохранения файла документа")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Загрузить фото кандидата
// @Tags Кандидат
// @Description Загрузить фото кандидата
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    	string  true    "ID кандидата"
// @Param   photo				formData	file 	true 	"Фото кандидата"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/upload-photo [post]
func (c *applicantApiController) uploadPhoto(ctx *fiber.Ctx) error {
	applicantID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	file, err := ctx.FormFile("photo")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	logger := c.GetLogger(ctx)
	buffer, err := file.Open()
	if err != nil {
		return c.SendError(ctx, logger, err, "Ошибка при получении файла с фото кандидата")
	}
	defer buffer.Close()
	fileBody, err := io.ReadAll(buffer)
	if err != nil {
		return c.SendError(ctx, logger, err, "Ошибка при загрузке файла с фото кандидата")
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = filestorage.Instance.Upload(ctx.UserContext(), spaceID, applicantID, fileBody, file.Filename, dbmodels.ApplicantPhoto)
	if err != nil {
		return c.SendError(ctx, logger, err, "Ошибка сохранения файла с фото кандидата")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Удалить фото кандидата
// @Tags Кандидат
// @Description Удалить фото кандидата
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    	string  true    "ID кандидата"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/photo [delete]
func (c *applicantApiController) deletePhoto(ctx *fiber.Ctx) error {
	applicantID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = filestorage.Instance.DeleteFileByType(ctx.UserContext(), spaceID, applicantID, dbmodels.ApplicantPhoto)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления файла с фото кандидата")
	}

	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Удалить резюме кандидата
// @Tags Кандидат
// @Description Удалить резюме кандидата
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    	string  true    "ID кандидата"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/resume [delete]
func (c *applicantApiController) deleteResume(ctx *fiber.Ctx) error {
	applicantID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = filestorage.Instance.DeleteFileByType(ctx.UserContext(), spaceID, applicantID, dbmodels.ApplicantResume)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления файла с резюме кандидата")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка выгрузки файла с документом кандидата")
	}

	return ctx.Send(body)
}

// @Summary Удалить документ кандидата
// @Tags Кандидат
// @Description Удалить документ кандидата
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    	string  true    "ID документа"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/doc/{id} [delete]
func (c *applicantApiController) deleteDoc(ctx *fiber.Ctx) error {
	docID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = filestorage.Instance.DeleteFile(ctx.UserContext(), spaceID, docID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления файла с документом кандидата")
	}

	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
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
	body, err := filestorage.Instance.GetFileByType(ctx.UserContext(), spaceID, applicantID, dbmodels.ApplicantResume)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка выгрузки файла с резюме кандидата")
	}

	return ctx.Send(body)
}

// @Summary Скачать фото кандидата
// @Tags Кандидат
// @Description Скачать фото кандидата
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "ID кандидата"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/photo [get]
func (c *applicantApiController) getPhoto(ctx *fiber.Ctx) error {
	applicantID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	body, err := filestorage.Instance.GetFileByType(ctx.UserContext(), spaceID, applicantID, dbmodels.ApplicantPhoto)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка выгрузки файла с фото кандидата")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка документов кандидата")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка кандидатов")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewScrollerResponse(list, rowCount))
}

// @Summary Список c причинами отказов
// @Tags Кандидат
// @Description Список c причинами отказов
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=applicantapimodels.RejectReasons}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/reject_list [post]
func (c *applicantApiController) rejectList(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(applicantdict.GetRejectReasonList()))
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка создания кандидата")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения данных кандидата")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка обновления данных кандидата")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка добавления тега к кандидату")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления тега у кандидата")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка объединения кандидатов")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка пометки кандидатов как разных")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка перевода кандидата на другой этап подбора")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения лога действий по кандидату")
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
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка добавления заметки по кандидату")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Отклонить кандидата
// @Tags Кандидат
// @Description Отклонить кандидата
// @Param   Authorization	 header		string	true	"Authorization token"
// @Param   id          	 path    	string  true    "Идентификатор кандидата"
// @Param	body body	 applicantapimodels.RejectRequest	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/{id}/reject [put]
func (c *applicantApiController) reject(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	var payload applicantapimodels.RejectRequest
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	userID := middleware.GetUserID(ctx)
	spaceID := middleware.GetUserSpace(ctx)
	err = applicant.Instance.ApplicantReject(spaceID, id, userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка отклонения кандидата")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Отклонить кандидатов
// @Tags Кандидат
// @Description Отклонить кандидатов
// @Param   Authorization	 header		string	true	"Authorization token"
// @Param	body body	 applicantapimodels.MultiRejectRequest	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/multi-actions/reject [put]
func (c *applicantApiController) multiReject(ctx *fiber.Ctx) error {
	var payload applicantapimodels.MultiRejectRequest
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	userID := middleware.GetUserID(ctx)
	spaceID := middleware.GetUserSpace(ctx)
	err := applicant.Instance.ApplicantMultiReject(spaceID, userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка отклонения кандидатов")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Перевести на другой этап подбора
// @Tags Кандидат
// @Description Перевести на другой этап подбора
// @Param   Authorization		header	string	true	"Authorization token"
// @Param	body body	 applicantapimodels.MultiChangeStageRequest	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/multi-actions/change_stage [put]
func (c *applicantApiController) multiChangeStage(ctx *fiber.Ctx) error {
	var payload applicantapimodels.MultiChangeStageRequest
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	err := applicant.Instance.MultiChangeStage(spaceID, userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка перевода кандидатов на другой этап подбора")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Отправить письма кандидатам
// @Tags Кандидат
// @Description Отправить письма кандидатам
// @Param   Authorization		header	string	true	"Authorization token"
// @Param	body body	 applicantapimodels.MultiChangeStageRequest	true	"request body"
// @Success 200 {object} apimodels.Response{data=applicantapimodels.MultiEmailResponse}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/multi-actions/send_email [put]
func (c *applicantApiController) multiSendMail(ctx *fiber.Ctx) error {
	var payload applicantapimodels.MultiEmailRequest
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	failMails, hMsg, err := messagetemplate.Instance.MultiSendEmail(spaceID, userID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка отправки писем кандидатам")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(failMails))
}

// @Summary Выгрузить в Excel
// @Tags Кандидат
// @Description Выгрузить в Excel
// @Param   Authorization		header	string	true	"Authorization token"
// @Param	body body	 applicantapimodels.XlsExportRequest	true	"request body"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/applicant/multi-actions/export_xls [put]
func (c *applicantApiController) multiExportXls(ctx *fiber.Ctx) error {
	var payload applicantapimodels.XlsExportRequest
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	data, err := applicant.Instance.ExportToXls(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка выгрузки списка кандидатов в Excel")
	}
	fileName := fmt.Sprintf("applicants-%v.xlsx", time.Now().Format("20060102-150405"))
	ctx.Set("Content-Type", "application/vnd.ms-excel")
	ctx.Set(fiber.HeaderContentDisposition, `attachment; filename="`+fileName+`"`)
	return ctx.SendStream(data)
}
