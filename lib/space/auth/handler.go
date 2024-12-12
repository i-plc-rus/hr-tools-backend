package spaceauthhandler

import (
	"fmt"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	emailverify "hr-tools-backend/lib/email-verify"
	"hr-tools-backend/lib/smtp"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	authapimodels "hr-tools-backend/models/api/auth"
	spaceapimodels "hr-tools-backend/models/api/space"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	SendEmailConfirmation(email string) error
	VerifyEmail(code string) error
	CheckEmail(email string) (bool, error)
	Login(email, password string) (response authapimodels.JWTResponse, err error)
	Me(ctx *fiber.Ctx) (spaceUser spaceapimodels.SpaceUser, err error)
	RefreshToken(ctx *fiber.Ctx, refreshToken string) (response authapimodels.JWTResponse, err error)
	PasswordRecovery(email string) error
	PasswordReset(resetCode, newPassword string) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		emailVerify:     emailverify.NewInstance(config.Conf.Smtp.EmailSendVerification),
		spaceUsersStore: spaceusersstore.NewInstance(db.DB),
		systemEmail:     config.Conf.Smtp.EmailSendVerification,
	}
}

type impl struct {
	emailVerify     emailverify.Provider
	spaceUsersStore spaceusersstore.Provider
	systemEmail     string
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
		tokenString, err := authutils.GetToken(userID, user.GetFullName(), user.SpaceID, user.Role.IsSpaceAdmin(), string(user.Role))
		if err != nil {
			log.WithError(err).Error("ошибка генерации JWT")
			return authapimodels.JWTResponse{}, err
		}
		refreshTokenString, err := authutils.GetRefreshToken(userID, user.GetFullName())
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
	user, err := i.spaceUsersStore.FindByEmail(email, false)
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
	if smtp.Instance.IsConfigured() && !user.Role.IsSpaceAdmin() && !user.IsEmailVerified {
		return authapimodels.JWTResponse{}, errors.New("необходимо подтвердить почту")
	}
	tokenString, err := authutils.GetToken(user.ID, user.GetFullName(), user.SpaceID, user.Role.IsSpaceAdmin(), string(user.Role))
	if err != nil {
		logger.WithError(err).Error("ошибка генерации JWT")
		return authapimodels.JWTResponse{}, err
	}
	refresTokenString, err := authutils.GetRefreshToken(user.ID, user.GetFullName())
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

func (i impl) PasswordRecovery(email string) error {
	logger := log.WithField("email", email)
	user, err := i.spaceUsersStore.FindByEmail(email, false)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка поиска пользователя по почте")
		return errors.New("ошибка поиска пользователя по почте")
	}
	if user == nil {
		logger.Debug("пользователь с такой почтой не найден")
		return errors.New("пользователь с такой почтой не найден")
	}

	if user.ResetTime.Add(time.Minute * 5).After(time.Now()) {
		// уже отправили
		return nil
	}
	if !smtp.Instance.IsConfigured() {
		logger.Error("восстановление пароля невозможно, почтовый клиент не настроен")
		return errors.New("восстановление пароля невозможно, обратитесь к администратору")
	}
	if !user.IsEmailVerified {
		logger.Error("восстановление пароля невозможно, емайл пользователя не подтвержден")
		return errors.New("восстановление пароля невозможно, обратитесь к администратору")
	}

	resetCode := uuid.New().String()
	updMap := map[string]interface{}{
		"reset_code": resetCode,
		"reset_time": time.Now(),
	}
	err = i.spaceUsersStore.Update(user.ID, updMap)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка сохранения кода для восстановления")
		return errors.New("произошла ошибка, попробуйте запросить восстановление пароля чуть позже")
	}
	message := "Здравствуйте,\r\n" +
		"Вы запросили сброс пароля вашей учетной записи.\r\n" +
		"Пожалуйста, нажмите кнопку ниже, чтобы создать новый пароль. Если вы не хотели сбрасывать свой пароль, просто проигнорируйте это письмо.\r\n" +
		fmt.Sprintf("[ %s?reset_code=%s ]\r\n", config.Conf.Smtp.ResetUI, resetCode) +
		"Обратите внимание, что эту ссылку можно использовать только один раз. Если вы отправили более 1 запроса на сброс пароля, используйте последнюю полученную вами ссылку."
	err = smtp.Instance.SendEMail(i.systemEmail, email, message, "Восстановление пароля")
	if err != nil {
		return err
	}

	return nil
}

func (i impl) PasswordReset(resetCode, newPassword string) error {
	logger := log.WithField("reset_code", resetCode)
	user, err := i.spaceUsersStore.GetByResetCode(resetCode)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка поиска пользователя по коду для сброса пароля")
		return errors.New("ссылка не найдена или более не актуальна, попробуйте выполнить восстановление пароля еще раз")
	}
	if resetCode == "" || user == nil || user.ResetTime.Add(time.Minute*15).Before(time.Now()) {
		return errors.New("ссылка не найдена или более не актуальна, попробуйте выполнить восстановление пароля еще раз")
	}
	updMap := map[string]interface{}{
		"reset_code": "",
		"reset_time": time.Now(),
		"password":   authutils.GetMD5Hash(newPassword),
	}
	err = i.spaceUsersStore.Update(user.ID, updMap)
	if err != nil {
		logger.
			WithError(err).
			Error("не удалось обновить пароль, ошибка сохранения нового пароля")
		return errors.New("не удалось обновить пароль, попробуйте выполнить восстановление пароля немного позже")
	}
	return nil
}
