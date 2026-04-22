package middleware

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

// LoggerConfig defines the config for the Logger middleware.
type LoggerConfig struct {
	// Output is the writer for log output. Default: os.Stdout.
	Output io.Writer

	// Format selects the log format: "json" (default), "text", or "common" (Apache Combined Log).
	Format string

	// TimeFormat is the time format string. Default: time.RFC3339.
	TimeFormat string

	// SkipPaths is a list of URL paths to skip logging for (e.g. "/health", "/metrics").
	SkipPaths []string

	// Level is the slog level for request logs. Default: slog.LevelInfo.
	Level slog.Level

	// skipMap is the O(1) lookup map built from SkipPaths.
	skipMap map[string]struct{}
}

// DefaultLoggerConfig returns a LoggerConfig with sane defaults.
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Output:     os.Stdout,
		Format:     "json",
		TimeFormat: time.RFC3339,
		Level:      slog.LevelInfo,
	}
}

// Logger returns a structured access logging middleware.
//
// Logs: method, path, status, latency, IP, request_id, user_agent, bytes_written.
// Latency is measured around c.Next() for nanosecond accuracy.
//
// Formats:
//   - "json"   — structured JSON via log/slog (default)
//   - "text"   — structured text via log/slog
//   - "common" — Apache Combined Log Format
//
// Usage:
//
//	app.Use(middleware.Logger())
//	app.Use(middleware.Logger(middleware.LoggerConfig{
//	    Format:    "json",
//	    SkipPaths: []string{"/health", "/metrics"},
//	}))
func Logger(config ...LoggerConfig) func(*rudraContext.Context) error {
	cfg := DefaultLoggerConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Apply defaults for zero-value fields.
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	if cfg.Format == "" {
		cfg.Format = "json"
	}
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}

	// Build skip paths map for O(1) lookup.
	cfg.skipMap = make(map[string]struct{}, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		cfg.skipMap[p] = struct{}{}
	}

	// Create the slog logger based on format.
	var logger *slog.Logger
	switch strings.ToLower(cfg.Format) {
	case "text":
		logger = slog.New(slog.NewTextHandler(cfg.Output, &slog.HandlerOptions{
			Level: cfg.Level,
		}))
	case "common":
		// Apache Combined Log uses raw fmt.Fprintf, not slog.
		logger = nil
	default: // "json"
		logger = slog.New(slog.NewJSONHandler(cfg.Output, &slog.HandlerOptions{
			Level: cfg.Level,
		}))
	}

	return func(c *rudraContext.Context) error {
		path := c.Path()

		// Skip logging for configured paths.
		if _, skip := cfg.skipMap[path]; skip {
			return c.Next()
		}

		// Wrap the response writer to capture status code and bytes written.
		rw := newResponseWriter(c.Writer())
		c.SetWriter(rw)

		// Record start time.
		start := time.Now()

		// Execute the handler chain.
		err := c.Next()

		// Calculate latency.
		latency := time.Since(start)

		// Gather log fields.
		method := c.Method()
		status := rw.statusCode
		bytesOut := rw.bytesWritten
		clientIP := c.RealIP()
		userAgent := c.UserAgent()
		requestID := c.RequestID()

		if cfg.Format == "common" {
			// Apache Combined Log Format:
			// host ident authuser date request status bytes "referer" "user-agent"
			ts := time.Now().Format("02/Jan/2006:15:04:05 -0700")
			referer := c.Header("Referer")
			fmt.Fprintf(cfg.Output, "%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %v\n",
				clientIP, ts, method, path, c.Request().Proto,
				status, bytesOut, referer, userAgent, latency,
			)
		} else {
			// Structured log via slog.
			attrs := []slog.Attr{
				slog.String("method", method),
				slog.String("path", path),
				slog.Int("status", status),
				slog.Duration("latency", latency),
				slog.String("ip", clientIP),
				slog.String("user_agent", userAgent),
				slog.Int64("bytes_out", bytesOut),
			}
			if requestID != "" {
				attrs = append(attrs, slog.String("request_id", requestID))
			}

			args := make([]any, len(attrs))
			for i, a := range attrs {
				args[i] = a
			}
			logger.LogAttrs(c.Request().Context(), cfg.Level, "request", attrs...)
		}

		return err
	}
}