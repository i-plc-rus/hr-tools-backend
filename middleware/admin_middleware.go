package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"hr-tools-backend/models"
	apimodels "hr-tools-backend/models/api"
)

func SuperAdminRole() fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		token := ctx.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		role := claims["role"].(string)
		if role != string(models.UserRoleSuperAdmin) {
			return ctx.Status(fiber.StatusForbidden).JSON(apimodels.NewError("операция недоступна"))
		}
		return ctx.Next()
	}
}

func GetUserSpace(ctx *fiber.Ctx) string {
	token := ctx.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	return claims["space"].(string)
}

func SpaceAdminUser() fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		token := ctx.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		if !claims["admin"].(bool) {
			return ctx.Status(fiber.StatusForbidden).JSON(apimodels.NewError("операция недоступна"))
		}
		return ctx.Next()
	}
}
