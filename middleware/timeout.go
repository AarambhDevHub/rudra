package middleware

import (
	"context"
	"time"

	rudraContext "github.com/AarambhDevHub/rudra/context"
	"github.com/AarambhDevHub/rudra/errors"
)

// TimeoutConfig defines the config for the Timeout middleware.
type TimeoutConfig struct {
	// Timeout is the maximum duration for a single request.
	// Default: 30s.
	Timeout time.Duration

	// OnTimeout is an optional handler called when a request times out.
	// If nil, the default 503 Service Unavailable JSON response is sent.
	OnTimeout func(c *rudraContext.Context) error
}

// DefaultTimeoutConfig returns a TimeoutConfig with sane defaults.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout: 30 * time.Second,
	}
}

// Timeout returns a per-request timeout middleware.
//
// It wraps the request context with context.WithTimeout and monitors for
// deadline exceedance. If the handler does not complete within the configured
// timeout, the middleware returns 503 Service Unavailable.
//
// The timeout context is propagated through r.Context(), making it compatible
// with downstream database and HTTP client timeouts that respect context deadlines.
//
// Usage:
//
//	app.Use(middleware.Timeout())
//	app.Use(middleware.Timeout(middleware.TimeoutConfig{
//	    Timeout: 5 * time.Second,
//	    OnTimeout: func(c *context.Context) error {
//	        return c.JSON(503, map[string]string{"error": "custom timeout"})
//	    },
//	}))
func Timeout(config ...TimeoutConfig) func(*rudraContext.Context) error {
	cfg := DefaultTimeoutConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	return func(c *rudraContext.Context) error {
		// Create a deadline context derived from the request context.
		ctx, cancel := context.WithTimeout(c.Request().Context(), cfg.Timeout)
		defer cancel()

		// Replace the request with one that carries the deadline context.
		c.SetRequest(c.Request().WithContext(ctx))

		// Run the handler chain in a separate goroutine so we can select on timeout.
		// All writes from the handler goroutine go through the context's normal
		// writer. We synchronize via channels — the handler result is consumed
		// exactly once, either as success or as timeout.
		type result struct {
			err   error
			panic any
		}
		ch := make(chan result, 1)

		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					ch <- result{panic: rec}
				}
			}()
			ch <- result{err: c.Next()}
		}()

		select {
		case res := <-ch:
			// Handler completed before timeout.
			if res.panic != nil {
				panic(res.panic) // re-panic — let Recovery middleware handle it
			}
			return res.err

		case <-ctx.Done():
			// Timeout exceeded. The handler goroutine may still be running,
			// but the context cancellation will signal it to stop via
			// r.Context().Done(). We return the timeout error immediately.
			//
			// Note: The handler goroutine may still write to the response
			// after this point. In production, the net/http server's
			// WriteTimeout guards against this. For the middleware layer,
			// the timeout response takes priority since we return first.
			if cfg.OnTimeout != nil {
				return cfg.OnTimeout(c)
			}

			return errors.NewHTTPError(503, "service unavailable: request timeout")
		}
	}
}
