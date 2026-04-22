package middleware

import (
	"crypto/rand"
	"fmt"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

// RequestIDConfig defines the config for the RequestID middleware.
type RequestIDConfig struct {
	// Generator is a function that returns a unique request ID.
	// Default: UUID v4 via crypto/rand.
	Generator func() string

	// Header is the HTTP header name used to read/write the request ID.
	// Default: "X-Request-ID".
	Header string
}

// DefaultRequestIDConfig returns a RequestIDConfig with sane defaults.
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Generator: generateUUIDv4,
		Header:    "X-Request-ID",
	}
}

// RequestID returns a middleware that generates or forwards a unique request ID.
//
// On every request:
//  1. Reads the configured header from the incoming request (forwarded from proxy)
//  2. If absent, generates a UUID v4 via crypto/rand
//  3. Sets the ID on the response header
//  4. Stores the ID on the context via c.Set("request_id", id) and c.SetRequestID(id)
//
// Usage:
//
//	app.Use(middleware.RequestID())
//	app.Use(middleware.RequestID(middleware.RequestIDConfig{
//	    Header: "X-Trace-ID",
//	    Generator: func() string { return myCustomID() },
//	}))
func RequestID(config ...RequestIDConfig) func(*rudraContext.Context) error {
	cfg := DefaultRequestIDConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Generator == nil {
		cfg.Generator = generateUUIDv4
	}
	if cfg.Header == "" {
		cfg.Header = "X-Request-ID"
	}

	return func(c *rudraContext.Context) error {
		// Check for forwarded request ID from upstream proxy.
		id := c.Header(cfg.Header)

		// Generate a new ID if none was forwarded.
		if id == "" {
			id = cfg.Generator()
		}

		// Set on response header.
		c.SetHeader(cfg.Header, id)

		// Store on context for downstream access.
		c.Set("request_id", id)
		c.SetRequestID(id)

		return c.Next()
	}
}

// generateUUIDv4 generates a cryptographically random UUID v4 string.
// Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
// Uses crypto/rand — safe for production use.
func generateUUIDv4() string {
	var uuid [16]byte
	_, _ = rand.Read(uuid[:])

	// Set version 4 (bits 12-15 of time_hi_and_version).
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// Set variant bits (10xxxxxx).
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}
