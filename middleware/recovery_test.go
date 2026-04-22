package middleware

import (
	"bytes"
	"strings"
	"testing"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

func TestRecoveryDefaultConfig(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	if !cfg.LogStackTrace {
		t.Error("expected LogStackTrace=true")
	}
	if cfg.Output == nil {
		t.Error("expected non-nil output")
	}
}

func TestRecoveryPanicInHandler(t *testing.T) {
	var logBuf bytes.Buffer
	mw := Recovery(RecoveryConfig{
		LogStackTrace: true,
		Output:        &logBuf,
	})

	c, w := newTestContext("GET", "/panic")
	c.SetNext(func() error {
		panic("test panic!")
	})

	// Recovery should not re-panic — it should catch it.
	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error from recovery: %v", err)
	}

	// Verify 500 JSON response was sent.
	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "internal server error") {
		t.Errorf("expected error message in body, got: %s", body)
	}

	// Verify stack trace was logged (not sent to client).
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "panic recovered") {
		t.Errorf("expected panic logged, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "test panic!") {
		t.Errorf("expected panic value in log, got: %s", logOutput)
	}

	// Verify stack trace is NOT in the response body.
	if strings.Contains(body, "goroutine") {
		t.Error("stack trace leaked to client response!")
	}
}

func TestRecoveryNoPanic(t *testing.T) {
	var logBuf bytes.Buffer
	mw := Recovery(RecoveryConfig{
		LogStackTrace: true,
		Output:        &logBuf,
	})

	c, w := newTestContext("GET", "/ok")
	c.SetNext(func() error {
		return c.String(200, "all good")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// No panic — no log output.
	if logBuf.Len() != 0 {
		t.Errorf("expected no log output without panic, got: %s", logBuf.String())
	}
}

func TestRecoveryOnPanicHook(t *testing.T) {
	var hookCalled bool
	var hookPanicValue any
	var hookStack []byte

	var logBuf bytes.Buffer
	mw := Recovery(RecoveryConfig{
		LogStackTrace: true,
		Output:        &logBuf,
		OnPanic: func(c *rudraContext.Context, err any, stack []byte) {
			hookCalled = true
			hookPanicValue = err
			hookStack = stack
		},
	})

	c, _ := newTestContext("GET", "/panic-hook")
	c.SetNext(func() error {
		panic("hook test")
	})

	_ = mw(c)

	if !hookCalled {
		t.Error("expected OnPanic hook to be called")
	}
	if hookPanicValue != "hook test" {
		t.Errorf("expected panic value 'hook test', got %v", hookPanicValue)
	}
	if len(hookStack) == 0 {
		t.Error("expected stack trace in hook, got empty")
	}
}

func TestRecoveryNoStackTrace(t *testing.T) {
	var logBuf bytes.Buffer
	mw := Recovery(RecoveryConfig{
		LogStackTrace: false,
		Output:        &logBuf,
	})

	c, w := newTestContext("GET", "/panic-no-trace")
	c.SetNext(func() error {
		panic("silent panic")
	})

	_ = mw(c)

	// Response should still be 500.
	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	// No log output since LogStackTrace=false.
	if logBuf.Len() != 0 {
		t.Errorf("expected no log output with LogStackTrace=false, got: %s", logBuf.String())
	}
}

func TestRecoveryPropagatesHandlerError(t *testing.T) {
	mw := Recovery()

	c, _ := newTestContext("GET", "/error")
	c.SetNext(func() error {
		return c.AbortWithError(400, &testErr{"bad request"})
	})

	err := mw(c)
	if err == nil {
		t.Fatal("expected error to be propagated, got nil")
	}
}

type testErr struct{ msg string }

func (e *testErr) Error() string { return e.msg }
