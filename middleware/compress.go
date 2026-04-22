package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

// CompressConfig defines the config for the Compress middleware.
type CompressConfig struct {
	// Level is the gzip compression level (1-9).
	// Use gzip.BestSpeed (1) for lowest latency, gzip.BestCompression (9) for smallest size.
	// Default: gzip.DefaultCompression (6).
	Level int

	// MinLength is the minimum response body size to trigger compression.
	// Responses smaller than this are sent uncompressed.
	// Default: 1024 bytes.
	MinLength int

	// ContentTypes is the list of content types eligible for compression.
	// If empty, defaults to common text-based types.
	// Matched by prefix: "text/" matches "text/html", "text/plain", etc.
	ContentTypes []string
}

// DefaultCompressConfig returns a CompressConfig with sane defaults.
func DefaultCompressConfig() CompressConfig {
	return CompressConfig{
		Level:     gzip.DefaultCompression,
		MinLength: 1024,
		ContentTypes: []string{
			"text/",
			"application/json",
			"application/xml",
			"application/javascript",
			"application/wasm",
			"image/svg+xml",
		},
	}
}

// Compress returns a gzip compression middleware.
//
// It negotiates compression via the Accept-Encoding header, wraps the
// response writer with a pooled gzip.Writer, and sets Content-Encoding
// and Vary headers. Responses below MinLength or with non-matching content
// types pass through uncompressed.
//
// Uses sync.Pool for gzip writers to minimize allocations.
//
// Usage:
//
//	app.Use(middleware.Compress())
//	app.Use(middleware.Compress(middleware.CompressConfig{
//	    Level:     gzip.BestSpeed,
//	    MinLength: 512,
//	}))
func Compress(config ...CompressConfig) func(*rudraContext.Context) error {
	cfg := DefaultCompressConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Level < gzip.HuffmanOnly || cfg.Level > gzip.BestCompression {
		cfg.Level = gzip.DefaultCompression
	}
	if cfg.MinLength <= 0 {
		cfg.MinLength = 1024
	}
	if len(cfg.ContentTypes) == 0 {
		cfg.ContentTypes = DefaultCompressConfig().ContentTypes
	}

	// Pool gzip writers to avoid per-request allocation.
	pool := &sync.Pool{
		New: func() any {
			w, _ := gzip.NewWriterLevel(io.Discard, cfg.Level)
			return w
		},
	}

	return func(c *rudraContext.Context) error {
		// Check if client accepts gzip.
		if !strings.Contains(c.Header("Accept-Encoding"), "gzip") {
			return c.Next()
		}

		// Wrap the response writer with the gzip writer.
		gw := pool.Get().(*gzip.Writer)
		origWriter := c.Writer()

		cw := &compressWriter{
			ResponseWriter: origWriter,
			gw:             gw,
			pool:           pool,
			minLength:      cfg.MinLength,
			contentTypes:   cfg.ContentTypes,
		}
		c.SetWriter(cw)

		// Set Vary header so caches know this response depends on Accept-Encoding.
		origWriter.Header().Set("Vary", "Accept-Encoding")

		err := c.Next()

		// Flush and return the gzip writer only if compression was activated.
		if cw.compressed {
			gw.Close()
			pool.Put(gw)
		}

		// Restore original writer.
		c.SetWriter(origWriter)
		return err
	}
}

// compressWriter wraps http.ResponseWriter with conditional gzip compression.
// It buffers the first write to check content type and length before deciding
// whether to compress.
type compressWriter struct {
	http.ResponseWriter
	gw           *gzip.Writer
	pool         *sync.Pool
	minLength    int
	contentTypes []string
	compressed   bool
	wroteHeader  bool
	statusCode   int
}

func (cw *compressWriter) WriteHeader(code int) {
	if cw.wroteHeader {
		return
	}
	cw.statusCode = code
	// Don't write header yet — we may need to add Content-Encoding first.
	// It will be written on the first Write call.
}

func (cw *compressWriter) Write(b []byte) (int, error) {
	if !cw.wroteHeader {
		cw.wroteHeader = true

		if cw.statusCode == 0 {
			cw.statusCode = http.StatusOK
		}

		// Check if we should compress.
		ct := cw.ResponseWriter.Header().Get("Content-Type")
		shouldCompress := len(b) >= cw.minLength && cw.matchContentType(ct)

		// Don't compress if Content-Encoding is already set (e.g., pre-compressed).
		if cw.ResponseWriter.Header().Get("Content-Encoding") != "" {
			shouldCompress = false
		}

		// Don't compress 204, 304, or 1xx responses.
		if cw.statusCode == http.StatusNoContent ||
			cw.statusCode == http.StatusNotModified ||
			cw.statusCode < 200 {
			shouldCompress = false
		}

		if shouldCompress {
			cw.compressed = true
			cw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
			cw.ResponseWriter.Header().Del("Content-Length")
			cw.ResponseWriter.WriteHeader(cw.statusCode)
			cw.gw.Reset(cw.ResponseWriter)
			return cw.gw.Write(b)
		}

		// No compression — write directly.
		cw.ResponseWriter.WriteHeader(cw.statusCode)
	}

	if cw.compressed {
		return cw.gw.Write(b)
	}
	return cw.ResponseWriter.Write(b)
}

func (cw *compressWriter) matchContentType(ct string) bool {
	if ct == "" {
		return false
	}
	for _, allowed := range cw.contentTypes {
		if strings.HasPrefix(ct, allowed) {
			return true
		}
	}
	return false
}

// Flush implements http.Flusher.
func (cw *compressWriter) Flush() {
	if cw.compressed {
		cw.gw.Flush()
	}
	if f, ok := cw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
