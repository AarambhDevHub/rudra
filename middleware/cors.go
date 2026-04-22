package middleware

import (
	"net/http"
	"strconv"
	"strings"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

// CORSConfig defines the config for the CORS middleware.
type CORSConfig struct {
	// AllowOrigins is a list of origins that are allowed to make cross-domain requests.
	// Use "*" to allow all origins (not safe with AllowCredentials=true).
	// Default: ["*"]
	AllowOrigins []string

	// AllowMethods specifies the methods allowed for cross-domain requests.
	// Default: GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD
	AllowMethods []string

	// AllowHeaders specifies the request headers allowed for cross-domain requests.
	// Default: Origin, Content-Type, Accept, Authorization
	AllowHeaders []string

	// ExposeHeaders specifies which response headers are safe to expose to the browser.
	// Default: [] (none)
	ExposeHeaders []string

	// AllowCredentials indicates whether the response to the request can include
	// credentials (cookies, authorization headers, TLS client certificates).
	// When true, AllowOrigins must NOT be ["*"] — per CORS spec.
	// Default: false
	AllowCredentials bool

	// MaxAge specifies how long (in seconds) the results of a preflight request
	// can be cached by the browser.
	// Default: 86400 (24 hours)
	MaxAge int

	// AllowOriginFunc is an optional dynamic origin validator.
	// If set, it is called for every request and takes precedence over AllowOrigins.
	// Return true to allow the origin.
	AllowOriginFunc func(origin string) bool

	// Pre-computed header values (set during init).
	allowMethodsStr  string
	allowHeadersStr  string
	exposeHeadersStr string
	maxAgeStr        string
	allowOriginsMap  map[string]struct{}
	allowAll         bool
}

// DefaultCORSConfig returns a permissive CORS config suitable for development.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// CORS returns a Cross-Origin Resource Sharing middleware.
//
// It handles both simple requests and preflight OPTIONS requests per the CORS spec.
// Preflight responses are returned with 204 No Content and do not proceed to the handler.
//
// The middleware enforces the CORS security rule: when AllowCredentials is true,
// the wildcard "*" origin is NOT allowed — the response must echo the specific origin.
//
// Usage:
//
//	app.Use(middleware.CORS())
//	app.Use(middleware.CORS(middleware.CORSConfig{
//	    AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
//	    AllowCredentials: true,
//	    MaxAge:           3600,
//	}))
func CORS(config ...CORSConfig) func(*rudraContext.Context) error {
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Apply defaults for zero-value slices.
	if len(cfg.AllowOrigins) == 0 {
		cfg.AllowOrigins = []string{"*"}
	}
	if len(cfg.AllowMethods) == 0 {
		cfg.AllowMethods = DefaultCORSConfig().AllowMethods
	}
	if len(cfg.AllowHeaders) == 0 {
		cfg.AllowHeaders = DefaultCORSConfig().AllowHeaders
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = 86400
	}

	// Pre-compute header strings to avoid per-request string joins.
	cfg.allowMethodsStr = strings.Join(cfg.AllowMethods, ", ")
	cfg.allowHeadersStr = strings.Join(cfg.AllowHeaders, ", ")
	if len(cfg.ExposeHeaders) > 0 {
		cfg.exposeHeadersStr = strings.Join(cfg.ExposeHeaders, ", ")
	}
	cfg.maxAgeStr = strconv.Itoa(cfg.MaxAge)

	// Build origins lookup map.
	cfg.allowOriginsMap = make(map[string]struct{}, len(cfg.AllowOrigins))
	for _, o := range cfg.AllowOrigins {
		if o == "*" {
			cfg.allowAll = true
		}
		cfg.allowOriginsMap[o] = struct{}{}
	}

	return func(c *rudraContext.Context) error {
		origin := c.Header("Origin")

		// No Origin header — not a CORS request, skip entirely.
		if origin == "" {
			return c.Next()
		}

		// Check if the origin is allowed.
		allowed := false
		if cfg.AllowOriginFunc != nil {
			allowed = cfg.AllowOriginFunc(origin)
		} else if cfg.allowAll {
			allowed = true
		} else {
			_, allowed = cfg.allowOriginsMap[origin]
		}

		if !allowed {
			// Origin not allowed — do not set CORS headers.
			// For preflight, return 204 without CORS headers (browser will block).
			if c.Method() == http.MethodOptions {
				c.SetStatus(http.StatusNoContent)
				return nil
			}
			return c.Next()
		}

		// Determine the Access-Control-Allow-Origin value.
		allowOriginValue := origin
		if cfg.allowAll && !cfg.AllowCredentials {
			// Wildcard is safe when credentials are not used.
			allowOriginValue = "*"
		}

		// Set CORS headers (both simple and preflight).
		c.SetHeader("Access-Control-Allow-Origin", allowOriginValue)
		if cfg.AllowCredentials {
			c.SetHeader("Access-Control-Allow-Credentials", "true")
		}
		if cfg.exposeHeadersStr != "" {
			c.SetHeader("Access-Control-Expose-Headers", cfg.exposeHeadersStr)
		}

		// Vary header — required when origin is not "*".
		if allowOriginValue != "*" {
			c.SetHeader("Vary", "Origin")
		}

		// Handle preflight OPTIONS request.
		if c.Method() == http.MethodOptions {
			reqMethod := c.Header("Access-Control-Request-Method")
			if reqMethod != "" {
				// This is a preflight request.
				c.SetHeader("Access-Control-Allow-Methods", cfg.allowMethodsStr)
				c.SetHeader("Access-Control-Allow-Headers", cfg.allowHeadersStr)
				c.SetHeader("Access-Control-Max-Age", cfg.maxAgeStr)

				// Return 204 No Content — do not proceed to handler.
				c.SetStatus(http.StatusNoContent)
				return nil
			}
		}

		// Simple request or actual request after preflight — proceed to handler.
		return c.Next()
	}
}
