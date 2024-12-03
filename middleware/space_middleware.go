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
	if space, exist := claims["space"]; exist {
		return space.(string)
	}
	return ""
}
func GetUserID(ctx *fiber.Ctx) string {
	claims := authutils.GetClaims(ctx)
	if sub, exist := claims["sub"]; exist {
		return sub.(string)
	}
	return ""
}

func GetSpaceRole(ctx *fiber.Ctx) models.UserRole {
	claims := authutils.GetClaims(ctx)
	if role, exist := claims["role"]; exist {
		if stringRole, ok := role.(string); ok && stringRole != "" {
			return models.UserRole(stringRole)
		}
	}
	return ""
}

func SpaceAdminRequired() fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		if GetSpaceRole(ctx) != models.SpaceAdminRole {
			return ctx.Status(fiber.StatusForbidden).JSON(apimodels.NewError("операция недоступна"))
		}
		return ctx.Next()
	}
}
