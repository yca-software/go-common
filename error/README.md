# Error

Typed API errors with HTTP status code, error code, message, and optional extra data. Supports `Unwrap` for `errors.Is` / `errors.As` and `AsError(err)` for handlers to respond with the correct status and JSON body. Used by go-api services and HTTP error handling.

## Structure

| File        | Role                                                                 |
| ----------- | -------------------------------------------------------------------- |
| **main.go** | `Error` struct, `AsError(err)`, constructors: `NewInternalServerError`, `NewUnprocessableEntityError`, `NewConflictError`, `NewNotFoundError`, `NewUnauthorizedError`, `NewBadRequestError`, `NewForbiddenError`, `NewPaymentRequiredError`. |
| **main_test.go** | Tests for `AsError`, `Unwrap`, and constructor behaviour.     |

## Error type

- **StatusCode** — HTTP status (not serialized in JSON).
- **Err** — Internal error for logging/unwrap (not exposed in JSON).
- **ErrorCode** — Stable code for clients (e.g. `NOT_FOUND`, `UNAUTHORIZED`).
- **Message** — Human-readable message (often overridden by translator in handlers).
- **Extra** — Optional payload (e.g. validation details).

## Usage

```go
import yca_error "github.com/yca-software/go-common/error"

// In service
if user == nil {
    return yca_error.NewNotFoundError(nil, "USER_NOT_FOUND", nil)
}

// In handler
if e, ok := yca_error.AsError(err); ok {
    return c.JSON(e.StatusCode, map[string]any{"errorCode": e.ErrorCode, "message": translated})
}
```

Handlers should use a **translator** (localizer) to replace `Message` with a localized string before responding.
