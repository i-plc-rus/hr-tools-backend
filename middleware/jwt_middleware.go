package middleware

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"hr-tools-backend/config"
)

func AuthorizationRequired() fiber.Handler {
	return jwtware.New(jwtware.Config{
		Claims: jwt.MapClaims{},
		SigningKey: jwtware.SigningKey{
			JWTAlg: "HS256",
			Key:    []byte(config.Conf.Auth.JWTSecret),
		},
	})
}
