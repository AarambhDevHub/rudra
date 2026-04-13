package benchmarks

import (
	"net/http"
	"net/http/httptest"
	"testing"

	rudraContext "github.com/AarambhDevHub/rudra/context"
	"github.com/AarambhDevHub/rudra/core"
	"github.com/AarambhDevHub/rudra/router"
)

func nopHandler(c *rudraContext.Context) error { return nil }

// --- Router micro-benchmarks ---

func BenchmarkRudraRouterStatic(b *testing.B) {
	r := router.New()
	r.Add(http.MethodGet, "/api/v1/users", func(ctx router.ContextParamSetter) error { return nil })

	c := rudraContext.New()
	c.Reset(nil, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.Reset(nil, nil)
		_ = r.Find(http.MethodGet, "/api/v1/users", c)
	}
}

func BenchmarkRudraRouterParams(b *testing.B) {
	r := router.New()
	r.Add(http.MethodGet, "/orgs/:org/repos/:repo/commits/:sha",
		func(ctx router.ContextParamSetter) error { return nil })

	c := rudraContext.New()
	c.Reset(nil, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.Reset(nil, nil)
		_ = r.Find(http.MethodGet, "/orgs/rudra/repos/core/commits/abc123", c)
	}
}

// --- Full framework benchmarks ---

func BenchmarkRudraEngineJSON(b *testing.B) {
	app := core.New()
	app.GET("/bench", func(c *rudraContext.Context) error {
		return c.JSON(200, map[string]string{"message": "hello"})
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/bench", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkRudraEngineParams(b *testing.B) {
	app := core.New()
	app.GET("/users/:id/posts/:postId", func(c *rudraContext.Context) error {
		return c.String(200, c.Param("id")+":"+c.Param("postId"))
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/users/42/posts/7", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkRudraEngineMiddleware(b *testing.B) {
	app := core.New()
	app.Use(func(c *rudraContext.Context) error { return c.Next() })
	app.Use(func(c *rudraContext.Context) error { return c.Next() })
	app.Use(func(c *rudraContext.Context) error { return c.Next() })

	app.GET("/bench", func(c *rudraContext.Context) error {
		return c.String(200, "ok")
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/bench", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}