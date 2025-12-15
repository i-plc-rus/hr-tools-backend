package rbac

import (
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	"regexp"
	"strings"
	"slices"
)

type Provider interface {
	GetRuleFunc(method, path string) (models.RbacFunc, bool)
	RegisterRule(module models.Module, permission models.Permission, roles []models.UserRole, swaggerPattern string, handler models.RbacFunc) error
	GetPermissions(role models.UserRole) map[models.Module][]models.Permission
}

var Instance Provider

func NewHandler() {
	i := &impl{
		rules:       map[HTTPMethod]*PathRule{},
		permissions: map[models.UserRole]map[models.Module][]models.Permission{},
	}
	Instance = i
	i.initRules()
}

type impl struct {
	rules       map[HTTPMethod]*PathRule
	permissions map[models.UserRole]map[models.Module][]models.Permission
}

func (i *impl) GetRuleFunc(method, path string) (models.RbacFunc, bool) {
	normalizedPath := normalizePath(path)
	httpMethod := HTTPMethod(strings.ToUpper(method))

	if pathRule, exists := i.rules[httpMethod]; exists {
		if handler, found := i.findInPathRule(pathRule, normalizedPath); found {
			return handler, true
		}
	}

	return nil, false
}

func (i *impl) RegisterRule(module models.Module, permission models.Permission, roles []models.UserRole, swaggerPattern string, handler models.RbacFunc) error {
	path, method, err := parseSwaggerPattern(swaggerPattern)
	if err != nil {
		panic(err.Error())
	}

	// заполнение структуры для фронта
	for _, role := range roles {
		_, ok := i.permissions[role]
		if !ok {
			i.permissions[role] = map[models.Module][]models.Permission{}
		}
		permissions := i.permissions[role][module]
		found := slices.Contains(permissions, permission)
		if found {
			continue
		}

		i.permissions[role][module] = append(permissions, permission)
	}

	if _, exists := i.rules[method]; !exists {
		i.rules[method] = &PathRule{
			Exact:    make(map[string]models.RbacFunc),
			Patterns: []PatternRule{},
		}
	}

	// Заполняем правила для фильтрации
	if handler == nil {
		handler = AllowByRoleFunc(roles)
	}
	pathRule := i.rules[method]
	// Определяем тип пути и добавляем в соответствующую категорию
	if isExactPath(path) {
		pathRule.Exact[path] = handler
	} else {
		// Конвертируем путь в regexp
		pattern := pathToRegex(path)
		if pattern == nil {
			// Если не удалось скомпилировать, добавляем как точное совпадение
			pathRule.Exact[path] = handler
		} else {
			pathRule.Patterns = append(pathRule.Patterns, PatternRule{
				Pattern: pattern,
				Handler: handler,
			})
		}
	}

	return nil
}

func (i *impl) GetPermissions(role models.UserRole) map[models.Module][]models.Permission {
	return i.permissions[role]
}

func isExactPath(path string) bool {
	return !strings.Contains(path, "{")
}

func pathToRegex(path string) *regexp.Regexp {
	// Экранируем специальные символы
	pattern := regexp.QuoteMeta(path)

	// Заменяем экранированные { и } на оригинальные для обработки параметров
	pattern = strings.ReplaceAll(pattern, "\\{", "{")
	pattern = strings.ReplaceAll(pattern, "\\}", "}")

	// Заменяем {param}
	pattern = regexp.MustCompile(`\{[^}]+?\}`).ReplaceAllString(pattern, `([^/]+)`)

	pattern = strings.ReplaceAll(pattern, `\*`, `.*?`)
	pattern = "^" + pattern + "$"

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	return regex
}

func (i *impl) findInPathRule(pathRule *PathRule, path string) (models.RbacFunc, bool) {
	if pathRule == nil {
		return nil, false
	}

	// 1. Проверяем точные совпадения
	if handler, exists := pathRule.Exact[path]; exists {
		return handler, true
	}

	// 3. Проверяем regexp паттерны
	for _, patternRule := range pathRule.Patterns {
		if patternRule.Pattern.MatchString(path) {
			return patternRule.Handler, true
		}
	}

	return nil, false
}

func AllowFunc() models.RbacFunc {
	return func(spaceID, userID string, role models.UserRole, uri string) bool {
		return true
	}
}

func AllowByRoleFunc(accessRoles []models.UserRole) models.RbacFunc {
	allowMap := map[models.UserRole]bool{}
	for _, role := range accessRoles {
		allowMap[role] = true
	}
	return func(spaceID, userID string, role models.UserRole, uri string) bool {
		return allowMap[role]
	}
}

// парсит строку в формате "/api/v1/users [post]"
func parseSwaggerPattern(pattern string) (path string, method HTTPMethod, err error) {
	pattern = strings.TrimSpace(pattern)

	bracketStart := strings.LastIndex(pattern, "[")
	bracketEnd := strings.LastIndex(pattern, "]")

	if bracketStart != -1 && bracketEnd != -1 && bracketEnd > bracketStart {
		path = strings.TrimSpace(pattern[:bracketStart])

		methodsStr := pattern[bracketStart+1 : bracketEnd]
		method = HTTPMethod(strings.ToUpper(strings.TrimSpace(methodsStr)))
	} else {
		return "", "", errors.Errorf("Method not provided for pattern (%v)", pattern)
	}

	path = normalizePath(path)

	return path, method, nil
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	return path
}
