package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"

	rudraContext "github.com/AarambhDevHub/rudra/context"
	"github.com/AarambhDevHub/rudra/errors"
)

// CSRFConfig defines the config for the CSRF middleware.
type CSRFConfig struct {
	// TokenLength is the length of the CSRF token in bytes (hex-encoded = 2x).
	// Default: 32.
	TokenLength int

	// CookieName is the name of the cookie that stores the CSRF token.
	// Default: "_csrf".
	CookieName string

	// HeaderName is the HTTP header where the client sends the token.
	// Default: "X-CSRF-Token".
	HeaderName string

	// FormField is the form field name for the CSRF token.
	// Default: "_csrf".
	FormField string

	// CookiePath is the path scope of the CSRF cookie.
	// Default: "/".
	CookiePath string

	// Secure marks the cookie as secure (HTTPS only).
	// Default: false.
	Secure bool

	// HttpOnly marks the cookie as HttpOnly.
	// Default: true.
	HttpOnly bool

	// SameSite sets the SameSite attribute of the cookie.
	// Default: http.SameSiteStrictMode.
	SameSite http.SameSite

	// MaxAge is the cookie max age in seconds.
	// Default: 86400 (24 hours).
	MaxAge int
}

// DefaultCSRFConfig returns a CSRFConfig with sane defaults.
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength: 32,
		CookieName:  "_csrf",
		HeaderName:  "X-CSRF-Token",
		FormField:   "_csrf",
		CookiePath:  "/",
		HttpOnly:    true,
		SameSite:    http.SameSiteStrictMode,
		MaxAge:      86400,
	}
}

// CSRF returns a Cross-Site Request Forgery protection middleware.
//
// It uses the double-submit cookie pattern:
//  1. On safe methods (GET, HEAD, OPTIONS), a CSRF token is generated and
//     set in a cookie + stored on the context for template rendering.
//  2. On unsafe methods (POST, PUT, PATCH, DELETE), the token from the
//     cookie is compared against the token in the header or form field.
//     A mismatch results in 403 Forbidden.
//
// Usage:
//
//	app.Use(middleware.CSRF())
//	app.Use(middleware.CSRF(middleware.CSRFConfig{
//	    Secure:   true,
//	    SameSite: http.SameSiteLaxMode,
//	}))
func CSRF(config ...CSRFConfig) func(*rudraContext.Context) error {
	cfg := DefaultCSRFConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.TokenLength <= 0 {
		cfg.TokenLength = 32
	}
	if cfg.CookieName == "" {
		cfg.CookieName = "_csrf"
	}
	if cfg.HeaderName == "" {
		cfg.HeaderName = "X-CSRF-Token"
	}
	if cfg.FormField == "" {
		cfg.FormField = "_csrf"
	}
	if cfg.CookiePath == "" {
		cfg.CookiePath = "/"
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = 86400
	}

	return func(c *rudraContext.Context) error {
		method := c.Method()

		// Safe methods: generate/refresh token.
		if method == http.MethodGet || method == http.MethodHead ||
			method == http.MethodOptions || method == http.MethodTrace {

			token := getCSRFCookie(c, cfg.CookieName)
			if token == "" {
				token = generateCSRFToken(cfg.TokenLength)
			}

			// Set/refresh the cookie.
			setCSRFCookie(c, cfg, token)

			// Store on context so templates can access it.
			c.Set("csrf_token", token)

			return c.Next()
		}

		// Unsafe methods: validate token.
		cookieToken := getCSRFCookie(c, cfg.CookieName)
		if cookieToken == "" {
			return errors.NewHTTPError(http.StatusForbidden, "csrf token missing from cookie")
		}

		// Check header first, then form field.
		clientToken := c.Header(cfg.HeaderName)
		if clientToken == "" {
			clientToken = c.Request().FormValue(cfg.FormField)
		}

		if clientToken == "" {
			return errors.NewHTTPError(http.StatusForbidden, "csrf token missing from request")
		}

		// Constant-time comparison to prevent timing attacks.
		if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(clientToken)) != 1 {
			return errors.NewHTTPError(http.StatusForbidden, "csrf token mismatch")
		}

		// Refresh the token on successful validation.
		newToken := generateCSRFToken(cfg.TokenLength)
		setCSRFCookie(c, cfg, newToken)
		c.Set("csrf_token", newToken)

		return c.Next()
	}
}

func generateCSRFToken(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func getCSRFCookie(c *rudraContext.Context, name string) string {
	cookie, err := c.Request().Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func setCSRFCookie(c *rudraContext.Context, cfg CSRFConfig, token string) {
	http.SetCookie(c.Writer(), &http.Cookie{
		Name:     cfg.CookieName,
		Value:    token,
		Path:     cfg.CookiePath,
		MaxAge:   cfg.MaxAge,
		Secure:   cfg.Secure,
		HttpOnly: cfg.HttpOnly,
		SameSite: cfg.SameSite,
	})
}
