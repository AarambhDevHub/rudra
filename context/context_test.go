package context_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

func TestContextReset(t *testing.T) {
	c := rudraContext.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test?foo=bar", nil)

	c.Reset(w, r)

	if c.Method() != http.MethodGet {
		t.Errorf("expected GET, got %s", c.Method())
	}
	if c.Path() != "/test" {
		t.Errorf("expected /test, got %s", c.Path())
	}
	if c.Query("foo") != "bar" {
		t.Errorf("expected bar, got %s", c.Query("foo"))
	}
}

func TestContextParams(t *testing.T) {
	c := rudraContext.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Reset(w, r)

	c.SetParam("id", "42")
	c.SetParam("name", "rudra")

	if c.Param("id") != "42" {
		t.Errorf("expected 42, got %s", c.Param("id"))
	}
	if c.Param("name") != "rudra" {
		t.Errorf("expected rudra, got %s", c.Param("name"))
	}
	if c.Param("missing") != "" {
		t.Errorf("expected empty, got %s", c.Param("missing"))
	}
}

func TestContextMaxParams(t *testing.T) {
	c := rudraContext.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Reset(w, r)

	for i := 0; i < 20; i++ {
		c.SetParam("p", "v")
	}

	if len(c.Params()) != 16 {
		t.Errorf("expected max 16 params, got %d", len(c.Params()))
	}
}

func TestContextStore(t *testing.T) {
	c := rudraContext.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Reset(w, r)

	c.Set("key", "value")
	val, ok := c.Get("key")
	if !ok || val != "value" {
		t.Errorf("expected value, got %v, ok=%v", val, ok)
	}

	_, ok = c.Get("missing")
	if ok {
		t.Error("expected missing key to return false")
	}
}

func TestContextQueryDefault(t *testing.T) {
	c := rudraContext.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/?page=3", nil)
	c.Reset(w, r)

	if c.QueryDefault("page", "1") != "3" {
		t.Errorf("expected 3, got %s", c.QueryDefault("page", "1"))
	}
	if c.QueryDefault("missing", "default") != "default" {
		t.Errorf("expected default, got %s", c.QueryDefault("missing", "default"))
	}
}

func TestContextAbort(t *testing.T) {
	c := rudraContext.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Reset(w, r)

	if c.IsAborted() {
		t.Error("should not be aborted initially")
	}
	c.Abort()
	if !c.IsAborted() {
		t.Error("should be aborted after Abort()")
	}
}

func TestContextRelease(t *testing.T) {
	c := rudraContext.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Reset(w, r)

	c.Release()
	if c.Writer() != nil {
		t.Error("writer should be nil after Release")
	}
	if c.Request() != nil {
		t.Error("request should be nil after Release")
	}
}

func TestContextRealIP(t *testing.T) {
	c := rudraContext.New()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.1, 70.41.3.18")
	r.RemoteAddr = "192.168.1.1:12345"
	c.Reset(w, r)

	if c.RealIP() != "203.0.113.1" {
		t.Errorf("expected 203.0.113.1, got %s", c.RealIP())
	}
}

func TestContextNext(t *testing.T) {
	c := rudraContext.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Reset(w, r)

	called := false
	c.SetNext(func() error {
		called = true
		return nil
	})

	if err := c.Next(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("next handler was not called")
	}
}

func TestContextPoolAcquireRelease(t *testing.T) {
	pool := rudraContext.NewPool()

	c1 := pool.Get()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c1.Reset(w, r)
	c1.SetParam("id", "42")
	pool.Put(c1)

	// Re-acquire from pool — Reset() must be called by the engine before use
	c2 := pool.Get()
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	c2.Reset(w2, r2)

	// After Reset, params should be cleared
	if c2.Param("id") != "" {
		t.Error("params should be cleared after Reset")
	}
}

func BenchmarkContextAcquireRelease(b *testing.B) {
	pool := rudraContext.NewPool()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c := pool.Get()
		c.Reset(w, r)
		c.SetParam("id", "42")
		_ = c.Param("id")
		pool.Put(c)
	}
}

func BenchmarkContextSetParam(b *testing.B) {
	c := rudraContext.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Reset(w, r)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.SetParam("id", "42")
		_ = c.Param("id")
	}
}