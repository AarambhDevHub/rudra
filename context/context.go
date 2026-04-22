package context

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

const maxParams = 16

// Param is a single key-value pair from the URL path.
type Param struct {
	Key   string
	Value string
}

// Context is Rudra's per-request state container.
// Pooled via sync.Pool — never create one directly.
type Context struct {
	writer     http.ResponseWriter
	request    *http.Request
	params     [maxParams]Param
	paramCount int
	store      map[string]any
	statusCode int
	written    bool
	next       func() error
	requestID  string
	body       []byte
	errors     []error
	index      int8
	aborted    bool
}

// New creates a new Context for pool initialization.
func New() *Context {
	return &Context{
		body: make([]byte, 0, 1024),
	}
}

// Reset resets the Context for reuse from the pool.
func (c *Context) Reset(w http.ResponseWriter, r *http.Request) {
	c.writer = w
	c.request = r
	c.paramCount = 0
	c.statusCode = 200
	c.written = false
	c.aborted = false
	c.index = 0
	c.next = nil
	c.requestID = ""
	c.body = c.body[:0]
	c.errors = c.errors[:0]
	if c.store != nil {
		for k := range c.store {
			delete(c.store, k)
		}
	}
}

// Release clears sensitive fields before returning to the pool.
func (c *Context) Release() {
	c.writer = nil
	c.request = nil
}

// SetParam adds a URL path parameter. Called by the router during matching.
func (c *Context) SetParam(key, value string) {
	if c.paramCount < maxParams {
		c.params[c.paramCount] = Param{Key: key, Value: value}
		c.paramCount++
	}
}

// Param retrieves a URL path parameter by name.
func (c *Context) Param(key string) string {
	for i := 0; i < c.paramCount; i++ {
		if c.params[i].Key == key {
			return c.params[i].Value
		}
	}
	return ""
}

// Params returns all captured path parameters.
func (c *Context) Params() []Param {
	return c.params[:c.paramCount]
}

// Query retrieves a URL query string parameter.
func (c *Context) Query(key string) string {
	return c.request.URL.Query().Get(key)
}

// QueryDefault retrieves a query parameter, returning def if absent.
func (c *Context) QueryDefault(key, def string) string {
	if v := c.Query(key); v != "" {
		return v
	}
	return def
}

// Header retrieves a request header.
func (c *Context) Header(key string) string {
	return c.request.Header.Get(key)
}

// SetHeader sets a response header.
func (c *Context) SetHeader(key, value string) {
	c.writer.Header().Set(key, value)
}

// Set stores a value in the per-request store.
func (c *Context) Set(key string, val any) {
	if c.store == nil {
		c.store = make(map[string]any, 4)
	}
	c.store[key] = val
}

// Get retrieves a value from the per-request store.
func (c *Context) Get(key string) (any, bool) {
	if c.store == nil {
		return nil, false
	}
	v, ok := c.store[key]
	return v, ok
}

// MustGet retrieves a value, panicking if absent.
func (c *Context) MustGet(key string) any {
	v, ok := c.Get(key)
	if !ok {
		panic("rudra: key not found in context: " + key)
	}
	return v
}

// Method returns the HTTP method of the request.
func (c *Context) Method() string { return c.request.Method }

// Path returns the URL path of the request.
func (c *Context) Path() string { return c.request.URL.Path }

// RealIP returns the real client IP, respecting X-Forwarded-For.
func (c *Context) RealIP() string {
	if ip := c.request.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.SplitN(ip, ",", 2)[0]
	}
	if ip := c.request.Header.Get("X-Real-Ip"); ip != "" {
		return ip
	}
	ip, _, _ := net.SplitHostPort(c.request.RemoteAddr)
	return ip
}

// StatusCode returns the response status code.
func (c *Context) StatusCode() int { return c.statusCode }

// Written returns whether the response has been written.
func (c *Context) Written() bool { return c.written }

// IsAborted returns whether the middleware chain was stopped.
func (c *Context) IsAborted() bool { return c.aborted }

// Abort stops the middleware chain from continuing.
func (c *Context) Abort() {
	c.aborted = true
}

// AbortWithError stops the chain and returns an error with HTTP code info.
func (c *Context) AbortWithError(code int, err error) error {
	c.Abort()
	c.SetStatus(code)
	return fmt.Errorf("rudra: %d: %w", code, err)
}

// Next calls the next handler in the middleware chain.
func (c *Context) Next() error {
	if c.next != nil && !c.aborted {
		return c.next()
	}
	return nil
}

// SetNext sets the next handler in the chain.
func (c *Context) SetNext(fn func() error) {
	c.next = fn
}

// RequestID returns the X-Request-ID for this request.
func (c *Context) RequestID() string { return c.requestID }

// SetRequestID sets the request ID.
func (c *Context) SetRequestID(id string) { c.requestID = id }

// Request returns the underlying *http.Request.
func (c *Context) Request() *http.Request { return c.request }

// Writer returns the underlying http.ResponseWriter.
func (c *Context) Writer() http.ResponseWriter { return c.writer }

// ContentType returns the Content-Type header without parameters.
func (c *Context) ContentType() string {
	ct := c.Header("Content-Type")
	for i := range ct {
		if ct[i] == ';' {
			return ct[:i]
		}
	}
	return ct
}

// Body returns the request body bytes (lazily read).
func (c *Context) Body() []byte { return c.body }

// SetBody sets the request body bytes (called by binders).
func (c *Context) SetBody(b []byte) { c.body = b }

// SetStatus writes the status code.
func (c *Context) SetStatus(code int) {
	if !c.written {
		c.statusCode = code
		c.writer.WriteHeader(code)
		c.written = true
	}
}

// AppendError accumulates a non-fatal error.
func (c *Context) AppendError(err error) { c.errors = append(c.errors, err) }

// Errors returns accumulated errors.
func (c *Context) Errors() []error { return c.errors }

// NoContent sends a 204 No Content response.
func (c *Context) NoContent() error {
	c.SetStatus(http.StatusNoContent)
	return nil
}

// Redirect sends a redirect response.
func (c *Context) Redirect(code int, url string) error {
	http.Redirect(c.writer, c.request, url, code)
	c.statusCode = code
	c.written = true
	return nil
}

// SetRequest replaces the underlying *http.Request.
// Used by middleware (e.g. Timeout) to inject a derived request with deadline context.
func (c *Context) SetRequest(r *http.Request) {
	c.request = r
}

// SetWriter replaces the underlying http.ResponseWriter.
// Used by middleware (e.g. Logger) to wrap the writer for interception.
func (c *Context) SetWriter(w http.ResponseWriter) {
	c.writer = w
}

// UserAgent returns the User-Agent header.
func (c *Context) UserAgent() string {
	return c.request.UserAgent()
}