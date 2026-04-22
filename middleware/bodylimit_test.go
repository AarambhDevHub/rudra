package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	rudraContext "github.com/AarambhDevHub/rudra/context"
)

func TestBodyLimitDefaultConfig(t *testing.T) {
	cfg := DefaultBodyLimitConfig()
	if cfg.Limit != 32<<20 {
		t.Errorf("expected limit 32MB, got %d", cfg.Limit)
	}
}

func TestBodyLimitAllowsSmallBody(t *testing.T) {
	mw := BodyLimit(BodyLimitConfig{Limit: 1024})

	body := bytes.NewReader([]byte("small body"))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", body)
	c := newTestContextFromReqResp(w, r)
	c.SetNext(func() error {
		return c.String(200, "ok")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBodyLimitRejectsLargeBody(t *testing.T) {
	mw := BodyLimit(BodyLimitConfig{Limit: 10})

	body := bytes.NewReader([]byte(strings.Repeat("x", 100)))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", body)
	c := newTestContextFromReqResp(w, r)
	c.SetNext(func() error {
		// Try to read the body — this should trigger the limit.
		buf := make([]byte, 200)
		_, readErr := c.Request().Body.Read(buf)
		if readErr != nil {
			return readErr
		}
		return c.String(200, "ok")
	})

	err := mw(c)
	if err == nil {
		t.Fatal("expected error for oversized body, got nil")
	}
}

func TestBodyLimitSkipsGET(t *testing.T) {
	mw := BodyLimit(BodyLimitConfig{Limit: 10})

	c, _ := newTestContext("GET", "/")
	c.SetNext(func() error {
		return c.String(200, "ok")
	})

	err := mw(c)
	if err != nil {
		t.Fatalf("unexpected error on GET: %v", err)
	}
}

func TestBodyLimitCustomHandler(t *testing.T) {
	customCalled := false
	mw := BodyLimit(BodyLimitConfig{
		Limit: 10,
		OnLimit: func(c *rudraContext.Context) error {
			customCalled = true
			return c.JSON(413, map[string]string{"error": "too big"})
		},
	})

	body := bytes.NewReader([]byte(strings.Repeat("x", 100)))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", body)
	c := newTestContextFromReqResp(w, r)
	c.SetNext(func() error {
		buf := make([]byte, 200)
		_, readErr := c.Request().Body.Read(buf)
		if readErr != nil {
			return readErr
		}
		return c.String(200, "ok")
	})

	_ = mw(c)
	if !customCalled {
		t.Error("expected custom OnLimit handler to be called")
	}
}

func newTestContextFromReqResp(w http.ResponseWriter, r *http.Request) *rudraContext.Context {
	c := rudraContext.New()
	c.Reset(w, r)
	return c
}
