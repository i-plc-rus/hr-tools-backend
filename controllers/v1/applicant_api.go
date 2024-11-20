package apiv1

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/controllers"
	filestorage "hr-tools-backend/lib/file-storage"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	"io"
)

type applicantApiController struct {
	controllers.BaseAPIController
}

func InitApplicantApiRouters(app *fiber.App) {
	controller := applicantApiController{}
	app.Route("applicant", func(router fiber.Router) {
		router.Get("doc/:id", controller.GetDoc) // скачать документ по id
		router.Route(":id", func(idRouter fiber.Router) {
			idRouter.Post("upload-resume", controller.UploadResume) // загрузить резюме кандидата
			idRouter.Post("upload-doc", controller.UploadDoc)       // загрузить документ кандидата
			idRouter.Get("doc/list", controller.GetDocList)         // получить список документов кандидата
			idRouter.Get("resume", controller.GetResume)            // скачать резюме кандидата
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
