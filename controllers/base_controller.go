package controllers

import (
	"hr-tools-backend/middleware"
	apimodels "hr-tools-backend/models/api"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type BaseAPIController struct{}

func (c *BaseAPIController) BodyParser(ctx *fiber.Ctx, out interface{}) error {
	if err := ctx.BodyParser(out); err != nil {
		c.GetLogger(ctx).WithError(err).Error("Некорректный запрос")
		return errors.New("Некорректный запрос")
	}
	return nil
}

func (c *BaseAPIController) GetIDByKey(ctx *fiber.Ctx, key string) (id string, err error) {

	value := ctx.Params(key)
	if value == "" {
		return "", errors.New("не указан id")
	}
	return value, nil
}

func (c *BaseAPIController) GetID(ctx *fiber.Ctx) (id string, err error) {
	return c.GetIDByKey(ctx, "id")
}

func (c *BaseAPIController) GetLogger(ctx *fiber.Ctx) *log.Entry {
	spaceID := middleware.GetUserSpace(ctx)
	userID := middleware.GetUserID(ctx)
	role := middleware.GetSpaceRole(ctx)
	logger := log.WithField("space_id", spaceID)
	if userID != "" {
		logger = logger.WithField("user_id", userID)
	}
	if role != "" {
		logger = logger.WithField("user_role", role)
	}
	return logger
}

func (c *BaseAPIController) SendError(ctx *fiber.Ctx, logger *log.Entry, err error, msg string) error {
	logger.WithError(err).Error(msg)
	return ctx.Status(fiber.StatusInternalServerError).JSON(apimodels.NewError(msg))
}
