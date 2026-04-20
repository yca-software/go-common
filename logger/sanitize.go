package logger

import (
	"regexp"
	"strings"
)

// Redact patterns that may appear in error messages (connection strings, DSNs, etc.).
var redactPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(password|passwd|pwd)=[^\s&]+`),
	regexp.MustCompile(`(?i)(secret|token|apikey|api_key)=[^\s&]+`),
	regexp.MustCompile(`(?i)(:[^:@]+)@`),           // :password@ in URLs
	regexp.MustCompile(`postgres(?:ql)?://[^\s]+`), // PostgreSQL DSN
	regexp.MustCompile(`redis(?:s)?://[^\s]+`),     // Redis URL
	regexp.MustCompile(`amqp(?:s)?://[^\s]+`),      // RabbitMQ URL
}

// Sensitive keys (case-insensitive) whose values are always redacted in SanitizeData.
var sensitiveKeys = map[string]bool{
	"password": true, "passwd": true, "pwd": true,
	"secret": true, "token": true, "apikey": true, "api_key": true,
	"authorization": true, "cookie": true, "session": true,
}

const redactedPlaceholder = "[REDACTED]"

var collapseRedactedRe = regexp.MustCompile(`(\[REDACTED\]\s*)+`)

// redactString applies redactPatterns to s and collapses consecutive placeholders.
func redactString(s string) string {
	for _, re := range redactPatterns {
		s = re.ReplaceAllString(s, redactedPlaceholder)
	}
	s = collapseRedactedRe.ReplaceAllString(s, redactedPlaceholder+" ")
	return strings.TrimSpace(s)
}

// SanitizeError returns an error that wraps err with a message that has sensitive patterns redacted.
// Use when logging errors that may contain connection strings, passwords, or other secrets.
// If err is nil, returns nil.
func SanitizeError(err error) error {
	if err == nil {
		return nil
	}
	msg := redactString(err.Error())
	if msg == "" {
		msg = "[error message redacted]"
	}
	return &sanitizedError{msg: msg, original: err}
}

// SanitizeData returns a copy of data with sensitive values redacted. Use when logging LogData.Data.
// - nil returns nil.
// - string: redacted by pattern matching.
// - map[string]any: values for sensitive keys (e.g. password, token) replaced with [REDACTED]; other values recursively sanitized.
// - []any: each element recursively sanitized.
// - Other types (numbers, bools, structs) are returned as-is.
func SanitizeData(data any) any {
	if data == nil {
		return nil
	}
	switch v := data.(type) {
	case string:
		return redactString(v)
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, val := range v {
			if sensitiveKeys[strings.ToLower(k)] {
				out[k] = redactedPlaceholder
			} else {
				out[k] = SanitizeData(val)
			}
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = SanitizeData(val)
		}
		return out
	default:
		return data
	}
}

type sanitizedError struct {
	msg      string
	original error
}

func (e *sanitizedError) Error() string { return e.msg }
func (e *sanitizedError) Unwrap() error { return e.original }
