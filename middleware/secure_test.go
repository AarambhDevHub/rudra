package middleware

import (
	"testing"
)

func TestSecureDefaultConfig(t *testing.T) {
	cfg := DefaultSecureConfig()
	if cfg.XFrameOptions != "DENY" {
		t.Errorf("expected XFrameOptions=DENY, got '%s'", cfg.XFrameOptions)
	}
	if !cfg.ContentTypeNosniff {
		t.Error("expected ContentTypeNosniff=true")
	}
	if cfg.HSTSMaxAge != 31536000 {
		t.Errorf("expected HSTSMaxAge=31536000, got %d", cfg.HSTSMaxAge)
	}
}

func TestSecureDefaultHeaders(t *testing.T) {
	mw := Secure()

	c, w := newTestContext("GET", "/")
	c.SetNext(func() error {
		return c.String(200, "ok")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("expected X-Content-Type-Options: nosniff")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("expected X-Frame-Options: DENY")
	}
	if w.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Error("expected X-XSS-Protection: 1; mode=block")
	}
	if w.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Error("expected Referrer-Policy header")
	}
}

func TestSecureCSP(t *testing.T) {
	mw := Secure(SecureConfig{
		ContentSecurityPolicy: "default-src 'self'",
	})

	c, w := newTestContext("GET", "/")
	c.SetNext(func() error { return c.String(200, "ok") })

	_ = mw(c)

	if w.Header().Get("Content-Security-Policy") != "default-src 'self'" {
		t.Errorf("expected CSP header, got '%s'", w.Header().Get("Content-Security-Policy"))
	}
}

func TestSecureHSTSOnlyOnHTTPS(t *testing.T) {
	mw := Secure()

	// Plain HTTP — should NOT set HSTS.
	c, w := newTestContext("GET", "/")
	c.SetNext(func() error { return c.String(200, "ok") })
	_ = mw(c)

	if w.Header().Get("Strict-Transport-Security") != "" {
		t.Error("HSTS should not be set on plain HTTP")
	}
}

func TestSecureHSTSWithProxy(t *testing.T) {
	mw := Secure()

	c, w := newTestContext("GET", "/")
	c.Request().Header.Set("X-Forwarded-Proto", "https")
	c.SetNext(func() error { return c.String(200, "ok") })

	_ = mw(c)

	hsts := w.Header().Get("Strict-Transport-Security")
	if hsts == "" {
		t.Error("expected HSTS header behind HTTPS proxy")
	}
	if hsts != "max-age=31536000; includeSubDomains" {
		t.Errorf("unexpected HSTS value: %s", hsts)
	}
}

func TestSecurePermissionsPolicy(t *testing.T) {
	mw := Secure(SecureConfig{
		PermissionsPolicy: "camera=(), microphone=()",
	})

	c, w := newTestContext("GET", "/")
	c.SetNext(func() error { return c.String(200, "ok") })
	_ = mw(c)

	if w.Header().Get("Permissions-Policy") != "camera=(), microphone=()" {
		t.Error("expected Permissions-Policy header")
	}
}

func TestSecureCustomFrameOptions(t *testing.T) {
	mw := Secure(SecureConfig{
		XFrameOptions: "SAMEORIGIN",
	})

	c, w := newTestContext("GET", "/")
	c.SetNext(func() error { return c.String(200, "ok") })
	_ = mw(c)

	if w.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Errorf("expected SAMEORIGIN, got '%s'", w.Header().Get("X-Frame-Options"))
	}
}
