package middleware

import (
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/lib/rbac"
)

func RbacMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Получаем данные пользователя из контекста
		userID := GetUserID(ctx)
		if userID == "" {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "RBAC_FORBIDDEN",
			})
		}
		spaceID := GetUserSpace(ctx)

		userRole := GetSpaceRole(ctx)
		if userRole == "" {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "RBAC_FORBIDDEN",
			})
		}

		// Ищем обработчик
		handler, found := rbac.Instance.GetRuleFunc(ctx.Method(), ctx.Path())
		if !found {
			return ctx.Next()
		}

		// Выполняем проверку
		if !handler(spaceID, userID, userRole, ctx.Path()) {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "RBAC_FORBIDDEN",
			})
		}

		return ctx.Next()
	}
}
