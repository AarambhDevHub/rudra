package errors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	rudraContext "github.com/AarambhDevHub/rudra/context"
	rudraErrors "github.com/AarambhDevHub/rudra/errors"
)

func TestRudraErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		err      *rudraErrors.RudraError
		code     int
		message  string
	}{
		{"BadRequest", rudraErrors.BadRequest("bad"), 400, "bad"},
		{"Unauthorized", rudraErrors.Unauthorized("unauth"), 401, "unauth"},
		{"Forbidden", rudraErrors.Forbidden("forbidden"), 403, "forbidden"},
		{"NotFound", rudraErrors.NotFound("missing"), 404, "missing"},
		{"MethodNotAllowed", rudraErrors.MethodNotAllowed("not allowed"), 405, "not allowed"},
		{"Conflict", rudraErrors.Conflict("conflict"), 409, "conflict"},
		{"UnprocessableEntity", rudraErrors.UnprocessableEntity("unprocessable"), 422, "unprocessable"},
		{"TooManyRequests", rudraErrors.TooManyRequests("slow down"), 429, "slow down"},
		{"InternalServerError", rudraErrors.InternalServerError("boom"), 500, "boom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("expected code %d, got %d", tt.code, tt.err.Code)
			}
			if tt.err.Message != tt.message {
				t.Errorf("expected message %q, got %q", tt.message, tt.err.Message)
			}
			if tt.err.Error() == "" {
				t.Error("Error() should not be empty")
			}
		})
	}
}

func TestRudraErrorDetail(t *testing.T) {
	err := rudraErrors.BadRequest("invalid input", map[string]string{"field": "email"})
	if err.Detail == nil {
		t.Error("expected detail to be set")
	}
}

func TestDefaultErrorHandler(t *testing.T) {
	c := rudraContext.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Reset(w, r)

	rudraErrors.DefaultErrorHandler(c, rudraErrors.NotFound("page not found"))

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != float64(404) {
		t.Errorf("expected code 404 in JSON, got %v", resp["code"])
	}
}

func TestDefaultErrorHandlerUnknownError(t *testing.T) {
	c := rudraContext.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Reset(w, r)

	rudraErrors.DefaultErrorHandler(c, http.ErrAbortHandler)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for unknown error, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "internal server error" {
		t.Errorf("should not leak internals, got %v", resp["message"])
	}
}

func TestErrorAs(t *testing.T) {
	err := rudraErrors.NotFound("missing")
	var re *rudraErrors.RudraError
	if !rudraErrors.As(err, &re) {
		t.Error("expected As to succeed for RudraError")
	}
	if re.Code != 404 {
		t.Errorf("expected 404, got %d", re.Code)
	}
}