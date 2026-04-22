package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

// SecureConfig defines the config for the Secure middleware.
type SecureConfig struct {
	// XSSProtection sets the X-XSS-Protection header.
	// Default: "1; mode=block".
	// Set to "" to disable.
	XSSProtection string

	// ContentTypeNosniff sets X-Content-Type-Options: nosniff when true.
	// Default: true.
	ContentTypeNosniff bool

	// XFrameOptions sets the X-Frame-Options header.
	// Values: "DENY", "SAMEORIGIN", "ALLOW-FROM uri".
	// Default: "DENY".
	XFrameOptions string

	// HSTSMaxAge sets the max-age for Strict-Transport-Security in seconds.
	// Default: 31536000 (1 year).
	// Set to 0 to disable HSTS.
	HSTSMaxAge int

	// HSTSIncludeSubdomains adds includeSubDomains to the HSTS header.
	// Default: true.
	HSTSIncludeSubdomains bool

	// HSTSPreload adds the preload directive to the HSTS header.
	// Default: false.
	HSTSPreload bool

	// ContentSecurityPolicy sets the Content-Security-Policy header.
	// Default: "" (not set).
	ContentSecurityPolicy string

	// ReferrerPolicy sets the Referrer-Policy header.
	// Default: "strict-origin-when-cross-origin".
	ReferrerPolicy string

	// PermissionsPolicy sets the Permissions-Policy header.
	// Default: "" (not set).
	PermissionsPolicy string

	// Pre-computed HSTS header value (set during init).
	hstsValue string
}

// DefaultSecureConfig returns a SecureConfig with production-safe defaults.
func DefaultSecureConfig() SecureConfig {
	return SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    true,
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}
}

// Secure returns a middleware that sets security-related HTTP response headers.
//
// Headers set:
//   - X-XSS-Protection
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options
//   - Strict-Transport-Security (HSTS)
//   - Content-Security-Policy
//   - Referrer-Policy
//   - Permissions-Policy
//
// Usage:
//
//	app.Use(middleware.Secure())
//	app.Use(middleware.Secure(middleware.SecureConfig{
//	    ContentSecurityPolicy: "default-src 'self'",
//	    HSTSPreload:           true,
//	}))
func Secure(config ...SecureConfig) func(*rudraContext.Context) error {
	cfg := DefaultSecureConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Pre-compute the HSTS header value to avoid per-request string formatting.
	if cfg.HSTSMaxAge > 0 {
		cfg.hstsValue = fmt.Sprintf("max-age=%s", strconv.Itoa(cfg.HSTSMaxAge))
		if cfg.HSTSIncludeSubdomains {
			cfg.hstsValue += "; includeSubDomains"
		}
		if cfg.HSTSPreload {
			cfg.hstsValue += "; preload"
		}
	}

	return func(c *rudraContext.Context) error {
		h := c.Writer().Header()

		// X-XSS-Protection
		if cfg.XSSProtection != "" {
			h.Set("X-XSS-Protection", cfg.XSSProtection)
		}

		// X-Content-Type-Options
		if cfg.ContentTypeNosniff {
			h.Set("X-Content-Type-Options", "nosniff")
		}

		// X-Frame-Options
		if cfg.XFrameOptions != "" {
			h.Set("X-Frame-Options", cfg.XFrameOptions)
		}

		// Strict-Transport-Security
		if cfg.hstsValue != "" {
			// Only set HSTS on HTTPS requests.
			if c.Request().TLS != nil || c.Header("X-Forwarded-Proto") == "https" {
				h.Set("Strict-Transport-Security", cfg.hstsValue)
			}
		}

		// Content-Security-Policy
		if cfg.ContentSecurityPolicy != "" {
			h.Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
		}

		// Referrer-Policy
		if cfg.ReferrerPolicy != "" {
			h.Set("Referrer-Policy", cfg.ReferrerPolicy)
		}

		// Permissions-Policy
		if cfg.PermissionsPolicy != "" {
			h.Set("Permissions-Policy", cfg.PermissionsPolicy)
		}

		return c.Next()
	}
}

// SecureRedirect returns a middleware that redirects HTTP requests to HTTPS.
//
// Usage:
//
//	app.Use(middleware.SecureRedirect(443))
func SecureRedirect(httpsPort int) func(*rudraContext.Context) error {
	portSuffix := ""
	if httpsPort != 443 {
		portSuffix = ":" + strconv.Itoa(httpsPort)
	}

	return func(c *rudraContext.Context) error {
		if c.Request().TLS == nil && c.Header("X-Forwarded-Proto") != "https" {
			target := "https://" + c.Request().Host + portSuffix + c.Request().URL.RequestURI()
			return c.Redirect(http.StatusMovedPermanently, target)
		}
		return c.Next()
	}
}
