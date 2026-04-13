package router_test

import (
	"net/http"
	"testing"

	rudraContext "github.com/AarambhDevHub/rudra/context"
	"github.com/AarambhDevHub/rudra/router"
)

func nopHandler(ctx router.ContextParamSetter) error { return nil }

func newTestContext() *rudraContext.Context {
	c := rudraContext.New()
	c.Reset(nil, nil)
	return c
}

func TestRouterStaticRoute(t *testing.T) {
	r := router.New()
	called := false
	r.Add(http.MethodGet, "/api/v1/users", func(ctx router.ContextParamSetter) error {
		called = true
		return nil
	})

	c := newTestContext()
	h := r.Find(http.MethodGet, "/api/v1/users", c)
	if h == nil {
		t.Fatal("expected handler, got nil")
	}
	if err := h(c); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestRouterNotFound(t *testing.T) {
	r := router.New()
	r.Add(http.MethodGet, "/api/v1/users", nopHandler)

	c := newTestContext()
	h := r.Find(http.MethodGet, "/api/v1/missing", c)
	if h != nil {
		t.Error("expected nil for missing route")
	}
}

func TestRouterMethodNotAllowed(t *testing.T) {
	r := router.New()
	r.Add(http.MethodGet, "/api/v1/users", nopHandler)

	c := newTestContext()
	h := r.Find(http.MethodPost, "/api/v1/users", c)
	if h != nil {
		t.Error("expected nil for wrong method")
	}
}

func TestRouterParamRoute(t *testing.T) {
	r := router.New()
	r.Add(http.MethodGet, "/users/:id", func(ctx router.ContextParamSetter) error {
		return nil
	})

	c := newTestContext()
	h := r.Find(http.MethodGet, "/users/42", c)
	if h == nil {
		t.Fatal("expected handler, got nil")
	}
	if err := h(c); err != nil {
		t.Fatal(err)
	}
	if c.Param("id") != "42" {
		t.Errorf("expected id=42, got %s", c.Param("id"))
	}
}

func TestRouterMultipleParams(t *testing.T) {
	r := router.New()
	r.Add(http.MethodGet, "/orgs/:org/repos/:repo/commits/:sha", nopHandler)

	c := newTestContext()
	h := r.Find(http.MethodGet, "/orgs/rudra/repos/core/commits/abc123", c)
	if h == nil {
		t.Fatal("expected handler, got nil")
	}
	h(c)

	if c.Param("org") != "rudra" {
		t.Errorf("expected org=rudra, got %s", c.Param("org"))
	}
	if c.Param("repo") != "core" {
		t.Errorf("expected repo=core, got %s", c.Param("repo"))
	}
	if c.Param("sha") != "abc123" {
		t.Errorf("expected sha=abc123, got %s", c.Param("sha"))
	}
}

func TestRouterWildcard(t *testing.T) {
	r := router.New()
	r.Add(http.MethodGet, "/files/*filepath", nopHandler)

	c := newTestContext()
	h := r.Find(http.MethodGet, "/files/docs/readme.md", c)
	if h == nil {
		t.Fatal("expected handler, got nil")
	}
	h(c)

	if c.Param("filepath") != "docs/readme.md" {
		t.Errorf("expected filepath=docs/readme.md, got %s", c.Param("filepath"))
	}
}

func TestRouterMixedStaticAndParam(t *testing.T) {
	r := router.New()
	r.Add(http.MethodGet, "/users/:id/posts", nopHandler)
	r.Add(http.MethodGet, "/users/list", nopHandler)

	c := newTestContext()
	// Static should match first
	h := r.Find(http.MethodGet, "/users/list", c)
	if h == nil {
		t.Fatal("expected handler for /users/list")
	}

	c2 := newTestContext()
	h2 := r.Find(http.MethodGet, "/users/42/posts", c2)
	if h2 == nil {
		t.Fatal("expected handler for /users/42/posts")
	}
	h2(c2)
	if c2.Param("id") != "42" {
		t.Errorf("expected id=42, got %s", c2.Param("id"))
	}
}

func TestRouterConflictDetection(t *testing.T) {
	r := router.New()
	r.Add(http.MethodGet, "/users/:id", nopHandler)

	defer func() {
		if rec := recover(); rec == nil {
			t.Error("expected panic on duplicate route registration")
		}
	}()
	r.Add(http.MethodGet, "/users/:id", nopHandler)
}

func TestRouterNamedRoutes(t *testing.T) {
	r := router.New()
	r.Add(http.MethodGet, "/users/:id", nopHandler)
	r.Name(http.MethodGet, "/users/:id", "user.profile")

	url := r.URL("user.profile", "42")
	if url != "/users/42" {
		t.Errorf("expected /users/42, got %s", url)
	}
}

func TestRouterGroup(t *testing.T) {
	r := router.New()
	g := router.NewGroup("/api/v1", r)
	g.GET("/users", nopHandler)

	c := newTestContext()
	h := r.Find(http.MethodGet, "/api/v1/users", c)
	if h == nil {
		t.Fatal("expected handler for grouped route")
	}
}

func TestRouterNestedGroup(t *testing.T) {
	r := router.New()
	api := router.NewGroup("/api", r)
	v1 := api.Group("/v1")
	v1.GET("/users", nopHandler)

	c := newTestContext()
	h := r.Find(http.MethodGet, "/api/v1/users", c)
	if h == nil {
		t.Fatal("expected handler for nested group route")
	}
}

func BenchmarkRouterStatic(b *testing.B) {
	r := router.New()
	r.Add(http.MethodGet, "/api/v1/users", nopHandler)

	c := newTestContext()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.Reset(nil, nil)
		h := r.Find(http.MethodGet, "/api/v1/users", c)
		_ = h
	}
}

func BenchmarkRouterParams(b *testing.B) {
	r := router.New()
	r.Add(http.MethodGet, "/orgs/:org/repos/:repo/commits/:sha", nopHandler)

	c := newTestContext()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.Reset(nil, nil)
		h := r.Find(http.MethodGet, "/orgs/rudra/repos/core/commits/abc123", c)
		_ = h
	}
}

func BenchmarkRouterWildcard(b *testing.B) {
	r := router.New()
	r.Add(http.MethodGet, "/files/*filepath", nopHandler)

	c := newTestContext()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.Reset(nil, nil)
		h := r.Find(http.MethodGet, "/files/docs/readme.md", c)
		_ = h
	}
}