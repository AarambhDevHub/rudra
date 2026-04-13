package errors

import (
	"net/http"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

// ErrorHandler processes errors returned by handlers.
type ErrorHandler func(*rudraContext.Context, error)

// DefaultErrorHandler sends a JSON error response and never leaks internals.
func DefaultErrorHandler(c *rudraContext.Context, err error) {
	var re *RudraError
	if As(err, &re) {
		_ = c.JSON(re.Code, re)
		return
	}
	_ = c.JSON(http.StatusInternalServerError, &RudraError{
		Code:    500,
		Message: "internal server error",
	})
}