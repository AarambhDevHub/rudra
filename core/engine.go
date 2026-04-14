package core

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"

	rudraContext "github.com/AarambhDevHub/rudra/context"
	"github.com/AarambhDevHub/rudra/errors"
	"github.com/AarambhDevHub/rudra/router"
)

// HandlerFunc is the core handler type in Rudra.
type HandlerFunc func(*rudraContext.Context) error

// Engine is the heart of Rudra. It owns the router, middleware chain,
// server configuration, and the context pool.
type Engine struct {
	router       *router.Router
	pool         sync.Pool
	middleware   []HandlerFunc
	errorHandler errors.ErrorHandler
	opts         *Options
	server       *http.Server
	shutdownCh   chan struct{}
	once         sync.Once
	mu           sync.RWMutex
}

// New creates a new Rudra Engine with default options.
func New(opts ...Option) *Engine {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	e := &Engine{
		router:       router.New(),
		opts:         options,
		shutdownCh:   make(chan struct{}),
		errorHandler: errors.DefaultErrorHandler,
	}

	e.pool = sync.Pool{
		New: func() any {
			return rudraContext.New()
		},
	}

	return e
}

// ServeHTTP implements http.Handler. Entry point for every incoming HTTP request.
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := e.pool.Get().(*rudraContext.Context)
	c.Reset(w, r)

	defer func() {
		if rec := recover(); rec != nil {
			e.errorHandler(c, errors.InternalServerError("internal server error"))
		}
		c.Release()
		e.pool.Put(c)
	}()

	// Trailing slash normalization
	path := r.URL.Path
	if !e.opts.StrictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		r.URL.Path = path[:len(path)-1]
	}

	// Find matching route
	h := e.router.Find(r.Method, r.URL.Path, c)
	if h == nil {
		e.errorHandler(c, errors.NotFound("route not found: "+r.URL.Path))
		return
	}

	// Build middleware chain: global middleware wraps the matched handler
	composed := e.composeHandler(h, e.middleware...)

	if err := composed(c); err != nil {
		e.errorHandler(c, err)
	}
}

// composeHandler wraps a router handler with core middleware in reverse order (onion model).
func (e *Engine) composeHandler(h router.HandlerFunc, middleware ...HandlerFunc) func(*rudraContext.Context) error {
	composed := func(c *rudraContext.Context) error {
		return h(c)
	}
	for i := len(middleware) - 1; i >= 0; i-- {
		mw := middleware[i]
		next := composed
		composed = func(c *rudraContext.Context) error {
			c.SetNext(func() error { return next(c) })
			return mw(c)
		}
	}
	return composed
}

// Use registers global middleware.
func (e *Engine) Use(middleware ...HandlerFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// GET registers a GET route.
func (e *Engine) GET(path string, h HandlerFunc, mw ...HandlerFunc) {
	e.addRoute(http.MethodGet, path, h, mw...)
}

// POST registers a POST route.
func (e *Engine) POST(path string, h HandlerFunc, mw ...HandlerFunc) {
	e.addRoute(http.MethodPost, path, h, mw...)
}

// PUT registers a PUT route.
func (e *Engine) PUT(path string, h HandlerFunc, mw ...HandlerFunc) {
	e.addRoute(http.MethodPut, path, h, mw...)
}

// PATCH registers a PATCH route.
func (e *Engine) PATCH(path string, h HandlerFunc, mw ...HandlerFunc) {
	e.addRoute(http.MethodPatch, path, h, mw...)
}

// DELETE registers a DELETE route.
func (e *Engine) DELETE(path string, h HandlerFunc, mw ...HandlerFunc) {
	e.addRoute(http.MethodDelete, path, h, mw...)
}

// OPTIONS registers an OPTIONS route.
func (e *Engine) OPTIONS(path string, h HandlerFunc, mw ...HandlerFunc) {
	e.addRoute(http.MethodOptions, path, h, mw...)
}

// HEAD registers a HEAD route.
func (e *Engine) HEAD(path string, h HandlerFunc, mw ...HandlerFunc) {
	e.addRoute(http.MethodHead, path, h, mw...)
}

// Any registers a handler for all HTTP methods.
func (e *Engine) Any(path string, h HandlerFunc, mw ...HandlerFunc) {
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodOptions, http.MethodHead}
	for _, m := range methods {
		e.addRoute(m, path, h, mw...)
	}
}

// addRoute registers a route with the router, adapting the handler type.
func (e *Engine) addRoute(method, path string, h HandlerFunc, mw ...HandlerFunc) {
	// Adapt core.HandlerFunc → router.HandlerFunc
	routerMW := make([]router.HandlerFunc, len(mw))
	for i, m := range mw {
		m := m
		routerMW[i] = func(ctx router.ContextParamSetter) error {
			return m(ctx.(*rudraContext.Context))
		}
	}

	routerHandler := func(ctx router.ContextParamSetter) error {
		return h(ctx.(*rudraContext.Context))
	}

	e.router.Add(method, path, routerHandler, routerMW...)
}

// Group creates a route group with a prefix and optional middleware.
func (e *Engine) Group(prefix string, mw ...HandlerFunc) *Group {
	routerMW := make([]router.HandlerFunc, len(mw))
	for i, m := range mw {
		m := m
		routerMW[i] = func(ctx router.ContextParamSetter) error {
			return m(ctx.(*rudraContext.Context))
		}
	}

	rg := router.NewGroup(prefix, e.router, routerMW...)
	return &Group{group: rg, engine: e}
}

// SetErrorHandler sets a custom error handler.
func (e *Engine) SetErrorHandler(fn errors.ErrorHandler) {
	e.errorHandler = fn
}

// Run starts the HTTP/1.1 server on the given address.
// Uses a custom TCP listener with SO_REUSEPORT + TCP_NODELAY when enabled.
func (e *Engine) Run(addr string) error {
	ln, err := newTCPListener(addr, e.opts)
	if err != nil {
		return err
	}

	e.mu.Lock()
	e.server = &http.Server{
		Handler:           e,
		ReadTimeout:       e.opts.ReadTimeout,
		WriteTimeout:      e.opts.WriteTimeout,
		IdleTimeout:       e.opts.IdleTimeout,
		ReadHeaderTimeout: e.opts.ReadHeaderTimeout,
		MaxHeaderBytes:    e.opts.MaxHeaderBytes,
	}
	e.mu.Unlock()

	log.Printf("rudra: listening on %s", addr)
	return e.server.Serve(ln)
}

// RunTLS starts the HTTPS server with hardened TLS configuration.
// Cipher suites are explicitly set to AEAD-only (AES-GCM + ChaCha20-Poly1305).
// TLS 1.2 minimum. Session resumption enabled via session tickets.
func (e *Engine) RunTLS(addr, certFile, keyFile string) error {
	ln, err := newTCPListener(addr, e.opts)
	if err != nil {
		return err
	}

	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		SessionTicketsDisabled: false,
	}

	e.mu.Lock()
	e.server = &http.Server{
		Handler:           e,
		TLSConfig:         tlsCfg,
		ReadTimeout:       e.opts.ReadTimeout,
		WriteTimeout:      e.opts.WriteTimeout,
		IdleTimeout:       e.opts.IdleTimeout,
		ReadHeaderTimeout: e.opts.ReadHeaderTimeout,
		MaxHeaderBytes:    e.opts.MaxHeaderBytes,
	}
	e.mu.Unlock()

	log.Printf("rudra: listening on %s (TLS)", addr)
	return e.server.ServeTLS(ln, certFile, keyFile)
}

// RunListener starts the server on a custom net.Listener.
func (e *Engine) RunListener(l net.Listener) error {
	e.mu.Lock()
	e.server = &http.Server{Handler: e}
	e.mu.Unlock()
	return e.server.Serve(l)
}

// Shutdown performs a graceful shutdown, draining active connections
// within the configured ShutdownTimeout. Safe to call multiple times.
func (e *Engine) Shutdown(ctx context.Context) error {
	var err error
	e.once.Do(func() {
		close(e.shutdownCh)
		e.mu.RLock()
		srv := e.server
		e.mu.RUnlock()
		if srv != nil {
			err = srv.Shutdown(ctx)
		}
	})
	return err
}

// Router returns the underlying router for advanced use.
func (e *Engine) Router() *router.Router {
	return e.router
}
