package error_helpers

import (
	"errors"
	"net/http"
)

type Error struct {
	StatusCode int    `json:"-"`
	Err        error  `json:"-"` // Internal error - not exposed in JSON responses
	ErrorCode  string `json:"errorCode"`
	Message    string `json:"message"`
	Extra      any    `json:"extra,omitempty"`
}

func (e *Error) Error() string {
	if e.Message == "" && e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *Error) Unwrap() error {
	return e.Err
}

// AsError returns *Error if err is or wraps this type; otherwise (nil, false).
// Handlers can use it to respond with StatusCode and JSON body.
func AsError(err error) (*Error, bool) {
	var e *Error
	ok := errors.As(err, &e)
	return e, ok
}

func newError(statusCode int, defaultCode string, err error, errorCode string, extra any) *Error {
	code := defaultCode
	if errorCode != "" {
		code = errorCode
	}
	return &Error{
		StatusCode: statusCode,
		Err:        err,
		ErrorCode:  code,
		Extra:      extra,
	}
}

func NewInternalServerError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err, errorCode, extra)
}

func NewUnprocessableEntityError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", err, errorCode, extra)
}

func NewConflictError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusConflict, "CONFLICT", err, errorCode, extra)
}

func NewNotFoundError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusNotFound, "NOT_FOUND", err, errorCode, extra)
}

func NewUnauthorizedError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusUnauthorized, "UNAUTHORIZED", err, errorCode, extra)
}

func NewBadRequestError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusBadRequest, "BAD_REQUEST", err, errorCode, extra)
}

func NewForbiddenError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusForbidden, "FORBIDDEN", err, errorCode, extra)
}

func NewPaymentRequiredError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusPaymentRequired, "PAYMENT_REQUIRED", err, errorCode, extra)
}

func NewTooManyRequestsError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusTooManyRequests, "TOO_MANY_REQUESTS", err, errorCode, extra)
}

func NewServiceUnavailableError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", err, errorCode, extra)
}
