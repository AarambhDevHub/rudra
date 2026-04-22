package middleware

import (
	"net/http"
	"testing"
)

func TestCORSDefaultConfig(t *testing.T) {
	cfg := DefaultCORSConfig()
	if len(cfg.AllowOrigins) != 1 || cfg.AllowOrigins[0] != "*" {
		t.Errorf("expected AllowOrigins=[*], got %v", cfg.AllowOrigins)
	}
	if cfg.MaxAge != 86400 {
		t.Errorf("expected MaxAge=86400, got %d", cfg.MaxAge)
	}
}

func TestCORSSimpleRequest(t *testing.T) {
	mw := CORS()

	c, w := newTestContext("GET", "/api/data")
	c.Request().Header.Set("Origin", "https://example.com")
	c.SetNext(func() error {
		return c.String(200, "data")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	acao := w.Header().Get("Access-Control-Allow-Origin")
	if acao != "*" {
		t.Errorf("expected Access-Control-Allow-Origin=*, got '%s'", acao)
	}
}

func TestCORSPreflightRequest(t *testing.T) {
	mw := CORS()

	c, w := newTestContext("OPTIONS", "/api/data")
	c.Request().Header.Set("Origin", "https://example.com")
	c.Request().Header.Set("Access-Control-Request-Method", "POST")
	handlerCalled := false
	c.SetNext(func() error {
		handlerCalled = true
		return c.String(200, "should not reach")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
	if handlerCalled {
		t.Error("handler should NOT be called for preflight")
	}

	acam := w.Header().Get("Access-Control-Allow-Methods")
	if acam == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
	acah := w.Header().Get("Access-Control-Allow-Headers")
	if acah == "" {
		t.Error("expected Access-Control-Allow-Headers header")
	}
	maxAge := w.Header().Get("Access-Control-Max-Age")
	if maxAge != "86400" {
		t.Errorf("expected Max-Age=86400, got '%s'", maxAge)
	}
}

func TestCORSSpecificOrigins(t *testing.T) {
	mw := CORS(CORSConfig{
		AllowOrigins: []string{"https://app.example.com", "https://admin.example.com"},
	})

	// Allowed origin.
	c1, w1 := newTestContext("GET", "/api")
	c1.Request().Header.Set("Origin", "https://app.example.com")
	c1.SetNext(func() error { return c1.String(200, "ok") })
	_ = mw(c1)

	if w1.Header().Get("Access-Control-Allow-Origin") != "https://app.example.com" {
		t.Errorf("expected specific origin, got '%s'", w1.Header().Get("Access-Control-Allow-Origin"))
	}

	// Disallowed origin.
	c2, w2 := newTestContext("GET", "/api")
	c2.Request().Header.Set("Origin", "https://evil.com")
	c2.SetNext(func() error { return c2.String(200, "ok") })
	_ = mw(c2)

	if w2.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no CORS headers for disallowed origin, got '%s'", w2.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSWithCredentials(t *testing.T) {
	mw := CORS(CORSConfig{
		AllowOrigins:     []string{"https://app.example.com"},
		AllowCredentials: true,
	})

	c, w := newTestContext("GET", "/api")
	c.Request().Header.Set("Origin", "https://app.example.com")
	c.SetNext(func() error { return c.String(200, "ok") })

	_ = mw(c)

	cred := w.Header().Get("Access-Control-Allow-Credentials")
	if cred != "true" {
		t.Errorf("expected credentials=true, got '%s'", cred)
	}

	// Should echo specific origin, not *.
	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin == "*" {
		t.Error("wildcard origin not allowed with credentials")
	}
}

func TestCORSWildcardWithCredentials(t *testing.T) {
	mw := CORS(CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
	})

	c, w := newTestContext("GET", "/api")
	c.Request().Header.Set("Origin", "https://example.com")
	c.SetNext(func() error { return c.String(200, "ok") })

	_ = mw(c)

	// When credentials=true, origin must be echoed, not *.
	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin == "*" {
		t.Error("wildcard origin must not be used with credentials")
	}
	if origin != "https://example.com" {
		t.Errorf("expected echoed origin, got '%s'", origin)
	}
}

func TestCORSNoOriginHeader(t *testing.T) {
	mw := CORS()

	c, w := newTestContext("GET", "/api")
	// No Origin header — not a CORS request.
	c.SetNext(func() error { return c.String(200, "ok") })

	_ = mw(c)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("should not set CORS headers without Origin header")
	}
}

func TestCORSAllowOriginFunc(t *testing.T) {
	mw := CORS(CORSConfig{
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://dynamic.example.com"
		},
	})

	// Allowed via function.
	c1, w1 := newTestContext("GET", "/api")
	c1.Request().Header.Set("Origin", "https://dynamic.example.com")
	c1.SetNext(func() error { return c1.String(200, "ok") })
	_ = mw(c1)

	if w1.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected CORS headers for dynamically allowed origin")
	}

	// Denied via function.
	c2, w2 := newTestContext("GET", "/api")
	c2.Request().Header.Set("Origin", "https://other.com")
	c2.SetNext(func() error { return c2.String(200, "ok") })
	_ = mw(c2)

	if w2.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no CORS headers for denied origin")
	}
}

func TestCORSExposeHeaders(t *testing.T) {
	mw := CORS(CORSConfig{
		AllowOrigins:  []string{"*"},
		ExposeHeaders: []string{"X-Custom-Header", "X-Another"},
	})

	c, w := newTestContext("GET", "/api")
	c.Request().Header.Set("Origin", "https://example.com")
	c.SetNext(func() error { return c.String(200, "ok") })

	_ = mw(c)

	expose := w.Header().Get("Access-Control-Expose-Headers")
	if expose != "X-Custom-Header, X-Another" {
		t.Errorf("expected expose headers, got '%s'", expose)
	}
}

func TestCORSVaryHeader(t *testing.T) {
	mw := CORS(CORSConfig{
		AllowOrigins: []string{"https://example.com"},
	})

	c, w := newTestContext("GET", "/api")
	c.Request().Header.Set("Origin", "https://example.com")
	c.SetNext(func() error { return c.String(200, "ok") })

	_ = mw(c)

	vary := w.Header().Get("Vary")
	if vary != "Origin" {
		t.Errorf("expected Vary=Origin for specific origin, got '%s'", vary)
	}
}
