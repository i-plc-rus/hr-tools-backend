package authutils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"hr-tools-backend/config"
	"hr-tools-backend/models"
	"time"
)

func GetToken(userID, name, spaceID string, isAdmin bool, role models.UserRole) (tokenString string, err error) {
	claims := jwt.MapClaims{
		"name":  name,
		"sub":   userID,
		"space": spaceID,
		"admin": isAdmin,
		"role":  string(role),
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
	token, ok := ctx.Locals("user").(*jwt.Token)
	if !ok {
		return jwt.MapClaims{}
	}
	return token.Claims.(jwt.MapClaims)
}
