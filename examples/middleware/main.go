// Package main demonstrates all v0.1.x middleware composed together.
//
// This example shows a production-ready Rudra server with:
//   - Recovery  — catches panics, returns 500 JSON, logs stack trace
//   - RequestID — generates UUID v4 per request, sets X-Request-ID header
//   - Logger    — structured JSON access logs with latency + bytes tracking
//   - Timeout   — 10s per-request deadline, propagated to downstream contexts
//   - CORS      — full CORS with preflight, specific origins, credentials
//
// Run:
//
//	go run ./examples/middleware
//
// Test:
//
//	curl -v http://localhost:8080/
//	curl -v http://localhost:8080/hello/rudra
//	curl -v http://localhost:8080/panic        # triggers recovery
//	curl -v http://localhost:8080/slow         # triggers timeout (if wait > 10s)
//	curl -v -X OPTIONS -H "Origin: https://app.example.com" \
//	  -H "Access-Control-Request-Method: POST" http://localhost:8080/api/data
package main

import (
	"log"
	"net/http"
	"time"

	rudraContext "github.com/AarambhDevHub/rudra/context"
	"github.com/AarambhDevHub/rudra/core"
	"github.com/AarambhDevHub/rudra/middleware"
)

func main() {
	app := core.New()

	// ── Global Middleware Stack (onion model — outermost first) ────────

	// 1. Recovery: outermost — catches panics from ALL layers.
	app.Use(middleware.Recovery())

	// 2. RequestID: generates/forwards X-Request-ID for every request.
	app.Use(middleware.RequestID())

	// 3. Logger: structured JSON access log with latency, bytes, request_id.
	//    Skips /health to reduce noise.
	app.Use(middleware.Logger(middleware.LoggerConfig{
		Format:    "json",
		SkipPaths: []string{"/health"},
	}))

	// 4. Timeout: 10s deadline per request — propagated to r.Context().
	app.Use(middleware.Timeout(middleware.TimeoutConfig{
		Timeout: 10 * time.Second,
	}))

	// 5. CORS: allow specific origins with credentials.
	app.Use(middleware.CORS(middleware.CORSConfig{
		AllowOrigins:     []string{"https://app.example.com", "https://admin.example.com"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	// ── Routes ────────────────────────────────────────────────────────

	app.GET("/", func(c *rudraContext.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"framework":  "Rudra",
			"version":    "0.1.5",
			"request_id": c.RequestID(),
		})
	})

	app.GET("/hello/:name", func(c *rudraContext.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message":    "Hello, " + c.Param("name") + "!",
			"request_id": c.RequestID(),
		})
	})

	app.GET("/health", func(c *rudraContext.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Demonstrates panic recovery — server continues after this.
	app.GET("/panic", func(c *rudraContext.Context) error {
		panic("intentional panic for testing recovery middleware")
	})

	// Demonstrates timeout — sleeps for the query param duration.
	// Try: /slow?wait=2s (fast) vs /slow?wait=15s (timeout)
	app.GET("/slow", func(c *rudraContext.Context) error {
		wait := c.QueryDefault("wait", "2s")
		d, err := time.ParseDuration(wait)
		if err != nil {
			d = 2 * time.Second
		}

		// This select respects the timeout context deadline.
		select {
		case <-time.After(d):
			return c.JSON(http.StatusOK, map[string]string{
				"message": "completed after " + d.String(),
			})
		case <-c.Request().Context().Done():
			// Context cancelled — timeout middleware handles the response.
			return c.Request().Context().Err()
		}
	})

	// API group with CORS-protected routes.
	api := app.Group("/api")
	api.GET("/data", func(c *rudraContext.Context) error {
		return c.JSON(http.StatusOK, map[string]any{
			"items": []string{"rudra", "ajaya", "vaya"},
			"total": 3,
		})
	})

	api.POST("/data", func(c *rudraContext.Context) error {
		return c.JSON(http.StatusCreated, map[string]string{
			"status": "created",
		})
	})

	// ── Start Server ──────────────────────────────────────────────────

	go func() {
		log.Println("rudra: middleware example starting on :8080")
		log.Println("rudra: middleware stack: Recovery → RequestID → Logger → Timeout → CORS")
		if err := app.Run(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("rudra: server error: %v", err)
		}
	}()

	app.ListenForShutdown()
}
