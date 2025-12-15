package rbac

import (
	"hr-tools-backend/models"
	"regexp"
)

type MethodRule struct {
	Method  HTTPMethod
	Handler models.RbacFunc
}

type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
	PATCH  HTTPMethod = "PATCH"
	ALL    HTTPMethod = "ALL"
)

type PathRule struct {
	// проверки (от быстрых к медленным)
	Exact    map[string]models.RbacFunc // Точные совпадения
	Patterns []PatternRule              // Regexp правила
}

type PatternRule struct {
	Pattern *regexp.Regexp
	Handler models.RbacFunc
}
