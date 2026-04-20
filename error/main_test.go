package error_helpers_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	error_helpers "github.com/yca-software/go-common/error"
)

type ErrorTestSuite struct {
	suite.Suite
}

func TestErrorTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorTestSuite))
}

func (s *ErrorTestSuite) TestError_Error() {
	// Test case 1: No message, no underlying error
	err := error_helpers.NewBadRequestError(nil, "TEST_CODE", nil)
	assert.Equal(s.T(), "", err.Error())

	// Test case 2: No message, but has underlying error
	err = error_helpers.NewBadRequestError(errors.New("underlying error"), "TEST_CODE", nil)
	assert.Equal(s.T(), "underlying error", err.Error())

	// Test case 3: Has message, takes precedence over underlying error
	err.Message = "Custom error message"
	assert.Equal(s.T(), "Custom error message", err.Error())

	// Test case 4: Has message, no underlying error
	err2 := error_helpers.NewInternalServerError(nil, "TEST_CODE", nil)
	err2.Message = "Explicit message"
	assert.Equal(s.T(), "Explicit message", err2.Error())

	// Test case 5: Empty message string, has underlying error
	err3 := error_helpers.NewNotFoundError(errors.New("db error"), "TEST_CODE", nil)
	err3.Message = "" // Explicitly set to empty string
	assert.Equal(s.T(), "db error", err3.Error())
}

func (s *ErrorTestSuite) TestError_Error_WithUnderlyingError() {
	underlyingErr := errors.New("database connection failed")
	err := error_helpers.NewInternalServerError(underlyingErr, "", nil)

	// When Message is empty, Error() should return underlying error message
	assert.Equal(s.T(), underlyingErr.Error(), err.Error())
}

func (s *ErrorTestSuite) TestNewInternalServerError() {
	err := error_helpers.NewInternalServerError(nil, "", nil)
	assert.Equal(s.T(), 500, err.StatusCode)
	assert.Equal(s.T(), "INTERNAL_SERVER_ERROR", err.ErrorCode)
	assert.Nil(s.T(), err.Err)
	assert.Nil(s.T(), err.Extra)
}

func (s *ErrorTestSuite) TestNewInternalServerError_CustomCode() {
	err := error_helpers.NewInternalServerError(nil, "CUSTOM_CODE", nil)
	assert.Equal(s.T(), 500, err.StatusCode)
	assert.Equal(s.T(), "CUSTOM_CODE", err.ErrorCode)
}

func (s *ErrorTestSuite) TestNewUnprocessableEntityError() {
	err := error_helpers.NewUnprocessableEntityError(nil, "", nil)
	assert.Equal(s.T(), 422, err.StatusCode)
	assert.Equal(s.T(), "UNPROCESSABLE_ENTITY", err.ErrorCode)
}

func (s *ErrorTestSuite) TestNewConflictError() {
	err := error_helpers.NewConflictError(nil, "", nil)
	assert.Equal(s.T(), 409, err.StatusCode)
	assert.Equal(s.T(), "CONFLICT", err.ErrorCode)
}

func (s *ErrorTestSuite) TestNewNotFoundError() {
	err := error_helpers.NewNotFoundError(nil, "", nil)
	assert.Equal(s.T(), 404, err.StatusCode)
	assert.Equal(s.T(), "NOT_FOUND", err.ErrorCode)
}

func (s *ErrorTestSuite) TestNewUnauthorizedError() {
	err := error_helpers.NewUnauthorizedError(nil, "", nil)
	assert.Equal(s.T(), 401, err.StatusCode)
	assert.Equal(s.T(), "UNAUTHORIZED", err.ErrorCode)
}

func (s *ErrorTestSuite) TestNewBadRequestError() {
	err := error_helpers.NewBadRequestError(nil, "", nil)
	assert.Equal(s.T(), 400, err.StatusCode)
	assert.Equal(s.T(), "BAD_REQUEST", err.ErrorCode)
}

func (s *ErrorTestSuite) TestNewForbiddenError() {
	err := error_helpers.NewForbiddenError(nil, "", nil)
	assert.Equal(s.T(), 403, err.StatusCode)
	assert.Equal(s.T(), "FORBIDDEN", err.ErrorCode)
}

func (s *ErrorTestSuite) TestError_WithExtra() {
	extra := map[string]any{"field": "email", "reason": "invalid format"}
	err := error_helpers.NewBadRequestError(nil, "INVALID_EMAIL", extra)

	assert.Equal(s.T(), extra, err.Extra)
}

func (s *ErrorTestSuite) TestError_WithUnderlyingError() {
	underlyingErr := errors.New("database error")
	err := error_helpers.NewInternalServerError(underlyingErr, "DB_ERROR", nil)

	assert.Equal(s.T(), underlyingErr, err.Err)
}

func (s *ErrorTestSuite) TestError_JSONMarshaling() {
	// Test JSON marshaling with all fields
	err := error_helpers.NewBadRequestError(errors.New("validation failed"), "INVALID_INPUT", map[string]any{"field": "email"})
	err.Message = "Email is invalid"

	jsonData, marshalErr := json.Marshal(err)
	assert.NoError(s.T(), marshalErr)

	var result map[string]any
	unmarshalErr := json.Unmarshal(jsonData, &result)
	assert.NoError(s.T(), unmarshalErr)

	// Verify JSON contains expected fields
	assert.Equal(s.T(), "INVALID_INPUT", result["errorCode"])
	assert.Equal(s.T(), "Email is invalid", result["message"])
	assert.Equal(s.T(), map[string]any{"field": "email"}, result["extra"])

	// Verify StatusCode and Err are not in JSON (they have json:"-" tag)
	assert.NotContains(s.T(), result, "statusCode")
	assert.NotContains(s.T(), result, "err")
}

func (s *ErrorTestSuite) TestError_JSONMarshaling_WithoutExtra() {
	// Test JSON marshaling without Extra field (should omit it)
	err := error_helpers.NewBadRequestError(nil, "TEST_CODE", nil)
	err.Message = "Test message"

	jsonData, marshalErr := json.Marshal(err)
	assert.NoError(s.T(), marshalErr)

	var result map[string]any
	unmarshalErr := json.Unmarshal(jsonData, &result)
	assert.NoError(s.T(), unmarshalErr)

	// Verify Extra is not present when nil
	assert.NotContains(s.T(), result, "extra")
	assert.Equal(s.T(), "TEST_CODE", result["errorCode"])
	assert.Equal(s.T(), "Test message", result["message"])
}

func (s *ErrorTestSuite) TestError_AllParameterCombinations() {
	testCases := []struct {
		name       string
		createFunc func(error, string, any) *error_helpers.Error
		err        error
		errorCode  string
		extra      any
	}{
		{"NilErr_EmptyCode_NilExtra", error_helpers.NewBadRequestError, nil, "", nil},
		{"NilErr_CustomCode_NilExtra", error_helpers.NewBadRequestError, nil, "CUSTOM", nil},
		{"NilErr_EmptyCode_WithExtra", error_helpers.NewBadRequestError, nil, "", map[string]any{"key": "value"}},
		{"NilErr_CustomCode_WithExtra", error_helpers.NewBadRequestError, nil, "CUSTOM", map[string]any{"key": "value"}},
		{"WithErr_EmptyCode_NilExtra", error_helpers.NewBadRequestError, errors.New("test"), "", nil},
		{"WithErr_CustomCode_NilExtra", error_helpers.NewBadRequestError, errors.New("test"), "CUSTOM", nil},
		{"WithErr_EmptyCode_WithExtra", error_helpers.NewBadRequestError, errors.New("test"), "", map[string]any{"key": "value"}},
		{"WithErr_CustomCode_WithExtra", error_helpers.NewBadRequestError, errors.New("test"), "CUSTOM", map[string]any{"key": "value"}},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := tc.createFunc(tc.err, tc.errorCode, tc.extra)
			assert.NotNil(s.T(), err)
			assert.Equal(s.T(), tc.err, err.Err)
			assert.Equal(s.T(), tc.extra, err.Extra)

			if tc.errorCode == "" {
				// Should use default code for the error type
				assert.NotEmpty(s.T(), err.ErrorCode)
			} else {
				assert.Equal(s.T(), tc.errorCode, err.ErrorCode)
			}
		})
	}
}

func (s *ErrorTestSuite) TestError_MessageField() {
	err := error_helpers.NewBadRequestError(nil, "TEST_CODE", nil)

	// Initially empty
	assert.Equal(s.T(), "", err.Message)

	// Can be set
	err.Message = "Custom message"
	assert.Equal(s.T(), "Custom message", err.Message)
	assert.Equal(s.T(), "Custom message", err.Error())
}

func (s *ErrorTestSuite) TestError_StatusCodeField() {
	testCases := []struct {
		name       string
		createFunc func(error, string, any) *error_helpers.Error
		expected   int
	}{
		{"BadRequest", error_helpers.NewBadRequestError, 400},
		{"Unauthorized", error_helpers.NewUnauthorizedError, 401},
		{"PaymentRequired", error_helpers.NewPaymentRequiredError, 402},
		{"Forbidden", error_helpers.NewForbiddenError, 403},
		{"NotFound", error_helpers.NewNotFoundError, 404},
		{"Conflict", error_helpers.NewConflictError, 409},
		{"UnprocessableEntity", error_helpers.NewUnprocessableEntityError, 422},
		{"TooManyRequests", error_helpers.NewTooManyRequestsError, 429},
		{"InternalServerError", error_helpers.NewInternalServerError, 500},
		{"ServiceUnavailable", error_helpers.NewServiceUnavailableError, 503},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := tc.createFunc(nil, "", nil)
			assert.Equal(s.T(), tc.expected, err.StatusCode)
		})
	}
}

func (s *ErrorTestSuite) TestNewPaymentRequiredError() {
	err := error_helpers.NewPaymentRequiredError(nil, "", nil)
	assert.Equal(s.T(), 402, err.StatusCode)
	assert.Equal(s.T(), "PAYMENT_REQUIRED", err.ErrorCode)
}

func (s *ErrorTestSuite) TestNewTooManyRequestsError() {
	err := error_helpers.NewTooManyRequestsError(nil, "", nil)
	assert.Equal(s.T(), 429, err.StatusCode)
	assert.Equal(s.T(), "TOO_MANY_REQUESTS", err.ErrorCode)
}

func (s *ErrorTestSuite) TestNewServiceUnavailableError() {
	err := error_helpers.NewServiceUnavailableError(nil, "", nil)
	assert.Equal(s.T(), 503, err.StatusCode)
	assert.Equal(s.T(), "SERVICE_UNAVAILABLE", err.ErrorCode)
}

func (s *ErrorTestSuite) TestNewServiceUnavailableError_CustomCode() {
	err := error_helpers.NewServiceUnavailableError(nil, "CUSTOM_CODE", nil)
	assert.Equal(s.T(), 503, err.StatusCode)
	assert.Equal(s.T(), "CUSTOM_CODE", err.ErrorCode)
}

func (s *ErrorTestSuite) TestError_AllStatusCodes() {
	testCases := []struct {
		name       string
		createFunc func(error, string, any) *error_helpers.Error
		expected   int
	}{
		{"BadRequest", error_helpers.NewBadRequestError, 400},
		{"Unauthorized", error_helpers.NewUnauthorizedError, 401},
		{"PaymentRequired", error_helpers.NewPaymentRequiredError, 402},
		{"Forbidden", error_helpers.NewForbiddenError, 403},
		{"NotFound", error_helpers.NewNotFoundError, 404},
		{"Conflict", error_helpers.NewConflictError, 409},
		{"UnprocessableEntity", error_helpers.NewUnprocessableEntityError, 422},
		{"TooManyRequests", error_helpers.NewTooManyRequestsError, 429},
		{"InternalServerError", error_helpers.NewInternalServerError, 500},
		{"ServiceUnavailable", error_helpers.NewServiceUnavailableError, 503},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := tc.createFunc(nil, "TEST_CODE", nil)
			assert.Equal(s.T(), tc.expected, err.StatusCode)
			assert.Equal(s.T(), "TEST_CODE", err.ErrorCode)
		})
	}
}

func (s *ErrorTestSuite) TestError_DefaultErrorCode() {
	testCases := []struct {
		name       string
		createFunc func(error, string, any) *error_helpers.Error
		expected   string
	}{
		{"BadRequest", error_helpers.NewBadRequestError, "BAD_REQUEST"},
		{"Unauthorized", error_helpers.NewUnauthorizedError, "UNAUTHORIZED"},
		{"PaymentRequired", error_helpers.NewPaymentRequiredError, "PAYMENT_REQUIRED"},
		{"Forbidden", error_helpers.NewForbiddenError, "FORBIDDEN"},
		{"NotFound", error_helpers.NewNotFoundError, "NOT_FOUND"},
		{"Conflict", error_helpers.NewConflictError, "CONFLICT"},
		{"UnprocessableEntity", error_helpers.NewUnprocessableEntityError, "UNPROCESSABLE_ENTITY"},
		{"TooManyRequests", error_helpers.NewTooManyRequestsError, "TOO_MANY_REQUESTS"},
		{"InternalServerError", error_helpers.NewInternalServerError, "INTERNAL_SERVER_ERROR"},
		{"ServiceUnavailable", error_helpers.NewServiceUnavailableError, "SERVICE_UNAVAILABLE"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := tc.createFunc(nil, "", nil)
			assert.Equal(s.T(), tc.expected, err.ErrorCode)
		})
	}
}

func (s *ErrorTestSuite) TestError_Unwrap() {
	underlying := errors.New("db failure")
	err := error_helpers.NewInternalServerError(underlying, "DB_ERROR", nil)

	unwrapped := errors.Unwrap(err)
	assert.Equal(s.T(), underlying, unwrapped)
	assert.True(s.T(), errors.Is(err, underlying))
}

func (s *ErrorTestSuite) TestError_Unwrap_NilUnderlying() {
	err := error_helpers.NewBadRequestError(nil, "", nil)
	assert.Nil(s.T(), errors.Unwrap(err))
}

func (s *ErrorTestSuite) TestAsError() {
	apiErr := error_helpers.NewNotFoundError(errors.New("missing"), "RESOURCE_NOT_FOUND", nil)
	apiErr.Message = "User not found"

	e, ok := error_helpers.AsError(apiErr)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), apiErr, e)
	assert.Equal(s.T(), 404, e.StatusCode)
	assert.Equal(s.T(), "RESOURCE_NOT_FOUND", e.ErrorCode)
	assert.Equal(s.T(), "User not found", e.Message)
}

func (s *ErrorTestSuite) TestAsError_Wrapped() {
	apiErr := error_helpers.NewForbiddenError(errors.New("denied"), "FORBIDDEN", nil)
	wrapped := errors.Join(apiErr, errors.New("extra"))

	e, ok := error_helpers.AsError(wrapped)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), apiErr, e)
	assert.Equal(s.T(), 403, e.StatusCode)
}

func (s *ErrorTestSuite) TestAsError_PlainError() {
	plain := errors.New("plain error")
	e, ok := error_helpers.AsError(plain)
	assert.False(s.T(), ok)
	assert.Nil(s.T(), e)
}
