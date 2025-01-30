package oauth2

import (
	"context"
	"encoding/json"
	"hr-tools-backend/db"
	pgadapter "hr-tools-backend/lib/oauth2/pg-adapter"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	"hr-tools-backend/models"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-session/session/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	log "github.com/sirupsen/logrus"
	pg "github.com/vgarvardt/go-oauth2-pg/v4"
)

type oauth2Controller struct {
	srv       *server.Server
	userStore spaceusersstore.Provider
}

func InitOauthRouters(app *fiber.App) {
	controller := oauth2Controller{
		userStore: spaceusersstore.NewInstance(db.DB),
	}
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.SetClientTokenCfg(manage.DefaultClientTokenCfg)

	// generate jwt access token
	// manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte("00000000"), jwt.SigningMethodHS512))
	manager.MapAccessGenerate(generates.NewAccessGenerate())

	// token store
	tokenStore, _ := pg.NewTokenStore(pgadapter.NewInstance(db.DB), pg.WithTokenStoreGCInterval(time.Minute))
	manager.MapTokenStorage(tokenStore)

	// client store
	clientStore, _ := pg.NewClientStore(pgadapter.NewInstance(db.DB))
	manager.MapClientStorage(clientStore)

	controller.srv = server.NewServer(server.NewConfig(), manager)

	controller.srv.SetPasswordAuthorizationHandler(controller.passwordAuthorizationHandler)

	controller.srv.SetUserAuthorizationHandler(userAuthorizeHandler)

	controller.srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.WithError(err).Error("Oauth2 Internal Error")
		return
	})

	controller.srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.WithError(re.Error).Error("Oauth2 Response Error")
	})
	controller.srv.SetAllowGetAccessRequest(true)
	controller.srv.SetClientInfoHandler(clientInfoHandler) // авторизация и через форму или basic authorization

	app.All("/login", adaptor.HTTPHandlerFunc(controller.loginHandler))
	app.All("/auth", adaptor.HTTPHandlerFunc(controller.authHandler))
	app.All("/authorize", adaptor.HTTPHandlerFunc(controller.hAuthorize))
	app.All("/oauth/token", adaptor.HTTPHandlerFunc(controller.hTokenRequest))
	app.All("/test", adaptor.HTTPHandlerFunc(controller.hTest))
	app.All("/userinfo", adaptor.HTTPHandlerFunc(controller.userInfo))
}

// @Summary Эндпоинт формы логина
// @Tags OAuth2
// @Description Эндпоинт формы логина
// @Success 200
// @Failure 400
// @router /oauth2/login [get]
// @router /oauth2/login [post]
func (c *oauth2Controller) loginHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		if r.Form == nil {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		username := r.Form.Get("username")
		password := r.Form.Get("password")
		clientID, userID := c.checkUser(username, password)
		if clientID != "" {
			store.Set("LoggedInUserID", userID)
			store.Save()

			w.Header().Set("Location", "/oauth2/auth")
			w.WriteHeader(http.StatusFound)
			return
		} else {
			http.Error(w, "Некорректный адрес электронной почты или пароль", http.StatusUnauthorized)
			return
		}
	}
	outputHTML(w, r, "./static/login.html")
}

// @Summary Эндпоинт формы подтверждения доступа
// @Tags OAuth2
// @Description Эндпоинт формы подтверждения доступа
// @Success 200
// @Failure 400
// @router /oauth2/auth [get]
// @router /oauth2/auth [post]
func (c *oauth2Controller) authHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(nil, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := store.Get("LoggedInUserID"); !ok {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}
	outputHTML(w, r, "./static/auth.html")
}

// @Summary Авторизация
// @Tags OAuth2
// @Description Авторизация
// @Success 200
// @Failure 400
// @router /oauth2/authorize [get]
// @router /oauth2/authorize [post]
func (c *oauth2Controller) hAuthorize(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var form url.Values
	if v, ok := store.Get("ReturnUri"); ok {
		form = v.(url.Values)
	}
	r.Form = form

	store.Delete("ReturnUri")
	store.Save()

	err = c.srv.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// @Summary Запрос токена
// @Tags OAuth2
// @Description Запрос токена
// @Success 200
// @Failure 400
// @router /oauth2/oauth/token [get]
// @router /oauth2/oauth/token [post]
func (c *oauth2Controller) hTokenRequest(w http.ResponseWriter, r *http.Request) {
	err := c.srv.HandleTokenRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Проверка токена
// @Tags OAuth2
// @Description Проверка токена
// @Param   Authorization		header	string	true	"Authorization oauth2 token"  example(Bearer MWI3Y2RHNZYTNMNHYS0ZNGRHLTHINJGTZWM2OTNLOTRKNJEW)
// @Success 200
// @Failure 400
// @router /oauth2/test [get]
// @router /oauth2/test [post]
func (c *oauth2Controller) hTest(w http.ResponseWriter, r *http.Request) {
	token, err := c.srv.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := map[string]interface{}{
		"expires_in": int64(token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Sub(time.Now()).Seconds()),
		"client_id":  token.GetClientID(),
		"user_id":    token.GetUserID(),
	}
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	e.Encode(data)
}

// @Summary Информация о пользователе
// @Tags OAuth2
// @Description Информация о пользователе
// @Param   Authorization		header	string	true	"Authorization oauth2 token"  example(Bearer MWI3Y2RHNZYTNMNHYS0ZNGRHLTHINJGTZWM2OTNLOTRKNJEW)
// @Success 200
// @Failure 400
// @router /oauth2/userinfo [get]
// @router /oauth2/userinfo [post]
func (c *oauth2Controller) userInfo(w http.ResponseWriter, r *http.Request) {
	token, err := c.srv.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := token.GetUserID()
	user, err := c.userStore.GetByID(userID)
	if err != nil || user == nil {
		http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
		return
	}

	w.Header().Add(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	data := map[string]interface{}{
		"username":   user.Email,
		"name":       user.GetFullName(),
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"roles":      []models.UserRole{user.Role},
	}
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	e.Encode(data)
}

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		return "", err
	}

	uid, ok := store.Get("LoggedInUserID")
	if !ok {
		if r.Form == nil {
			r.ParseForm()
		}
		store.Set("ReturnUri", r.Form)
		store.Save()

		w.Header().Set("Location", "/oauth2/login")
		w.WriteHeader(http.StatusFound)
		return "", nil
	}

	userID = uid.(string)
	store.Delete("LoggedInUserID")
	store.Save()
	return userID, nil
}

func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}

func clientInfoHandler(r *http.Request) (string, string, error) {
	auth := r.Header.Get("Authorization")
	if auth != "" {
		return server.ClientBasicHandler(r)
	}
	return server.ClientFormHandler(r)
}

func (c *oauth2Controller) passwordAuthorizationHandler(ctx context.Context, clientID, username, password string) (userID string, err error) {
	if clientID == "" {
		return "", errors.ErrInvalidClient
	}
	storedClientID, userID := c.checkUser(username, password)
	if storedClientID != clientID {
		return "", errors.ErrInvalidClient
	}
	return userID, nil
}

func (c *oauth2Controller) checkUser(username, password string) (clientID, userID string) {
	user, err := c.userStore.FindByEmail(username, false)
	if err != nil || user == nil {
		return "", ""
	}
	if authutils.GetMD5Hash(password) != user.Password {
		return "", ""
	}
	return user.SpaceID, user.ID
}
