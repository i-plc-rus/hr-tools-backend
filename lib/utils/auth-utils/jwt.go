package authutils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"hr-tools-backend/config"
	"time"
)

func GetToken(userID, name, spaceID string, isAdmin bool, role string) (tokenString string, err error) {
	claims := jwt.MapClaims{
		"name":  name,
		"sub":   userID,
		"space": spaceID,
		"admin": isAdmin,
		"role":  role,
		"exp":   time.Now().Add(time.Second * time.Duration(config.Conf.Auth.JWTExpireInSec)).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Conf.Auth.JWTSecret))
}

func GetRefreshToken(userID, name string) (tokenString string, err error) {
	claims := jwt.MapClaims{
		"name": name,
		"sub":  userID,
		"exp":  time.Now().Add(time.Second * time.Duration(config.Conf.Auth.JWTRefreshExpireInSec)).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Conf.Auth.JWTSecret))
}

func GetClaims(ctx *fiber.Ctx) jwt.MapClaims {
	token := ctx.Locals("user").(*jwt.Token)
	return token.Claims.(jwt.MapClaims)
}
