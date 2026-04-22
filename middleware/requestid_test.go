package middleware

import (
	"strings"
	"testing"
)

func TestRequestIDDefaultConfig(t *testing.T) {
	cfg := DefaultRequestIDConfig()
	if cfg.Header != "X-Request-ID" {
		t.Errorf("expected header 'X-Request-ID', got '%s'", cfg.Header)
	}
	if cfg.Generator == nil {
		t.Error("expected non-nil generator")
	}
}

func TestRequestIDGeneratesUUID(t *testing.T) {
	mw := RequestID()

	c, w := newTestContext("GET", "/")
	c.SetNext(func() error {
		return c.String(200, "ok")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check response header.
	id := w.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected X-Request-ID header, got empty")
	}

	// UUID v4 format: 8-4-4-4-12 = 36 characters.
	if len(id) != 36 {
		t.Errorf("expected UUID length 36, got %d: %s", len(id), id)
	}

	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Errorf("expected 5 UUID parts, got %d: %s", len(parts), id)
	}

	// Check context storage.
	val, ok := c.Get("request_id")
	if !ok {
		t.Fatal("expected request_id in context store")
	}
	if val.(string) != id {
		t.Errorf("context request_id doesn't match header: %s vs %s", val, id)
	}

	// Check convenience accessor.
	if c.RequestID() != id {
		t.Errorf("RequestID() doesn't match: %s vs %s", c.RequestID(), id)
	}
}

func TestRequestIDForwardsExisting(t *testing.T) {
	mw := RequestID()

	c, w := newTestContext("GET", "/")
	c.Request().Header.Set("X-Request-ID", "forwarded-123-abc")
	c.SetNext(func() error {
		return c.String(200, "ok")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Forwarded ID should be preserved.
	id := w.Header().Get("X-Request-ID")
	if id != "forwarded-123-abc" {
		t.Errorf("expected forwarded ID 'forwarded-123-abc', got '%s'", id)
	}

	if c.RequestID() != "forwarded-123-abc" {
		t.Errorf("expected forwarded ID in context, got '%s'", c.RequestID())
	}
}

func TestRequestIDCustomGenerator(t *testing.T) {
	counter := 0
	mw := RequestID(RequestIDConfig{
		Generator: func() string {
			counter++
			return "custom-" + strings.Repeat("x", counter)
		},
	})

	c, w := newTestContext("GET", "/")
	c.SetNext(func() error {
		return c.String(200, "ok")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	id := w.Header().Get("X-Request-ID")
	if id != "custom-x" {
		t.Errorf("expected custom ID 'custom-x', got '%s'", id)
	}
}

func TestRequestIDCustomHeader(t *testing.T) {
	mw := RequestID(RequestIDConfig{
		Header: "X-Trace-ID",
	})

	c, w := newTestContext("GET", "/")
	c.SetNext(func() error {
		return c.String(200, "ok")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	id := w.Header().Get("X-Trace-ID")
	if id == "" {
		t.Fatal("expected X-Trace-ID header, got empty")
	}

	// Default header should NOT be set.
	defaultID := w.Header().Get("X-Request-ID")
	if defaultID != "" {
		t.Errorf("expected empty X-Request-ID with custom header, got '%s'", defaultID)
	}
}

func TestRequestIDUniquePerRequest(t *testing.T) {
	mw := RequestID()
	ids := make(map[string]struct{}, 100)

	for i := 0; i < 100; i++ {
		c, w := newTestContext("GET", "/")
		c.SetNext(func() error {
			return c.String(200, "ok")
		})

		_ = mw(c)
		id := w.Header().Get("X-Request-ID")
		if _, dup := ids[id]; dup {
			t.Fatalf("duplicate request ID generated: %s", id)
		}
		ids[id] = struct{}{}
	}
}

func TestGenerateUUIDv4Format(t *testing.T) {
	for i := 0; i < 100; i++ {
		uuid := generateUUIDv4()

		if len(uuid) != 36 {
			t.Fatalf("expected UUID length 36, got %d: %s", len(uuid), uuid)
		}

		// Check version 4 marker: 13th character should be '4'.
		if uuid[14] != '4' {
			t.Errorf("expected version '4' at position 14, got '%c' in %s", uuid[14], uuid)
		}

		// Check variant bits: 19th character should be one of 8, 9, a, b.
		variant := uuid[19]
		if variant != '8' && variant != '9' && variant != 'a' && variant != 'b' {
			t.Errorf("expected variant [89ab] at position 19, got '%c' in %s", variant, uuid)
		}
	}
}
