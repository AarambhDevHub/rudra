package middleware

import (
	"net/http"
	"strings"
	"testing"
)

func TestETagDefaultConfig(t *testing.T) {
	cfg := DefaultETagConfig()
	if cfg.Weak {
		t.Error("expected Weak=false")
	}
}

func TestETagGeneratesHeader(t *testing.T) {
	mw := ETag()

	c, w := newTestContext("GET", "/data")
	c.SetNext(func() error {
		return c.String(200, "hello world")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	etag := w.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header")
	}
	if !strings.HasPrefix(etag, "\"") || !strings.HasSuffix(etag, "\"") {
		t.Errorf("expected quoted ETag, got '%s'", etag)
	}
}

func TestETagWeakFormat(t *testing.T) {
	mw := ETag(ETagConfig{Weak: true})

	c, w := newTestContext("GET", "/data")
	c.SetNext(func() error {
		return c.String(200, "hello world")
	})

	_ = mw(c)

	etag := w.Header().Get("ETag")
	if !strings.HasPrefix(etag, "W/\"") {
		t.Errorf("expected weak ETag starting with W/, got '%s'", etag)
	}
}

func TestETag304OnMatch(t *testing.T) {
	mw := ETag()

	// First request to get the ETag.
	c1, w1 := newTestContext("GET", "/data")
	c1.SetNext(func() error { return c1.String(200, "hello world") })
	_ = mw(c1)

	etag := w1.Header().Get("ETag")
	if etag == "" {
		t.Fatal("no ETag from first request")
	}

	// Second request with If-None-Match.
	c2, w2 := newTestContext("GET", "/data")
	c2.Request().Header.Set("If-None-Match", etag)
	c2.SetNext(func() error { return c2.String(200, "hello world") })
	_ = mw(c2)

	if w2.Code != http.StatusNotModified {
		t.Errorf("expected 304, got %d", w2.Code)
	}

	// Body should be empty on 304.
	if w2.Body.Len() > 0 {
		t.Error("expected empty body on 304")
	}
}

func TestETagDifferentContent(t *testing.T) {
	mw := ETag()

	// First request.
	c1, w1 := newTestContext("GET", "/data")
	c1.SetNext(func() error { return c1.String(200, "version 1") })
	_ = mw(c1)
	etag1 := w1.Header().Get("ETag")

	// Second request with different content.
	c2, w2 := newTestContext("GET", "/data")
	c2.SetNext(func() error { return c2.String(200, "version 2") })
	_ = mw(c2)
	etag2 := w2.Header().Get("ETag")

	if etag1 == etag2 {
		t.Error("different content should produce different ETags")
	}
}

func TestETagSkipsPOST(t *testing.T) {
	mw := ETag()

	c, w := newTestContext("POST", "/data")
	c.SetNext(func() error { return c.String(201, "created") })

	_ = mw(c)

	// POST responses should still have the body.
	if w.Body.String() == "" {
		t.Error("POST body should not be empty")
	}
}

func TestETagConsistentHash(t *testing.T) {
	mw := ETag()

	// Same content should produce the same ETag.
	var etags []string
	for i := 0; i < 3; i++ {
		c, w := newTestContext("GET", "/data")
		c.SetNext(func() error { return c.String(200, "consistent content") })
		_ = mw(c)
		etags = append(etags, w.Header().Get("ETag"))
	}

	for i := 1; i < len(etags); i++ {
		if etags[i] != etags[0] {
			t.Errorf("inconsistent ETags: %s vs %s", etags[0], etags[i])
		}
	}
}
