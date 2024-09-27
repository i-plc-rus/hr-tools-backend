package fiberlog

import "github.com/sirupsen/logrus"

// Config is config for middleware
type Config struct {
	Logger *logrus.Logger
	Tags   []string
}

// ConfigDefault is the default config
var ConfigDefault Config = Config{
	Logger: nil,
	Tags: []string{
		TagStatus,
		TagLatency,
		TagMethod,
		TagPath,
	},
}
