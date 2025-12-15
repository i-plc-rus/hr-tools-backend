package adminpanelauthhandler

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	adminpaneluserstore "hr-tools-backend/lib/admin-panel/store"
	authhelpers "hr-tools-backend/lib/utils/auth-helpers"
	authapimodels "hr-tools-backend/models/api/auth"
	"time"
)

type Provider interface {
	Login(email, password string) (response authapimodels.JWTResponse, err error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store: adminpaneluserstore.NewInstance(db.DB),
	}
}

type impl struct {
	store adminpaneluserstore.Provider
}

func (i impl) Login(email, password string) (response authapimodels.JWTResponse, err error) {
	logger := log.WithField("email", email)
	user, err := i.store.FindByEmail(email)
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
	if authhelpers.GetMD5Hash(password) != user.Password {
		logger.Debug("пользователь не прошел проверку пароля")
		return authapimodels.JWTResponse{}, errors.New("пользователь не прошел проверку пароля")
	}
	claims := jwt.MapClaims{
		"name": fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Second * time.Duration(config.Conf.AdminPanelAuth.JWTExpireInSec)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Conf.AdminPanelAuth.JWTSecret))
	if err != nil {
		logger.WithError(err).Error("ошибка генерации JWT")
		return authapimodels.JWTResponse{}, err
	}
	err = i.store.Update(user.ID, map[string]interface{}{"LastLogin": time.Now()})
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка обновления даты последнего входа")
	}
	return authapimodels.JWTResponse{
		Token: tokenString,
	}, nil
}
