# Localizer

i18n translation using **go-i18n**: load JSON message files from a directory and translate by key (with optional template data). Supports **error localization**: translate go-common `error.Error` by error code and optional namespace. Used by go-api for user-facing messages and HTTP error responses.

## Structure

| File        | Role                                                                 |
| ----------- | -------------------------------------------------------------------- |
| **main.go** | `Translator` interface (`Translate`, `TranslateError`), `NewTranslator(supportedLanguages, defaultLanguage, localesPath)`. |
| **mock.go** | Mock of `Translator` for tests.                                      |
| **main_test.go** | Tests with testdata locales.                                         |
| **testdata/locales/** | Sample JSON files (en, es, tr, error.json) for tests.        |

## Translator

- **Translate(lang, key, data)** — Returns the message for `key` in `lang`; falls back to default language if unsupported. `data` is optional template data.
- **TranslateError(lang, err, data)** — Fills `err.Message` (and optionally `err.ErrorCode` from a namespace like `error`) from the bundle so handlers can return localized errors.

## Usage

```go
import "github.com/yca-software/go-common/localizer"

tr := localizer.NewTranslator([]string{"en", "tr"}, "en", "/path/to/locales")
msg := tr.Translate("en", "welcome.message", nil)
tr.TranslateError("tr", domainErr, nil)
// domainErr.Message now holds translated string for error code
```

Locale files: one JSON per language (e.g. `en.json`); error codes often in a nested file (e.g. `error.json`) or namespace. See go-api `locales/` for layout.
