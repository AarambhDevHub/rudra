package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

func newTestContext(method, path string) (*rudraContext.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)
	c := rudraContext.New()
	c.Reset(w, r)
	return c, w
}

func TestLoggerDefaultConfig(t *testing.T) {
	cfg := DefaultLoggerConfig()
	if cfg.Format != "json" {
		t.Errorf("expected format 'json', got '%s'", cfg.Format)
	}
	if cfg.Output == nil {
		t.Error("expected non-nil output")
	}
}

func TestLoggerJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	mw := Logger(LoggerConfig{
		Output: &buf,
		Format: "json",
	})

	c, _ := newTestContext("GET", "/hello")
	c.SetNext(func() error {
		return c.JSON(200, map[string]string{"ok": "true"})
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify JSON log output.
	output := buf.String()
	if output == "" {
		t.Fatal("expected log output, got empty string")
	}

	var logEntry map[string]any
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("expected valid JSON log, got: %s (err: %v)", output, err)
	}

	if logEntry["method"] != "GET" {
		t.Errorf("expected method=GET, got %v", logEntry["method"])
	}
	if logEntry["path"] != "/hello" {
		t.Errorf("expected path=/hello, got %v", logEntry["path"])
	}
}

func TestLoggerTextFormat(t *testing.T) {
	var buf bytes.Buffer
	mw := Logger(LoggerConfig{
		Output: &buf,
		Format: "text",
	})

	c, _ := newTestContext("POST", "/api/users")
	c.SetNext(func() error {
		return c.String(201, "created")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "method=POST") {
		t.Errorf("expected text log to contain method=POST, got: %s", output)
	}
}

func TestLoggerCommonFormat(t *testing.T) {
	var buf bytes.Buffer
	mw := Logger(LoggerConfig{
		Output: &buf,
		Format: "common",
	})

	c, _ := newTestContext("GET", "/index.html")
	c.SetNext(func() error {
		return c.String(200, "hello")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "GET /index.html") {
		t.Errorf("expected common log to contain request line, got: %s", output)
	}
}

func TestLoggerSkipPaths(t *testing.T) {
	var buf bytes.Buffer
	mw := Logger(LoggerConfig{
		Output:    &buf,
		Format:    "json",
		SkipPaths: []string{"/health", "/metrics"},
	})

	// Request to a skipped path — should produce no log.
	c, _ := newTestContext("GET", "/health")
	c.SetNext(func() error {
		return c.String(200, "ok")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no log output for skipped path, got: %s", buf.String())
	}

	// Request to a non-skipped path — should produce log.
	buf.Reset()
	c2, _ := newTestContext("GET", "/api")
	c2.SetNext(func() error {
		return c2.String(200, "hello")
	})

	err = mw(c2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected log output for non-skipped path, got empty")
	}
}

func TestLoggerCapturesBytesWritten(t *testing.T) {
	var buf bytes.Buffer
	mw := Logger(LoggerConfig{
		Output: &buf,
		Format: "json",
	})

	c, _ := newTestContext("GET", "/data")
	c.SetNext(func() error {
		return c.String(200, "hello, world!")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var logEntry map[string]any
	if err := json.Unmarshal([]byte(buf.String()), &logEntry); err != nil {
		t.Fatalf("failed to parse log: %v", err)
	}

	bytesOut, ok := logEntry["bytes_out"].(float64)
	if !ok || bytesOut <= 0 {
		t.Errorf("expected positive bytes_out, got %v", logEntry["bytes_out"])
	}
}

func TestLoggerCapturesStatusCode(t *testing.T) {
	var buf bytes.Buffer
	mw := Logger(LoggerConfig{
		Output: &buf,
		Format: "json",
	})

	c, _ := newTestContext("GET", "/notfound")
	c.SetNext(func() error {
		return c.String(404, "not found")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var logEntry map[string]any
	if err := json.Unmarshal([]byte(buf.String()), &logEntry); err != nil {
		t.Fatalf("failed to parse log: %v", err)
	}

	status, ok := logEntry["status"].(float64)
	if !ok || int(status) != 404 {
		t.Errorf("expected status=404, got %v", logEntry["status"])
	}
}

func TestResponseWriterInterfaces(t *testing.T) {
	w := httptest.NewRecorder()
	rw := newResponseWriter(w)

	// Verify Flush interface.
	rw.Flush() // should not panic

	// Verify Write tracks bytes.
	n, err := rw.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}
	if rw.bytesWritten != 5 {
		t.Errorf("expected bytesWritten=5, got %d", rw.bytesWritten)
	}

	// Verify status code tracking.
	rw2 := newResponseWriter(httptest.NewRecorder())
	rw2.WriteHeader(http.StatusCreated)
	if rw2.statusCode != 201 {
		t.Errorf("expected statusCode=201, got %d", rw2.statusCode)
	}

	// Verify double WriteHeader is ignored.
	rw2.WriteHeader(http.StatusBadRequest)
	if rw2.statusCode != 201 {
		t.Errorf("expected statusCode still 201 after double write, got %d", rw2.statusCode)
	}
}
