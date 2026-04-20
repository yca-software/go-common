# Logger

Structured logging with **zerolog**: console output (dev) or JSON (production), with **sanitization** of sensitive fields so DSNs, secrets, and tokens are never logged. Used by go-api and other services for request and audit logging.

## Structure

| File          | Role                                                                 |
| ------------- | -------------------------------------------------------------------- |
| **main.go**   | `LogData` struct, `Logger` interface (`Log`), `New()`, `NewWithWriter(w)` for tests. |
| **sanitize.go** | `SanitizeForLog(value, secretKeys []string)` — redacts known keys in maps/slices/strings. |
| **mock.go**   | Mock of `Logger` for tests.                                          |
| **main_test.go**, **sanitize_test.go** | Unit tests.                          |

## LogData

- **Level**, **RequestID**, **UserID**, **APIKeyID**, **Location** — Context.
- **Error** — Optional underlying error.
- **Message** — Log message.
- **Data** — Optional map/slice; sanitized when secret keys are provided.

Use **SecretKeys** (e.g. from config) when calling `Log` so that sensitive config values are redacted from `Data`.

## Usage

```go
import "github.com/yca-software/go-common/logger"

log := logger.New()
log.Log(logger.LogData{Level: "info", Message: "request", Data: map[string]any{"user_id": id}})
// With sanitization: pass secret keys so DSNs, API keys, etc. are redacted
```

`NewWithWriter(w)` is used in tests to capture log output for assertions.
