package middleware

import (
	"net/http"
	"testing"
)

func TestCSRFDefaultConfig(t *testing.T) {
	cfg := DefaultCSRFConfig()
	if cfg.CookieName != "_csrf" {
		t.Errorf("expected CookieName '_csrf', got '%s'", cfg.CookieName)
	}
	if cfg.HeaderName != "X-CSRF-Token" {
		t.Errorf("expected HeaderName 'X-CSRF-Token', got '%s'", cfg.HeaderName)
	}
}

func TestCSRFSetsTokenOnGET(t *testing.T) {
	mw := CSRF()

	c, w := newTestContext("GET", "/form")
	c.SetNext(func() error { return c.String(200, "form page") })

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check cookie was set.
	cookies := w.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "_csrf" {
			found = true
			if cookie.Value == "" {
				t.Error("CSRF cookie value is empty")
			}
		}
	}
	if !found {
		t.Error("expected _csrf cookie to be set")
	}

	// Check context has token.
	token, ok := c.Get("csrf_token")
	if !ok || token.(string) == "" {
		t.Error("expected csrf_token in context store")
	}
}

func TestCSRFBlocksPOSTWithoutToken(t *testing.T) {
	mw := CSRF()

	c, _ := newTestContext("POST", "/submit")
	c.SetNext(func() error { return c.String(200, "ok") })

	err := mw(c)
	if err == nil {
		t.Fatal("expected error for POST without CSRF token")
	}
}

func TestCSRFAllowsPOSTWithValidToken(t *testing.T) {
	mw := CSRF()

	// Step 1: GET to obtain the token.
	c1, w1 := newTestContext("GET", "/form")
	c1.SetNext(func() error { return c1.String(200, "form") })
	_ = mw(c1)

	// Extract token from cookie.
	var csrfToken string
	for _, cookie := range w1.Result().Cookies() {
		if cookie.Name == "_csrf" {
			csrfToken = cookie.Value
			break
		}
	}

	if csrfToken == "" {
		t.Fatal("no CSRF token from GET")
	}

	// Step 2: POST with the token.
	c2, _ := newTestContext("POST", "/submit")
	c2.Request().Header.Set("X-CSRF-Token", csrfToken)
	c2.Request().AddCookie(&http.Cookie{Name: "_csrf", Value: csrfToken})
	c2.SetNext(func() error { return c2.String(200, "ok") })

	err := mw(c2)
	if err != nil {
		t.Fatalf("expected valid POST to succeed, got: %v", err)
	}
}

func TestCSRFRejectsMismatchedToken(t *testing.T) {
	mw := CSRF()

	c, _ := newTestContext("POST", "/submit")
	c.Request().AddCookie(&http.Cookie{Name: "_csrf", Value: "cookie-token-abc"})
	c.Request().Header.Set("X-CSRF-Token", "wrong-token-xyz")
	c.SetNext(func() error { return c.String(200, "ok") })

	err := mw(c)
	if err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestCSRFSkipsSafeMethods(t *testing.T) {
	mw := CSRF()

	for _, method := range []string{"GET", "HEAD", "OPTIONS"} {
		c, _ := newTestContext(method, "/")
		c.SetNext(func() error { return c.String(200, "ok") })

		err := mw(c)
		if err != nil {
			t.Errorf("expected %s to be allowed, got: %v", method, err)
		}
	}
}
