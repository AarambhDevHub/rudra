package middleware

import (
	"compress/gzip"
	"io"
	"strings"
	"testing"
)

func TestCompressDefaultConfig(t *testing.T) {
	cfg := DefaultCompressConfig()
	if cfg.Level != gzip.DefaultCompression {
		t.Errorf("expected default level, got %d", cfg.Level)
	}
	if cfg.MinLength != 1024 {
		t.Errorf("expected MinLength 1024, got %d", cfg.MinLength)
	}
}

func TestCompressGzipResponse(t *testing.T) {
	mw := Compress(CompressConfig{
		Level:     gzip.BestSpeed,
		MinLength: 10, // Low threshold for testing.
	})

	c, w := newTestContext("GET", "/data")
	c.Request().Header.Set("Accept-Encoding", "gzip")
	largeBody := strings.Repeat("hello world ", 100) // >10 bytes
	c.SetNext(func() error {
		return c.JSON(200, map[string]string{"data": largeBody})
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check Content-Encoding header.
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding=gzip, got '%s'", w.Header().Get("Content-Encoding"))
	}

	// Verify Vary header.
	if w.Header().Get("Vary") != "Accept-Encoding" {
		t.Errorf("expected Vary=Accept-Encoding, got '%s'", w.Header().Get("Vary"))
	}

	// Verify the body is valid gzip.
	reader, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}
	if len(decompressed) == 0 {
		t.Error("decompressed body is empty")
	}
}

func TestCompressSkipsSmallResponse(t *testing.T) {
	mw := Compress(CompressConfig{
		MinLength: 10000, // High threshold.
	})

	c, w := newTestContext("GET", "/small")
	c.Request().Header.Set("Accept-Encoding", "gzip")
	c.SetNext(func() error {
		return c.String(200, "tiny")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Small response should NOT be compressed.
	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Error("small response should not be compressed")
	}
}

func TestCompressSkipsWithoutAcceptEncoding(t *testing.T) {
	mw := Compress()

	c, w := newTestContext("GET", "/data")
	// No Accept-Encoding header.
	c.SetNext(func() error {
		return c.String(200, strings.Repeat("x", 2000))
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Error("should not compress without Accept-Encoding")
	}
}

func TestCompressVaryHeaderAlwaysSet(t *testing.T) {
	mw := Compress()

	c, w := newTestContext("GET", "/data")
	c.Request().Header.Set("Accept-Encoding", "gzip")
	c.SetNext(func() error {
		return c.String(200, "small")
	})

	_ = mw(c)

	if w.Header().Get("Vary") != "Accept-Encoding" {
		t.Error("Vary header should always be set when client accepts gzip")
	}
}
