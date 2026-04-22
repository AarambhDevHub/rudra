package middleware

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

func TestTimeoutDefaultConfig(t *testing.T) {
	cfg := DefaultTimeoutConfig()
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", cfg.Timeout)
	}
}

func TestTimeoutFastHandler(t *testing.T) {
	mw := Timeout(TimeoutConfig{Timeout: 5 * time.Second})

	c, w := newTestContext("GET", "/fast")
	c.SetNext(func() error {
		return c.String(200, "fast response")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestTimeoutExceeded(t *testing.T) {
	mw := Timeout(TimeoutConfig{Timeout: 50 * time.Millisecond})

	c, _ := newTestContext("GET", "/slow")
	c.SetNext(func() error {
		// Respect the context cancellation — this is the correct pattern.
		select {
		case <-time.After(5 * time.Second):
			return c.String(200, "too late")
		case <-c.Request().Context().Done():
			return c.Request().Context().Err()
		}
	})

	err := mw(c)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout in error message, got: %s", err.Error())
	}
}

func TestTimeoutCustomHandler(t *testing.T) {
	var customCalled atomic.Bool
	mw := Timeout(TimeoutConfig{
		Timeout: 50 * time.Millisecond,
		OnTimeout: func(c *rudraContext.Context) error {
			customCalled.Store(true)
			return nil
		},
	})

	c, _ := newTestContext("GET", "/slow-custom")
	c.SetNext(func() error {
		// Respect context cancellation.
		select {
		case <-time.After(5 * time.Second):
			return nil
		case <-c.Request().Context().Done():
			return c.Request().Context().Err()
		}
	})

	_ = mw(c)

	// Give goroutines time to settle.
	time.Sleep(10 * time.Millisecond)

	if !customCalled.Load() {
		t.Error("expected custom timeout handler to be called")
	}
}

func TestTimeoutPropagatesContext(t *testing.T) {
	mw := Timeout(TimeoutConfig{Timeout: 5 * time.Second})

	var deadlineSet bool
	c, _ := newTestContext("GET", "/context")
	c.SetNext(func() error {
		_, deadlineSet = c.Request().Context().Deadline()
		return c.String(200, "ok")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deadlineSet {
		t.Error("expected deadline to be set on request context")
	}
}
