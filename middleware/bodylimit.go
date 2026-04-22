package middleware

import (
	"net/http"

	rudraContext "github.com/AarambhDevHub/rudra/context"
	"github.com/AarambhDevHub/rudra/errors"
)

// BodyLimitConfig defines the config for the BodyLimit middleware.
type BodyLimitConfig struct {
	// Limit is the maximum allowed request body size in bytes.
	// Default: 32MB (32 << 20).
	Limit int64

	// OnLimit is an optional handler called when the body exceeds the limit.
	// If nil, the default 413 Payload Too Large response is returned.
	OnLimit func(c *rudraContext.Context) error
}

// DefaultBodyLimitConfig returns a BodyLimitConfig with sane defaults.
func DefaultBodyLimitConfig() BodyLimitConfig {
	return BodyLimitConfig{
		Limit: 32 << 20, // 32MB
	}
}

// BodyLimit returns a middleware that limits the maximum request body size.
//
// It wraps r.Body with http.MaxBytesReader, which returns an error when
// the body exceeds the configured limit. The middleware returns 413 Payload
// Too Large to the client.
//
// Usage:
//
//	app.Use(middleware.BodyLimit())                                     // 32MB default
//	app.Use(middleware.BodyLimit(middleware.BodyLimitConfig{Limit: 1 << 20}))  // 1MB
func BodyLimit(config ...BodyLimitConfig) func(*rudraContext.Context) error {
	cfg := DefaultBodyLimitConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 32 << 20
	}

	return func(c *rudraContext.Context) error {
		// Only limit body for methods that typically have a body.
		if c.Method() != http.MethodGet && c.Method() != http.MethodHead &&
			c.Method() != http.MethodOptions {
			// http.MaxBytesReader wraps the body and returns an error
			// when the limit is exceeded. It also sets a flag on the
			// response to close the connection after the response.
			c.Request().Body = http.MaxBytesReader(c.Writer(), c.Request().Body, cfg.Limit)
		}

		err := c.Next()

		// Check if the error is due to body exceeding the limit.
		if err != nil {
			// http.MaxBytesError is available in Go 1.19+
			if isMaxBytesError(err) {
				if cfg.OnLimit != nil {
					return cfg.OnLimit(c)
				}
				return errors.NewHTTPError(http.StatusRequestEntityTooLarge,
					"request body too large")
			}
		}

		return err
	}
}

// isMaxBytesError checks if the error is from http.MaxBytesReader.
func isMaxBytesError(err error) bool {
	// http.MaxBytesError was added in Go 1.19.
	// The error message from MaxBytesReader is "http: request body too large".
	if err == nil {
		return false
	}
	// Check for the *http.MaxBytesError type.
	_, ok := err.(*http.MaxBytesError)
	if ok {
		return true
	}
	// Fallback: check the error message for older Go versions.
	return err.Error() == "http: request body too large"
}
