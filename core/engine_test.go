package core_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AarambhDevHub/rudra/core"
	rudraContext "github.com/AarambhDevHub/rudra/context"
)

func TestEngineBasicRoute(t *testing.T) {
	app := core.New()

	app.GET("/hello", func(c *rudraContext.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "hello"})
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["message"] != "hello" {
		t.Errorf("expected hello, got %s", resp["message"])
	}
}

func TestEngineParamRoute(t *testing.T) {
	app := core.New()

	app.GET("/users/:id", func(c *rudraContext.Context) error {
		return c.String(http.StatusOK, c.Param("id"))
	})

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "42" {
		t.Errorf("expected 42, got %s", w.Body.String())
	}
}

func TestEngineNotFound(t *testing.T) {
	app := core.New()

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestEngineMiddleware(t *testing.T) {
	app := core.New()

	order := []string{}
	app.Use(func(c *rudraContext.Context) error {
		order = append(order, "mw1-before")
		err := c.Next()
		order = append(order, "mw1-after")
		return err
	})
	app.Use(func(c *rudraContext.Context) error {
		order = append(order, "mw2-before")
		err := c.Next()
		order = append(order, "mw2-after")
		return err
	})

	app.GET("/test", func(c *rudraContext.Context) error {
		order = append(order, "handler")
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected order %v, got %v", expected, order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("expected %s at %d, got %s", v, i, order[i])
		}
	}
}

func TestEngineMiddlewareAbort(t *testing.T) {
	app := core.New()

	app.Use(func(c *rudraContext.Context) error {
		c.Abort()
		return c.String(http.StatusForbidden, "forbidden")
	})

	handlerCalled := false
	app.GET("/test", func(c *rudraContext.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if handlerCalled {
		t.Error("handler should not be called after abort")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestEngineRouteGroup(t *testing.T) {
	app := core.New()

	api := app.Group("/api/v1")
	api.GET("/users", func(c *rudraContext.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"path": "users"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestEngineNestedGroups(t *testing.T) {
	app := core.New()

	api := app.Group("/api")
	v1 := api.Group("/v1")
	v1.GET("/users", func(c *rudraContext.Context) error {
		return c.String(http.StatusOK, "users-v1")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Body.String() != "users-v1" {
		t.Errorf("expected users-v1, got %s", w.Body.String())
	}
}

func TestEngineAllMethods(t *testing.T) {
	app := core.New()

	methods := []struct {
		method string
		fn     func(path string, h core.HandlerFunc, mw ...core.HandlerFunc)
	}{
		{http.MethodGet, app.GET},
		{http.MethodPost, app.POST},
		{http.MethodPut, app.PUT},
		{http.MethodPatch, app.PATCH},
		{http.MethodDelete, app.DELETE},
		{http.MethodOptions, app.OPTIONS},
		{http.MethodHead, app.HEAD},
	}

	for _, m := range methods {
		m.fn("/test", func(c *rudraContext.Context) error {
			return c.String(http.StatusOK, m.method)
		})

		req := httptest.NewRequest(m.method, "/test", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200 for %s, got %d", m.method, w.Code)
		}
	}
}

func TestEngineJSONResponse(t *testing.T) {
	app := core.New()

	app.GET("/json", func(c *rudraContext.Context) error {
		return c.JSON(http.StatusOK, map[string]any{
			"framework": "Rudra",
			"version":   "0.0.1",
			"features":  []string{"fast", "fierce"},
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/json", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("expected JSON content type, got %s", w.Header().Get("Content-Type"))
	}
}

func TestEngineHTMLResponse(t *testing.T) {
	app := core.New()

	app.GET("/html", func(c *rudraContext.Context) error {
		return c.HTML(http.StatusOK, "<h1>Rudra</h1>")
	})

	req := httptest.NewRequest(http.MethodGet, "/html", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("expected HTML content type, got %s", w.Header().Get("Content-Type"))
	}
	if w.Body.String() != "<h1>Rudra</h1>" {
		t.Errorf("expected <h1>Rudra</h1>, got %s", w.Body.String())
	}
}

func TestEngineSetErrorHandler(t *testing.T) {
	app := core.New()

	app.SetErrorHandler(func(c *rudraContext.Context, err error) {
		_ = c.JSON(http.StatusNotFound, map[string]string{
			"custom": "error",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["custom"] != "error" {
		t.Errorf("expected custom error, got %v", resp)
	}
}

func TestEngineContextPool(t *testing.T) {
	app := core.New()

	app.GET("/pool", func(c *rudraContext.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Make multiple requests — context should be recycled
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/pool", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200 on iteration %d, got %d", i, w.Code)
		}
	}
}

func BenchmarkEngineServeHTTP(b *testing.B) {
	app := core.New()
	app.GET("/bench", func(c *rudraContext.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/bench", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkEngineServeHTTPWithParams(b *testing.B) {
	app := core.New()
	app.GET("/users/:id/posts/:postId", func(c *rudraContext.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"id":     c.Param("id"),
			"postId": c.Param("postId"),
		})
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/users/42/posts/7", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkEngineMiddlewareChain(b *testing.B) {
	app := core.New()

	app.Use(func(c *rudraContext.Context) error { return c.Next() })
	app.Use(func(c *rudraContext.Context) error { return c.Next() })
	app.Use(func(c *rudraContext.Context) error { return c.Next() })

	app.GET("/bench", func(c *rudraContext.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/bench", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}