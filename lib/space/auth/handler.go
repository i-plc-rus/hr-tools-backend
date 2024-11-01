package spaceauthhandler

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	emailverify "hr-tools-backend/lib/email-verify"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	authapimodels "hr-tools-backend/models/api/auth"
	spaceapimodels "hr-tools-backend/models/api/space"
)

type Provider interface {
	SendEmailConfirmation(email string) error
	VerifyEmail(code string) error
	CheckEmail(email string) (bool, error)
	Login(email, password string) (response authapimodels.JWTResponse, err error)
	Me(ctx *fiber.Ctx) (spaceUser spaceapimodels.SpaceUser, err error)
	RefreshToken(ctx *fiber.Ctx, refreshToken string) (response authapimodels.JWTResponse, err error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		emailVerify:     emailverify.NewInstance(config.Conf.Smtp.EmailSendVerification),
		spaceUsersStore: spaceusersstore.NewInstance(db.DB),
	}
}

type impl struct {
	emailVerify     emailverify.Provider
	spaceUsersStore spaceusersstore.Provider
}

func (i impl) RefreshToken(ctx *fiber.Ctx, refreshToken string) (response authapimodels.JWTResponse, err error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Conf.Auth.JWTSecret), nil
	})
	if err != nil {
		return authapimodels.JWTResponse{}, err
	}

	if claimsReq, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claimsReq["sub"].(string)
		user, err := i.spaceUsersStore.GetByID(userID)
		if err != nil {
			log.
				WithField("user_id", userID).
				WithError(err).
				Error("ошибка поиска пользователя")
			return authapimodels.JWTResponse{}, err
		}
		tokenString, err := authutils.GetToken(userID, fmt.Sprintf("%s %s", user.FirstName, user.LastName), user.SpaceID, user.IsAdmin)
		if err != nil {
			log.WithError(err).Error("ошибка генерации JWT")
			return authapimodels.JWTResponse{}, err
		}
		refreshTokenString, err := authutils.GetRefreshToken(userID, fmt.Sprintf("%s %s", user.FirstName, user.LastName))
		if err != nil {
			log.WithError(err).Error("ошибка генерации refresh JWT")
			return authapimodels.JWTResponse{}, err
		}
		return authapimodels.JWTResponse{
			Token:        tokenString,
			RefreshToken: refreshTokenString,
		}, nil

	}
	return authapimodels.JWTResponse{}, errors.New("refresh token is not valid")
}

func (i impl) Me(ctx *fiber.Ctx) (spaceUser spaceapimodels.SpaceUser, err error) {
	token := ctx.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)
	logger := log.WithField("user_id", userID)
	user, err := i.spaceUsersStore.GetByID(userID)
	if err != nil {
		logger.WithError(err).Error("ошибка поиска пользователя")
		return spaceapimodels.SpaceUser{}, err
	}
	return user.ToModel(), nil

}

func (i impl) Login(email, password string) (response authapimodels.JWTResponse, err error) {
	logger := log.WithField("email", email)
	user, err := i.spaceUsersStore.FindByEmail(email)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка поиска пользователя по почте")
		return authapimodels.JWTResponse{}, err
	}
	if user == nil {
		logger.Debug("пользователь с такой почтой не найден")
		return authapimodels.JWTResponse{}, errors.New("пользователь с такой почтой не найден")
	}
	if authutils.GetMD5Hash(password) != user.Password {
		logger.Debug("пользователь не прошел проверку пароля")
		return authapimodels.JWTResponse{}, errors.New("пользователь не прошел проверку пароля")
	}
	tokenString, err := authutils.GetToken(user.ID, fmt.Sprintf("%s %s", user.FirstName, user.LastName), user.SpaceID, user.IsAdmin)
	if err != nil {
		logger.WithError(err).Error("ошибка генерации JWT")
		return authapimodels.JWTResponse{}, err
	}
	refresTokenString, err := authutils.GetRefreshToken(user.ID, fmt.Sprintf("%s %s", user.FirstName, user.LastName))
	if err != nil {
		logger.WithError(err).Error("ошибка генерации refresh JWT")
		return authapimodels.JWTResponse{}, err
	}
	return authapimodels.JWTResponse{
		Token:        tokenString,
		RefreshToken: refresTokenString,
	}, nil
}

func (i impl) CheckEmail(email string) (passed bool, err error) {
	exist, err := i.spaceUsersStore.ExistByEmail(email)
	if err != nil {
		return false, err
	}
	return !exist, nil
}

func (i impl) VerifyEmail(code string) error {
	err := i.emailVerify.VerifyCode(code)
	if err != nil {
		return err
	}
	return nil
}

func (i impl) SendEmailConfirmation(email string) error {
	err := i.emailVerify.SendVerifyCode(email)
	if err != nil {
		return err
	}
	return nil
}
