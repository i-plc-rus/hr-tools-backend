package helpers

import (
	"context"
	"regexp"
	"strings"
)

func IsContextDone(ctx context.Context) bool {
	if ctx == nil {
		return true
	}
	select {
	case <-ctx.Done():
		return true
	default:
	}
	return false
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
