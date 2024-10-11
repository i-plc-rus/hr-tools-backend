package apiv1

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/controllers"
	handler "hr-tools-backend/lib/admin-panel"
	adminpanelauthhandler "hr-tools-backend/lib/admin-panel/auth"
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"
	adminpanelapimodels "hr-tools-backend/models/api/admin-panel"
	authapimodels "hr-tools-backend/models/api/auth"
)

type adminApiController struct {
	controllers.BaseAPIController
}

func InitAdminApiRouters(app *fiber.App) {
	controller := adminApiController{}
	app.Post("login", controller.login)

	// доступ всем авторизованным пользователям
	//otherApi := fiber.New()
	//app.Mount("/otherApi", otherApi)
	//otherApi.Use(middleware.AdminPanelAuthorizationRequired())
	//otherApi.Post("list", controller.userList)

	// доступ суперадминам
	user := fiber.New()
	app.Mount("/user", user)
	user.Use(middleware.AdminPanelAuthorizationRequired())
	user.Use(middleware.SuperAdminRole())
	user.Get("get/:userID", controller.userGet)
	user.Post("create", controller.userCreate)
	user.Put("update/:userID", controller.userUpdate)
	user.Delete("delete/:userID", controller.userDelete)
	user.Post("list", controller.userList)
}

// @Summary Аутентификация пользователя
// @Tags Админ панель
// @Description Аутентификация пользователя
// @Param	body				body		authapimodels.LoginRequest	true	"request body"
// @Success 200 {object} apimodels.Response{data=authapimodels.JWTResponse}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/admin_panel/login [post]
func (a *adminApiController) login(ctx *fiber.Ctx) error {
	var payload authapimodels.LoginRequest
	if err := a.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	resp, err := adminpanelauthhandler.Instance.Login(payload.Email, payload.Password)
	if err != nil {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(resp))
}

// @Summary Создание пользователя
// @Tags Админ панель. Пользователи
// @Description Создание пользователя
// @Param   Authorization		header		string	true	"Authorization token"
// @Param	body body	 adminpanelapimodels.User	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/admin_panel/user/create [post]
func (a *adminApiController) userCreate(ctx *fiber.Ctx) error {
	var payload adminpanelapimodels.User
	if err := a.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	if err := payload.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}
	err := handler.Instance.CreateUser(payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Изменение пользователя
// @Tags Админ панель. Пользователи
// @Description Изменение пользователя
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   userID          		path    string  				    	true         "user ID"
// @Param	body body	 adminpanelapimodels.UserUpdate	true	"request body"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/admin_panel/user/update/{userID} [put]
func (a *adminApiController) userUpdate(ctx *fiber.Ctx) error {
	value := ctx.Params("userID")
	if value == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("ID пользователя не указан"))
	}
	var payload adminpanelapimodels.UserUpdate
	if err := a.BodyParser(ctx, &payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError(err.Error()))
	}

	err := handler.Instance.UpdateUser(value, payload)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Удаление пользователя
// @Tags Админ панель. Пользователи
// @Description Удаление пользователя
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   userID          		path    string  				    	true         "user ID"
// @Success 200 {object} apimodels.Response
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/admin_panel/user/delete/{userID} [delete]
func (a *adminApiController) userDelete(ctx *fiber.Ctx) error {
	value := ctx.Params("userID")
	if value == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("ID пользователя не указан"))
	}
	err := handler.Instance.DeleteUser(value)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(nil))
}

// @Summary Получение пользователя
// @Tags Админ панель. Пользователи
// @Description Получение пользователя
// @Param   Authorization		header		string	true	"Authorization token"
// @Param   userID          		path    string  				    	true         "user ID"
// @Success 200 {object} apimodels.Response{data=adminpanelapimodels.UserView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/admin_panel/user/get/{userID} [get]
func (a *adminApiController) userGet(ctx *fiber.Ctx) error {
	value := ctx.Params("userID")
	if value == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(apimodels.NewError("ID пользователя не указан"))
	}

	user, err := handler.Instance.GetUser(value)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(user))
}

// @Summary Получение списка пользователей
// @Tags Админ панель. Пользователи
// @Description Получение списка пользователей
// @Param   Authorization		header		string	true	"Authorization token"
// @Success 200 {object} apimodels.Response{data=[]adminpanelapimodels.UserView}
// @Failure 400 {object} apimodels.Response
// @Failure 403
// @Failure 500 {object} apimodels.Response
// @router /api/v1/admin_panel/user/list [post]
func (a *adminApiController) userList(ctx *fiber.Ctx) error {
	users, err := handler.Instance.List()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(err.Error()))
	}
	return ctx.Status(fiber.StatusOK).JSON(apimodels.NewResponse(users))
}
