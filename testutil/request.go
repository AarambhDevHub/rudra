package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/AarambhDevHub/rudra/core"
)

// TestRequest is a fluent builder for test HTTP requests.
type TestRequest struct {
	method  string
	path    string
	headers map[string]string
	body    any
	app     *core.Engine
}

// NewTestRequest creates a new test request builder.
func NewTestRequest(app *core.Engine) *TestRequest {
	return &TestRequest{app: app, headers: make(map[string]string)}
}

// GET sets the method to GET.
func (t *TestRequest) GET(path string) *TestRequest {
	t.method = http.MethodGet
	t.path = path
	return t
}

// POST sets the method to POST with a JSON body.
func (t *TestRequest) POST(path string, body any) *TestRequest {
	t.method = http.MethodPost
	t.path = path
	t.body = body
	t.headers["Content-Type"] = "application/json"
	return t
}

// PUT sets the method to PUT with a JSON body.
func (t *TestRequest) PUT(path string, body any) *TestRequest {
	t.method = http.MethodPut
	t.path = path
	t.body = body
	t.headers["Content-Type"] = "application/json"
	return t
}

// PATCH sets the method to PATCH with a JSON body.
func (t *TestRequest) PATCH(path string, body any) *TestRequest {
	t.method = http.MethodPatch
	t.path = path
	t.body = body
	t.headers["Content-Type"] = "application/json"
	return t
}

// DELETE sets the method to DELETE.
func (t *TestRequest) DELETE(path string) *TestRequest {
	t.method = http.MethodDelete
	t.path = path
	return t
}

// Header adds a header to the request.
func (t *TestRequest) Header(key, value string) *TestRequest {
	t.headers[key] = value
	return t
}

// Bearer adds an Authorization Bearer token.
func (t *TestRequest) Bearer(token string) *TestRequest {
	t.headers["Authorization"] = "Bearer " + token
	return t
}

// Do executes the request and returns a TestResponse.
func (t *TestRequest) Do() *TestResponse {
	var bodyReader io.Reader
	if t.body != nil {
		data, _ := json.Marshal(t.body)
		bodyReader = strings.NewReader(string(data))
	}

	req := httptest.NewRequest(t.method, t.path, bodyReader)
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	t.app.ServeHTTP(w, req)

	return &TestResponse{recorder: w}
}