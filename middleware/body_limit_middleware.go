package middleware

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func WithBodyLimit(limit int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if strings.Contains(c.Path(), "public/survey/upload-answer") ||
			strings.Contains(c.Path(), "public/survey/upload-stream") {
			return c.Next()
		}
		contentLength := c.Get("Content-Length")
		if contentLength != "" && contentLength != "0" {
			size, err := strconv.ParseInt(contentLength, 10, 64)
			if err != nil && size > limit {
				return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
					"error": fmt.Sprintf("Request body too large. Maximum allowed: %d bytes", limit),
				})
			}
		}

		return c.Next()
	}
}
