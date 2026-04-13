package testutil

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// TestResponse wraps an httptest.ResponseRecorder with assertion helpers.
type TestResponse struct {
	recorder *httptest.ResponseRecorder
}

// Status asserts the response status code.
func (r *TestResponse) Status(t *testing.T, code int) *TestResponse {
	t.Helper()
	if r.recorder.Code != code {
		t.Errorf("expected status %d, got %d", code, r.recorder.Code)
	}
	return r
}

// Body returns the response body bytes.
func (r *TestResponse) Body() []byte {
	return r.recorder.Body.Bytes()
}

// BodyString returns the response body as a string.
func (r *TestResponse) BodyString() string {
	return r.recorder.Body.String()
}

// JSON unmarshals the response body into v.
func (r *TestResponse) JSON(t *testing.T, v any) *TestResponse {
	t.Helper()
	if err := json.Unmarshal(r.recorder.Body.Bytes(), v); err != nil {
		t.Errorf("failed to unmarshal JSON: %v", err)
	}
	return r
}

// HasHeader asserts the response has the given header value.
func (r *TestResponse) HasHeader(t *testing.T, key, val string) *TestResponse {
	t.Helper()
	got := r.recorder.Header().Get(key)
	if got != val {
		t.Errorf("expected header %s=%q, got %q", key, val, got)
	}
	return r
}

// Header returns a response header value.
func (r *TestResponse) Header(key string) string {
	return r.recorder.Header().Get(key)
}

// Code returns the HTTP status code.
func (r *TestResponse) Code() int {
	return r.recorder.Code
}