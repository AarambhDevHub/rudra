package middleware

import (
	"testing"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

func TestRateLimitDefaultConfig(t *testing.T) {
	cfg := DefaultRateLimitConfig()
	if cfg.Rate != 10 {
		t.Errorf("expected rate 10, got %f", cfg.Rate)
	}
	if cfg.Burst != 20 {
		t.Errorf("expected burst 20, got %d", cfg.Burst)
	}
}

func TestRateLimitAllowsWithinLimit(t *testing.T) {
	mw := RateLimit(RateLimitConfig{
		Rate:  100,
		Burst: 10,
	})

	// First request should be allowed.
	c, w := newTestContext("GET", "/api")
	c.SetNext(func() error { return c.String(200, "ok") })

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Check rate limit headers.
	if w.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("expected X-RateLimit-Limit header")
	}
	if w.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("expected X-RateLimit-Remaining header")
	}
	if w.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("expected X-RateLimit-Reset header")
	}
}

func TestRateLimitExceedsLimit(t *testing.T) {
	mw := RateLimit(RateLimitConfig{
		Rate:  0.001, // very slow refill
		Burst: 2,
	})

	// Exhaust the bucket.
	for i := 0; i < 2; i++ {
		c, _ := newTestContext("GET", "/api")
		c.SetNext(func() error { return c.String(200, "ok") })
		_ = mw(c)
	}

	// Third request should be rate limited.
	c, w := newTestContext("GET", "/api")
	c.SetNext(func() error { return c.String(200, "ok") })

	err := mw(c)
	if err == nil {
		t.Fatal("expected rate limit error, got nil")
	}

	if w.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429")
	}
}

func TestRateLimitCustomKeyFunc(t *testing.T) {
	mw := RateLimit(RateLimitConfig{
		Rate:  0.001,
		Burst: 1,
		KeyFunc: func(c *rudraContext.Context) string {
			return c.Header("X-API-Key")
		},
	})

	// Exhaust bucket for key "key-a".
	c1, _ := newTestContext("GET", "/api")
	c1.Request().Header.Set("X-API-Key", "key-a")
	c1.SetNext(func() error { return c1.String(200, "ok") })
	_ = mw(c1)

	// Key "key-a" should be limited.
	c2, _ := newTestContext("GET", "/api")
	c2.Request().Header.Set("X-API-Key", "key-a")
	c2.SetNext(func() error { return c2.String(200, "ok") })
	err := mw(c2)
	if err == nil {
		t.Error("expected rate limit for key-a")
	}

	// Key "key-b" should still be allowed.
	c3, _ := newTestContext("GET", "/api")
	c3.Request().Header.Set("X-API-Key", "key-b")
	c3.SetNext(func() error { return c3.String(200, "ok") })
	err = mw(c3)
	if err != nil {
		t.Errorf("key-b should not be limited: %v", err)
	}
}

func TestRateLimitCustomOnLimit(t *testing.T) {
	customCalled := false
	mw := RateLimit(RateLimitConfig{
		Rate:  0.001,
		Burst: 1,
		OnLimit: func(c *rudraContext.Context) error {
			customCalled = true
			return c.JSON(429, map[string]string{"error": "slow down"})
		},
	})

	// Exhaust bucket.
	c1, _ := newTestContext("GET", "/api")
	c1.SetNext(func() error { return c1.String(200, "ok") })
	_ = mw(c1)

	// Trigger limit.
	c2, _ := newTestContext("GET", "/api")
	c2.SetNext(func() error { return c2.String(200, "ok") })
	_ = mw(c2)

	if !customCalled {
		t.Error("expected custom OnLimit handler to be called")
	}
}
