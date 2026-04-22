package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

// ETagConfig defines the config for the ETag middleware.
type ETagConfig struct {
	// Weak generates weak ETags (W/"...") instead of strong ETags.
	// Default: false.
	Weak bool
}

// DefaultETagConfig returns an ETagConfig with sane defaults.
func DefaultETagConfig() ETagConfig {
	return ETagConfig{
		Weak: false,
	}
}

// ETag returns a middleware that generates ETag headers from response body hashes.
//
// It computes a SHA-256 hash of the response body and sets the ETag header.
// If the client sends an If-None-Match header that matches the ETag,
// the middleware returns 304 Not Modified with no body.
//
// This middleware should be placed BEFORE compression middleware so the ETag
// is computed on the uncompressed body.
//
// Usage:
//
//	app.Use(middleware.ETag())
//	app.Use(middleware.ETag(middleware.ETagConfig{Weak: true}))
func ETag(config ...ETagConfig) func(*rudraContext.Context) error {
	cfg := DefaultETagConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *rudraContext.Context) error {
		// Wrap the writer to capture the response body for hashing.
		origWriter := c.Writer()
		ew := &etagWriter{
			ResponseWriter: origWriter,
		}
		c.SetWriter(ew)

		err := c.Next()

		// Restore original writer.
		c.SetWriter(origWriter)

		// Only generate ETag for successful GET/HEAD responses.
		if c.Method() != http.MethodGet && c.Method() != http.MethodHead {
			// For non-GET/HEAD, flush the buffered body directly.
			if ew.statusCode > 0 {
				origWriter.WriteHeader(ew.statusCode)
			}
			if len(ew.body) > 0 {
				origWriter.Write(ew.body)
			}
			return err
		}

		// Skip ETag for non-2xx responses or empty bodies.
		if ew.statusCode >= 300 || ew.statusCode < 200 || len(ew.body) == 0 {
			if ew.statusCode > 0 {
				origWriter.WriteHeader(ew.statusCode)
			}
			if len(ew.body) > 0 {
				origWriter.Write(ew.body)
			}
			return err
		}

		// Compute ETag from body hash.
		hash := sha256.Sum256(ew.body)
		var etag string
		if cfg.Weak {
			etag = fmt.Sprintf("W/\"%x\"", hash[:16])
		} else {
			etag = fmt.Sprintf("\"%x\"", hash[:16])
		}

		// Check If-None-Match.
		ifNoneMatch := c.Header("If-None-Match")
		if ifNoneMatch != "" && ifNoneMatch == etag {
			origWriter.Header().Set("ETag", etag)
			origWriter.WriteHeader(http.StatusNotModified)
			return nil
		}

		// Set ETag header and write the response.
		origWriter.Header().Set("ETag", etag)
		if ew.statusCode > 0 {
			origWriter.WriteHeader(ew.statusCode)
		}
		origWriter.Write(ew.body)
		return err
	}
}

// etagWriter buffers the response body so we can compute the hash before sending.
type etagWriter struct {
	http.ResponseWriter
	body        []byte
	statusCode  int
	wroteHeader bool
}

func (ew *etagWriter) WriteHeader(code int) {
	if !ew.wroteHeader {
		ew.statusCode = code
		ew.wroteHeader = true
	}
	// Don't forward — we buffer until we compute the ETag.
}

func (ew *etagWriter) Write(b []byte) (int, error) {
	if !ew.wroteHeader {
		ew.statusCode = http.StatusOK
		ew.wroteHeader = true
	}
	ew.body = append(ew.body, b...)
	return len(b), nil
}
