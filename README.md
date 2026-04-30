# 2Chi Go Common Libraries

Shared Go packages for the 2Chi framework, published from this standalone `go-common` repository.

## Project Policies

- Contribution guide: `CONTRIBUTING.md`
- Security policy: `SECURITY.md`
- Code of conduct: `CODE_OF_CONDUCT.md`
- License: `LICENSE`

## Requirements

- Go 1.26.2+

## Packages

| Package                                | Description                                                                            |
| -------------------------------------- | -------------------------------------------------------------------------------------- |
| [**database**](database/README.md)     | SQL client (Postgres via pgx/sqlx) with pool settings and optional ping timeout        |
| [**email**](email/README.md)           | Email sending (Resend) and HTML template rendering                                     |
| [**error**](error/README.md)           | Typed API errors (status code, error code, message, extra) with `Unwrap` and `AsError` |
| [**http**](http/README.md)             | HTTP helpers (e.g. `ParseLimitOffset`, `GetLanguage` from Accept-Language) for Echo    |
| [**localizer**](localizer/README.md)   | i18n translation (go-i18n) for messages and error localization                         |
| [**logger**](logger/README.md)         | Structured logging (zerolog) with sanitization of sensitive data                       |
| [**observer**](observer/README.md)     | Prometheus metrics (HTTP, DB) and `QueryMetricsHook` for repository observability      |
| [**password**](password/README.md)     | Argon2id hashing and constant-time comparison                                          |
| [**repository**](repository/README.md) | Generic repository over sqlx (squirrel, metrics hook, error wrapping)                  |
| [**token**](token/README.md)           | Secure random token generation and SHA256 hashing                                      |
| [**validator**](validator/README.md)   | Struct validation (go-playground/validator) with typed errors                          |

## Usage

Add the module to your project:

```bash
go get github.com/yca-software/go-common/database@latest
# or specific packages:
go get github.com/yca-software/go-common/error@latest
go get github.com/yca-software/go-common/logger@latest
```

Import by package path (no `pkg/` prefix):

```go
import (
    yca_error "github.com/yca-software/go-common/error"
    yca_log "github.com/yca-software/go-common/logger"
    yca_validate "github.com/yca-software/go-common/validator"
)
```

### Local development

When working on `go-common` and a consumer app in the same workspace, use a `replace` in the consumer's `go.mod`:

```go
replace github.com/yca-software/go-common => ../go-common
```

Then run `go mod tidy` and build/test from the consumer.

## Tests

```bash
go test ./... -count=1
```

Some packages (e.g. `database`, `repository`) use testcontainers for integration tests; ensure Docker is available.
