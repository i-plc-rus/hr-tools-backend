package helpers

import (
	"context"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/h2non/filetype"
)

const (
	HeaderLogIgnore = "X-Content-Log"
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

func GetFileContentType(file *multipart.FileHeader) string {
	if types, ok := file.Header["Content-Type"]; ok && len(types) > 0 {
		return types[0]
	}
	return ""
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

func DetectFileContentType(filename string, data []byte) string {
	// пробуем определить по содержимому
	if mime := detectTypeFromContent(data); mime != "application/octet-stream" {
		return mime
	}

	// по расширению файла
	return detectFromExtension(filename)
}

type MimeDetector struct{}

func detectTypeFromContent(data []byte) string {
	kind, err := filetype.Match(data)
	if err == nil && kind.MIME.Value != "" {
		return kind.MIME.Value
	}

	return http.DetectContentType(data)
}

func detectFromExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	mimeTypes := map[string]string{
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		".webm": "video/webm",
		".mkv":  "video/x-matroska",
		".m4v":  "video/x-m4v",
		".3gp":  "video/3gpp",
		".ts":   "video/mp2t",
		".mpeg": "video/mpeg",
		".mpg":  "video/mpeg",
	}

	if mime, exists := mimeTypes[ext]; exists {
		return mime
	}

	return "application/octet-stream"
}

func WithRetry[T any](tries int, delaySec int, fn func() (T, error)) (T, error) {
	var zero T
	if tries <= 0 {
		tries = 1
	}

	var lastErr error
	for range tries {
		if val, err := fn(); err == nil {
			return val, nil
		} else {
			time.Sleep(time.Second * time.Duration(delaySec))
			lastErr = err
		}
	}
	return zero, lastErr
}
