package errors

import (
	"errors"
	"fmt"
)

// RudraError is the standard error type for all HTTP errors in Rudra.
type RudraError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
	Cause   error  `json:"-"`
}

func (e *RudraError) Error() string {
	return fmt.Sprintf("rudra error %d: %s", e.Code, e.Message)
}

func (e *RudraError) Unwrap() error {
	return e.Cause
}

func newErr(code int, msg string, detail ...any) *RudraError {
	e := &RudraError{Code: code, Message: msg}
	if len(detail) > 0 {
		e.Detail = detail[0]
	}
	return e
}

// BadRequest returns a 400 RudraError.
func BadRequest(msg string, detail ...any) *RudraError { return newErr(400, msg, detail...) }

// Unauthorized returns a 401 RudraError.
func Unauthorized(msg string, detail ...any) *RudraError { return newErr(401, msg, detail...) }

// Forbidden returns a 403 RudraError.
func Forbidden(msg string, detail ...any) *RudraError { return newErr(403, msg, detail...) }

// NotFound returns a 404 RudraError.
func NotFound(msg string, detail ...any) *RudraError { return newErr(404, msg, detail...) }

// MethodNotAllowed returns a 405 RudraError.
func MethodNotAllowed(msg string, detail ...any) *RudraError { return newErr(405, msg, detail...) }

// Conflict returns a 409 RudraError.
func Conflict(msg string, detail ...any) *RudraError { return newErr(409, msg, detail...) }

// UnprocessableEntity returns a 422 RudraError.
func UnprocessableEntity(msg string, detail ...any) *RudraError { return newErr(422, msg, detail...) }

// TooManyRequests returns a 429 RudraError.
func TooManyRequests(msg string, detail ...any) *RudraError { return newErr(429, msg, detail...) }

// InternalServerError returns a 500 RudraError.
func InternalServerError(msg string, detail ...any) *RudraError { return newErr(500, msg, detail...) }

// NewHTTPError creates a RudraError with the given code and message.
func NewHTTPError(code int, msg string) *RudraError {
	return newErr(code, msg)
}

// As is a wrapper around errors.As for RudraError.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is is a wrapper around errors.Is for error comparison.
func Is(err, target error) bool {
	return errors.Is(err, target)
}