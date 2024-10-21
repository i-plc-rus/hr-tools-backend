package initializers

import (
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/fiberlog"
)

func InitLogger() *fiberlog.Config {
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyTime: "@timestamp",
			log.FieldKeyMsg:  "message",
		},
	})
	log.SetLevel(log.InfoLevel)

	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyTime: "@timestamp",
			log.FieldKeyMsg:  "message",
		},
	})
	logger.SetLevel(log.DebugLevel)
	return &fiberlog.Config{
		Logger: logger,
		Tags: []string{
			fiberlog.TagBody,
			fiberlog.TagResBody,
			fiberlog.TagMethod,
			fiberlog.TagPath,
			fiberlog.TagStatus,
			fiberlog.RequestID,
		},
	}
}
