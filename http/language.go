package http

import (
	"slices"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// ParseLimitOffset parses limit and offset from query params. defaultLimit and maxLimit (e.g. 20 and 100) constrain limit; offset defaults to 0 and must be >= 0.
func ParseLimitOffset(c echo.Context, defaultLimit, maxLimit int) (limit, offset int) {
	limit = defaultLimit
	offset = 0
	if c == nil {
		return limit, offset
	}
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= maxLimit {
			limit = parsed
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return limit, offset
}

func GetLanguage(c echo.Context, supportedLanguages []string, defaultLanguage string) string {
	if c == nil {
		return strings.ToLower(strings.TrimSpace(defaultLanguage))
	}

	lang := c.Request().Header.Get("Accept-Language")
	extracted := ""
	if len(lang) >= 2 {
		extracted = lang[:2]
	} else {
		extracted = lang
	}
	extracted = strings.ToLower(strings.TrimSpace(extracted))

	normalized := make([]string, 0, len(supportedLanguages))
	for _, l := range supportedLanguages {
		normalized = append(normalized, strings.ToLower(strings.TrimSpace(l)))
	}
	defaultLang := strings.ToLower(strings.TrimSpace(defaultLanguage))

	if extracted != "" && slices.Contains(normalized, extracted) {
		return extracted
	}
	return defaultLang
}
