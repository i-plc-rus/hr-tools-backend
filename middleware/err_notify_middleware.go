package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

func ErrNotify(addr string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		statusCode := c.Response().StatusCode()

		if statusCode >= http.StatusInternalServerError {
			body := string(c.Response().Body())

			var data struct {
				Status  string `json:"status"`
				Message string `json:"message"`
			}
			unmErr := json.Unmarshal(c.Response().Body(), &data)
			if unmErr != nil {
				log.WithError(err).Warn("error unmarshalling response body in middleware")
			}

			method := c.Method()
			path := c.OriginalURL()
			if r := c.Route(); r != nil {
				path = r.Path
			}

			msg := data.Message
			if msg == "" {
				msg = body
			}

			go func() {
				payload := fmt.Sprintf(
					`{"code":%d,"method":%q,"path":%q,"error":%q}`,
					statusCode, method, path, msg)
				if _, reqErr := http.Post(addr, "application/json", strings.NewReader(payload)); reqErr != nil {
					log.WithError(reqErr).Warn("error sending error notification")
				}
			}()
		}

		return err
	}
}
