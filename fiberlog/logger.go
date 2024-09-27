package fiberlog

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

// getLogrusFields calls FuncTag functions on matching keys
func getLogrusFields(ftm map[string]FuncTag, c *fiber.Ctx, d *data) log.Fields {
	f := make(log.Fields)
	for k, ft := range ftm {
		value := ft(c, d)
		strValue, ok := value.(string)
		if ok {
			if strValue != "" {
				f[k] = strValue
			}
		} else {
			f[k] = value
		}
	}
	return f
}

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
	var cfg Config
	if len(config) == 0 {
		cfg = ConfigDefault
	} else {
		cfg = config[0]
	}
	d := new(data)
	// Set PID once
	d.pid = os.Getpid()
	ftm := getFuncTagMap(cfg, d)
	return func(c *fiber.Ctx) error {
		d.start = time.Now()
		err := c.Next()
		d.end = time.Now()
		if c.Method() == "OPTIONS" {
			return err
		}

		message := getMessage(c)
		switch cfg.Logger {
		case nil:
			log.WithFields(getLogrusFields(ftm, c, d)).Info(message)
		default:
			entity := cfg.Logger.WithFields(getLogrusFields(ftm, c, d))
			if c.Response() != nil && c.Response().StatusCode() >= 300 {
				entity.Warn(message)
			} else {
				entity.Info(message)
			}
		}

		return err
	}
}

func getMessage(c *fiber.Ctx) string {
	return "запрос api"
}
