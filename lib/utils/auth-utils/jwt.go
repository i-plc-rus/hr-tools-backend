package authutils

import (
	"github.com/golang-jwt/jwt/v5"
	"hr-tools-backend/config"
	"time"
)

func GetToken(subject, name string, isAdmin bool) (tokenString string, err error) {
	claims := jwt.MapClaims{
		"name":  name,
		"sub":   subject,
		"admin": isAdmin,
		"exp":   time.Now().Add(time.Second * time.Duration(config.Conf.Auth.JWTExpireInSec)).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Conf.Auth.JWTSecret))
}

func GetRefreshToken(subject, name string) (tokenString string, err error) {
	claims := jwt.MapClaims{
		"name": name,
		"sub":  subject,
		"exp":  time.Now().Add(time.Second * time.Duration(config.Conf.Auth.JWTRefreshExpireInSec)).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Conf.Auth.JWTSecret))
}
