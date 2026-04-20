# HTTP

Helpers for Echo-based HTTP handlers: query parsing and Accept-Language handling. Keeps handler code DRY and consistent across go-api routes.

## Structure

| File          | Role                                                                 |
| ------------- | -------------------------------------------------------------------- |
| **language.go** | `ParseLimitOffset(c, defaultLimit, maxLimit)`, `GetLanguage(c, supported, default)`. |
| **language_test.go** | Unit tests for both functions.                              |

## Functions

- **ParseLimitOffset** — Reads `limit` and `offset` from query params. Constrains `limit` to `[1, maxLimit]`; `offset` ≥ 0. Returns defaults when context is nil or params missing/invalid.
- **GetLanguage** — Extracts language from `Accept-Language` (first two characters), normalizes and checks against `supportedLanguages`; returns `defaultLanguage` when unsupported or missing.

## Usage

```go
import "github.com/yca-software/go-common/http"

// In handler
limit, offset := http.ParseLimitOffset(c, 20, 100)
lang := http.GetLanguage(c, []string{"en", "tr", "es"}, "en")
```

Use `GetLanguage` when calling the translator for error messages or localized content.
