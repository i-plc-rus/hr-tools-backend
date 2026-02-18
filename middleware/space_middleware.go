package middleware

import (
	"hr-tools-backend/db"
	licensestore "hr-tools-backend/lib/licence/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	"hr-tools-backend/models"
	apimodels "hr-tools-backend/models/api"
	"net/http"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

var initOnce sync.Once
var noLicenseRequiredMap map[string]string

func init() {
	initOnce.Do(func() {
		noLicenseRequiredMap = map[string]string{
			"list": http.MethodPost,
			"find": http.MethodPost,
		}
	})
}

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

func LicenseRequired() fiber.Handler {

	store := licensestore.NewInstance(db.DB)
	return func(ctx *fiber.Ctx) (err error) {
		if ctx.Method() != http.MethodGet || ctx.Method() != http.MethodOptions {
			parts := strings.Split(ctx.Path(), "/")
			if noLicenseRequiredMap[parts[len(parts)-1]] == ctx.Method() {
				return ctx.Next()
			}
			spaceID := GetUserSpace(ctx)
			license, err := store.GetBySpace(spaceID)
			if err != nil {
				log.
					WithError(err).
					WithField("space_id", spaceID).
					Error("Ошибка проверки лицензии")
				return ctx.Status(fiber.StatusForbidden).JSON(apimodels.NewError("Ошибка проверки лицензии"))
			}
			if license == nil || license.Status.IdReadOnly() {
				return ctx.Status(fiber.StatusForbidden).JSON(apimodels.NewError("Лицензия истекла - продлите"))
			}
		}
		return ctx.Next()
	}
}
