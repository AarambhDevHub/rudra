package middleware

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"

	rudraContext "github.com/AarambhDevHub/rudra/context"
	"github.com/AarambhDevHub/rudra/errors"
)

// RecoveryConfig defines the config for the Recovery middleware.
type RecoveryConfig struct {
	// LogStackTrace controls whether the stack trace is logged on panic.
	// Default: true.
	LogStackTrace bool

	// Output is the writer for stack trace output. Default: os.Stderr.
	Output io.Writer

	// OnPanic is an optional hook called when a panic is recovered.
	// Receives the context, the panic value, and the stack trace bytes.
	// This is called BEFORE the error response is sent to the client.
	OnPanic func(c *rudraContext.Context, err any, stack []byte)
}

// DefaultRecoveryConfig returns a RecoveryConfig with sane defaults.
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		LogStackTrace: true,
		Output:        os.Stderr,
	}
}

// Recovery returns a panic recovery middleware.
//
// It wraps the handler chain in a defer/recover block. On panic:
//   - Captures the panic value and full stack trace
//   - Logs the stack trace to the configured writer (never sent to client)
//   - Returns 500 Internal Server Error JSON to the client
//   - Server continues running — does not crash
//
// Usage:
//
//	app.Use(middleware.Recovery())
//	app.Use(middleware.Recovery(middleware.RecoveryConfig{
//	    LogStackTrace: true,
//	    OnPanic: func(c *context.Context, err any, stack []byte) {
//	        // send alert to monitoring
//	    },
//	}))
func Recovery(config ...RecoveryConfig) func(*rudraContext.Context) error {
	cfg := DefaultRecoveryConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	logger := slog.New(slog.NewJSONHandler(cfg.Output, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	return func(c *rudraContext.Context) error {
		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()

				// Log the stack trace.
				if cfg.LogStackTrace {
					logger.Error("panic recovered",
						slog.String("error", fmt.Sprintf("%v", rec)),
						slog.String("stack", string(stack)),
						slog.String("method", c.Method()),
						slog.String("path", c.Path()),
						slog.String("ip", c.RealIP()),
					)
				}

				// Call the custom panic hook if provided.
				if cfg.OnPanic != nil {
					cfg.OnPanic(c, rec, stack)
				}

				// Send 500 JSON to the client — never leak stack trace.
				// We use the error handler pattern to ensure consistent error format.
				_ = c.JSON(500, &errors.RudraError{
					Code:    500,
					Message: "internal server error",
				})
			}
		}()

		return c.Next()
	}
}
