package helpers

import (
	"context"
	"regexp"
	"strings"
	"time"
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

func ParseHhTime(timeStr string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05-0700", timeStr)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func ParseAvitoTime(timeInt int64) time.Time {
	if timeInt == 0 {
		return time.Time{}
	}
	return time.Unix(timeInt, 0)
}
