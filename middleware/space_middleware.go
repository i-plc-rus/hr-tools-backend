package middleware

import (
	"github.com/gofiber/fiber/v2"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	"hr-tools-backend/models"
	apimodels "hr-tools-backend/models/api"
)

func SuperAdminRoleRequired() fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		if GetSpaceRole(ctx) != models.UserRoleSuperAdmin {
			return ctx.Status(fiber.StatusForbidden).JSON(apimodels.NewError("операция недоступна"))
		}
		return ctx.Next()
	}
}

func GetUserSpace(ctx *fiber.Ctx) string {
	claims := authutils.GetClaims(ctx)
	return claims["space"].(string)
}
func GetUserID(ctx *fiber.Ctx) string {
	claims := authutils.GetClaims(ctx)
	return claims["sub"].(string)
}

func GetSpaceRole(ctx *fiber.Ctx) models.UserRole {
	claims := authutils.GetClaims(ctx)
	return claims["role"].(models.UserRole)
}

func SpaceAdminRequired() fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		if GetSpaceRole(ctx) != models.SpaceAdminRole {
			return ctx.Status(fiber.StatusForbidden).JSON(apimodels.NewError("операция недоступна"))
		}
		return ctx.Next()
	}
}
