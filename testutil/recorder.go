package testutil

import "net/http/httptest"

// ResponseRecorder wraps httptest.ResponseRecorder for convenience.
type ResponseRecorder struct {
	*httptest.ResponseRecorder
}

// NewRecorder creates a new ResponseRecorder.
func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{httptest.NewRecorder()}
}