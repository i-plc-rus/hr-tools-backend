package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type BaseAPIController struct{}

func (c *BaseAPIController) BodyParser(ctx *fiber.Ctx, out interface{}) error {
	if err := ctx.BodyParser(out); err != nil {
		log.WithError(err).Error("ошибка распознавания запроса")
		return errors.New("не удалось получить данные из запроса")
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
