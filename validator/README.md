# Validator

Struct validation using **go-playground/validator**: validate request DTOs and return a map of field → **ValidationError** (tag, param, value, message). Used by go-api services to validate input before calling repositories; handlers map validation errors to 422 or 400 with go-common error.

## Structure

| File        | Role                                                                 |
| ----------- | -------------------------------------------------------------------- |
| **main.go** | `ValidationError` struct, `Validator` interface (`ValidateStruct(s)`), `New()`. |
| **main_test.go** | Tests for valid/invalid structs and error shape.                     |

## ValidationError

- **Tag** — Validator tag that failed (e.g. `required`, `email`).
- **Param** — Tag parameter if any.
- **Value** — Field value (for debugging; may be omitted in API response).
- **Error** — Full validation message.

`ValidateStruct(s)` returns `nil` when valid; otherwise a `*map[string]ValidationError` keyed by field name (or `""` for non-field errors).

## Usage

```go
import "github.com/yca-software/go-common/validator"

v := validator.New()
errs := v.ValidateStruct(req)
if errs != nil {
    return yca_error.NewUnprocessableEntityError(nil, "VALIDATION_FAILED", errs)
}
```

Use struct tags (e.g. `validate:"required,email"`) on request types. Prefer this package over ad-hoc validation so error shape and HTTP mapping stay consistent.
