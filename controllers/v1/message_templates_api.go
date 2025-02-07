package apiv1

import (
	"hr-tools-backend/controllers"
	pdfexport "hr-tools-backend/lib/export/pdf"
	filestorage "hr-tools-backend/lib/file-storage"
	messagetemplate "hr-tools-backend/lib/message-template"
	"hr-tools-backend/lib/utils/helpers"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	msgtemplateapimodels "hr-tools-backend/models/api/message-template"
	dbmodels "hr-tools-backend/models/db"
	"io"
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

type msgTemplateApiController struct {
	controllers.BaseAPIController
}

func InitMsgTemplateApiRouters(app *fiber.App) {
	controller := msgTemplateApiController{}
	app.Route("msg-templates", func(router fiber.Router) {
		router.Post("send-email-msg", controller.SendEmailMessage) // отправить сообщение на почту кандидату
		router.Get("list", controller.GetTemplatesList)            // получить список шаблонов сообщений
		router.Post("", controller.create)
		router.Get("variables", controller.variables)
		router.Route(":id", func(idRoute fiber.Router) {
			idRoute.Put("", controller.update)
			idRoute.Get("", controller.get)
			idRoute.Delete("", controller.delete)
			idRoute.Post("logo", controller.uploadLogo)
			idRoute.Get("logo", controller.getLogo)
			idRoute.Delete("logo", controller.deleteLogo)
			idRoute.Post("sign", controller.uploadSign)
			idRoute.Get("sign", controller.getSign)
			idRoute.Delete("sign", controller.deleteSign)
			idRoute.Post("stamp", controller.uploadStamp)
			idRoute.Get("stamp", controller.getStamp)
			idRoute.Delete("stamp", controller.deleteStamp)
			idRoute.Get("preview_pdf", controller.previewPdf)
		})
	})
}

// @Summary Отправить сообщение кандидату на почту
// @Tags Шаблоны сообщений
// @Description Отправить сообщение кандидату на почту
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body				body		msgtemplateapimodels.SendMessage	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/send-email-msg [post]
func (c *msgTemplateApiController) SendEmailMessage(ctx *fiber.Ctx) error {
	var payload msgtemplateapimodels.SendMessage
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	hMsg, err := messagetemplate.Instance.SendEmailMessage(ctx.Context(), spaceID, payload.MsgTemplateID, payload.ApplicantID, userID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка отправки сообщения кандидату")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Список шаблонов сообщений
// @Tags Шаблоны сообщений
// @Description Список шаблонов сообщений
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=[]msgtemplateapimodels.MsgTemplateView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/list [get]
func (c *msgTemplateApiController) GetTemplatesList(ctx *fiber.Ctx) error {
	spaceID := middleware.GetUserSpace(ctx)
	list, err := messagetemplate.Instance.GetListTemplates(spaceID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения списка шаблонов")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(list))
}

// @Summary Создание
// @Tags Шаблоны сообщений
// @Description Создание
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 msgtemplateapimodels.MsgTemplateData	true	"request body"
// @Success 200 {object} apimodels.Response{data=string}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router  /api/v1/space/msg-templates [post]
func (c *msgTemplateApiController) create(ctx *fiber.Ctx) error {
	var payload msgtemplateapimodels.MsgTemplateData
	if err := c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	id, err := messagetemplate.Instance.Create(spaceID, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка добавления шаблона сообщений")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(id))
}

// @Summary Переменные шаблона
// @Tags Шаблоны сообщений
// @Description Переменные шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=[]msgtemplateapimodels.TemplateItem}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/variables [get]
func (c *msgTemplateApiController) variables(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(messagetemplate.GetVariables()))
}

// @Summary Обновление
// @Tags Шаблоны сообщений
// @Description Обновление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 msgtemplateapimodels.MsgTemplateData	true	"request body"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id} [put]
func (c *msgTemplateApiController) update(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	var payload msgtemplateapimodels.MsgTemplateData
	if err = c.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err = payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = messagetemplate.Instance.Update(spaceID, id, payload)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка изменения шаблона сообщений")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение по ИД
// @Tags Шаблоны сообщений
// @Description Получение по ИД
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response{data=dictapimodels.CompanyView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id} [get]
func (c *msgTemplateApiController) get(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	resp, err := messagetemplate.Instance.GetByID(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения шаблона сообщений")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Удаление
// @Tags Шаблоны сообщений
// @Description Удаление
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   id          		path    string  				    	true         "rec ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id} [delete]
func (c *msgTemplateApiController) delete(ctx *fiber.Ctx) error {
	id, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	spaceID := middleware.GetUserSpace(ctx)
	err = messagetemplate.Instance.Delete(spaceID, id)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления причины шаблона сообщений")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Загрузить изображение логотипа компании для шаблона
// @Tags Шаблоны сообщений
// @Description Загрузить изображение логотипа компании для шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   photo				formData	file 	true 	"Фото"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id}/logo [post]
func (c *msgTemplateApiController) uploadLogo(ctx *fiber.Ctx) error {
	tplID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	file, err := ctx.FormFile("photo")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	return c.uploadImage(ctx, tplID, file, dbmodels.CompanyLogo)
}

// @Summary Загрузить изображение подписи для шаблона
// @Tags Шаблоны сообщений
// @Description Загрузить изображение подписи для шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   photo				formData	file 	true 	"Фото"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id}/sign [post]
func (c *msgTemplateApiController) uploadSign(ctx *fiber.Ctx) error {
	tplID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	file, err := ctx.FormFile("photo")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	return c.uploadImage(ctx, tplID, file, dbmodels.CompanySign)
}

// @Summary Загрузить изображение печати компании для шаблона
// @Tags Шаблоны сообщений
// @Description Загрузить изображение печати компании для шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   photo				formData	file 	true 	"Фото"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id}/stamp [post]
func (c *msgTemplateApiController) uploadStamp(ctx *fiber.Ctx) error {
	tplID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	file, err := ctx.FormFile("photo")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	return c.uploadImage(ctx, tplID, file, dbmodels.CompanyStamp)
}

// @Summary Скачать изображение логотипа компании для шаблона
// @Tags Шаблоны сообщений
// @Description Скачать изображение логотипа компании для шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id}/logo [get]
func (c *msgTemplateApiController) getLogo(ctx *fiber.Ctx) error {
	tplID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	return c.getImage(ctx, tplID, dbmodels.CompanyLogo)
}

// @Summary Скачать изображение подписи для шаблона
// @Tags Шаблоны сообщений
// @Description Скачать изображение подписи для шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id}/sign [get]
func (c *msgTemplateApiController) getSign(ctx *fiber.Ctx) error {
	tplID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	return c.getImage(ctx, tplID, dbmodels.CompanySign)
}

// @Summary Скачать изображение печати компании для шаблона
// @Tags Шаблоны сообщений
// @Description Скачать изображение печати компании для шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id}/stamp [get]
func (c *msgTemplateApiController) getStamp(ctx *fiber.Ctx) error {
	tplID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	return c.getImage(ctx, tplID, dbmodels.CompanyStamp)
}

// @Summary Удалить изображение логотипа компании для шаблона
// @Tags Шаблоны сообщений
// @Description Удалить изображение логотипа компании для шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id}/logo [delete]
func (c *msgTemplateApiController) deleteLogo(ctx *fiber.Ctx) error {
	tplID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	return c.deletePhoto(ctx, tplID, dbmodels.CompanyLogo)
}

// @Summary Удалить изображение подписи для шаблона
// @Tags Шаблоны сообщений
// @Description Удалить изображение подписи для шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id}/sign [delete]
func (c *msgTemplateApiController) deleteSign(ctx *fiber.Ctx) error {
	tplID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	return c.deletePhoto(ctx, tplID, dbmodels.CompanySign)
}

// @Summary Удалить изображение печати компании для шаблона
// @Tags Шаблоны сообщений
// @Description Удалить изображение печати компании для шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id}/stamp [delete]
func (c *msgTemplateApiController) deleteStamp(ctx *fiber.Ctx) error {
	tplID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	return c.deletePhoto(ctx, tplID, dbmodels.CompanyStamp)
}

// @Summary Предпросмотр pdf на основе шаблона
// @Tags Шаблоны сообщений
// @Description Предпросмотр pdf на основе шаблона
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/space/msg-templates/{id}/preview_pdf [get]
func (c *msgTemplateApiController) previewPdf(ctx *fiber.Ctx) error {
	tplID, err := c.GetID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	body, hMsg, err := messagetemplate.Instance.PdfPreview(ctx.Context(), spaceID, tplID, userID)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка генерации pdf на основе шаблона")
	}
	if hMsg != "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(hMsg))
	}
	ctx.Set(fiber.HeaderContentType, "application/pdf")
	return ctx.Send(body)
}

func (c *msgTemplateApiController) uploadImage(ctx *fiber.Ctx, tplID string, file *multipart.FileHeader, fileType dbmodels.FileType) error {
	err := checkImgExt(file.Filename)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Неподдерживаемый тип файла")
	}
	buffer, err := file.Open()
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка при получении файла")
	}
	defer buffer.Close()
	fileBody, err := io.ReadAll(buffer)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка при загрузке файла")
	}

	spaceID := middleware.GetUserSpace(ctx)
	contentType := helpers.GetFileContentType(file)
	err = filestorage.Instance.Upload(ctx.UserContext(), spaceID, tplID, fileBody, file.Filename, fileType, contentType)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка сохранения изображения")
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

func (c *msgTemplateApiController) getImage(ctx *fiber.Ctx, tplID string, fileType dbmodels.FileType) error {
	spaceID := middleware.GetUserSpace(ctx)
	body, file, err := filestorage.Instance.GetFileByType(ctx.UserContext(), spaceID, tplID, fileType)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка получения изображения")
	}
	if file != nil && file.ContentType != "" {
		ctx.Set(fiber.HeaderContentType, file.ContentType)
		ctx.Set(fiber.HeaderContentDisposition, `inline; filename="`+file.Name+`"`)
	}
	return ctx.Send(body)
}

func (c *msgTemplateApiController) deletePhoto(ctx *fiber.Ctx, tplID string, fileType dbmodels.FileType) error {
	spaceID := middleware.GetUserSpace(ctx)
	err := filestorage.Instance.DeleteFileByType(ctx.UserContext(), spaceID, tplID, fileType)
	if err != nil {
		return c.SendError(ctx, c.GetLogger(ctx), err, "Ошибка удаления изображения")
	}

	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// форматы поддерживаемые fpdf
func checkImgExt(fileName string) error {
	imageType, err := pdfexport.GetImgType(fileName)
	if err != nil {
		return err
	}
	imageType = strings.ToLower(imageType)
	if imageType == "jpeg" {
		imageType = "jpg"
	}
	switch imageType {
	case "jpg":
	case "png":
	case "gif":
		return nil
	default:
		return errors.Errorf("неподдерживаемый тип изображения: %s", imageType)
	}
	return nil
}
