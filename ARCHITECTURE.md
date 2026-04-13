# Rudra (रुद्र) — Architecture Documentation

> **"Fierce. Fast. Fearless."**
> The Unconquerable Go Web Framework by Aarambh Dev Hub

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Design Philosophy](#2-design-philosophy)
3. [Repository Structure](#3-repository-structure)
4. [Core Architecture](#4-core-architecture)
5. [HTTP/1.1 Engine](#5-http11-engine)
6. [HTTP/2 Engine](#6-http2-engine)
7. [WebSocket Engine](#7-websocket-engine)
8. [Server-Sent Events (SSE)](#8-server-sent-events-sse)
9. [Radix Tree Router](#9-radix-tree-router)
10. [Context System](#10-context-system)
11. [Middleware Chain](#11-middleware-chain)
12. [Binding & Validation](#12-binding--validation)
13. [Rendering Engine](#13-rendering-engine)
14. [Error Handling](#14-error-handling)
15. [Configuration System](#15-configuration-system)
16. [Built-in Middleware Catalog](#16-built-in-middleware-catalog)
17. [Graceful Shutdown](#17-graceful-shutdown)
18. [Zero-Allocation Optimizations](#18-zero-allocation-optimizations)
19. [Benchmarking Strategy](#19-benchmarking-strategy)
20. [Security Architecture](#20-security-architecture)
21. [Testing Architecture](#21-testing-architecture)
22. [Feature Comparison Matrix](#22-feature-comparison-matrix)
23. [Internal Data Flow](#23-internal-data-flow)
24. [Memory Model](#24-memory-model)
25. [Future Architecture Targets](#25-future-architecture-targets)

---

## 1. Project Overview

**Rudra** is a high-performance, batteries-included Go web framework built from scratch on top of the Go standard library (`net/http`). It targets zero-allocation hot paths, sub-microsecond routing, and feature parity with — and benchmark superiority over — **Gin**, **Echo**, **Fiber**, and **net/http** vanilla handlers.

| Property         | Value                                      |
|------------------|--------------------------------------------|
| Language         | Go 1.22+                                   |
| Module           | `github.com/AarambhDevHub/rudra`           |
| Base             | `net/http` stdlib + `golang.org/x/net/http2` |
| Router           | Custom zero-alloc radix tree               |
| License          | MIT + Apache 2.0                           |
| Sister Project   | Ajaya (Rust) — AarambhDevHub/ajaya        |
| Tagline          | Fierce. Fast. Fearless.                    |

---

## 2. Design Philosophy

### 2.1 Core Principles

```
┌─────────────────────────────────────────────────────────────────┐
│  1. ZERO ALLOCATION on hot path (routing, context, response)    │
│  2. STDLIB COMPATIBILITY — works with any net/http middleware   │
│  3. BATTERIES INCLUDED — HTTP/1.1, HTTP/2, WS, SSE, gRPC-web  │
│  4. EXPLICIT OVER MAGIC — no reflection in hot paths           │
│  5. COMPOSABLE — every layer is replaceable                     │
│  6. OBSERVABLE — built-in tracing, metrics, structured logging │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Comparison With Existing Frameworks

| Concern                  | Gin        | Echo       | Fiber       | Rudra         |
|--------------------------|------------|------------|-------------|---------------|
| Base                     | net/http   | net/http   | fasthttp    | net/http      |
| stdlib compat            | ✅         | ✅         | ❌          | ✅            |
| HTTP/2 native            | via TLS    | via TLS    | ❌          | ✅ h2c + TLS  |
| WebSocket                | 3rd party  | 3rd party  | built-in    | built-in      |
| SSE                      | manual     | manual     | built-in    | built-in      |
| Zero-alloc routing       | partial    | partial    | ✅          | ✅            |
| Context pooling          | ✅         | ✅         | ✅          | ✅            |
| Radix tree router        | ✅         | ✅         | ✅          | ✅            |
| Validation built-in      | ❌         | ✅         | ❌          | ✅            |
| Graceful shutdown        | manual     | built-in   | built-in    | built-in      |
| gRPC-Web bridge          | ❌         | ❌         | ❌          | ✅ (v0.4.x)   |

---

## 3. Repository Structure

```
rudra/
│
├── rudra.go                    # Public API surface — Engine factory
├── go.mod
├── go.sum
│
├── core/                       # Engine, server lifecycle, config
│   ├── engine.go               # Main Engine struct
│   ├── server.go               # HTTP/1.1 + HTTP/2 server bootstrap
│   ├── options.go              # Functional options pattern
│   └── signals.go              # OS signal handling, graceful shutdown
│
├── router/                     # Radix tree router
│   ├── tree.go                 # Radix tree implementation
│   ├── node.go                 # Tree node with param capture
│   ├── router.go               # Route registration API
│   ├── group.go                # Route groups with prefix + middleware
│   └── params.go               # Zero-alloc param storage
│
├── context/                    # Request/Response context
│   ├── context.go              # RudraContext struct
│   ├── pool.go                 # sync.Pool for context recycling
│   ├── request.go              # Request accessors (headers, body, params)
│   └── response.go             # Response writers (JSON, HTML, Stream)
│
├── middleware/                 # All built-in middleware
│   ├── logger.go               # Structured access logger
│   ├── recovery.go             # Panic recovery
│   ├── cors.go                 # CORS with preflight cache
│   ├── ratelimit.go            # Token bucket rate limiter
│   ├── compress.go             # gzip / brotli / zstd response compression
│   ├── cache.go                # In-memory response cache (LRU)
│   ├── auth/
│   │   ├── jwt.go              # JWT authentication middleware
│   │   ├── basic.go            # HTTP Basic Auth
│   │   └── apikey.go           # API Key Auth
│   ├── timeout.go              # Per-request timeout
│   ├── requestid.go            # X-Request-ID injection
│   ├── secure.go               # Security headers (HSTS, CSP, X-Frame)
│   ├── csrf.go                 # CSRF token generation/validation
│   ├── bodylimit.go            # Request body size limiter
│   ├── etag.go                 # ETag generation for caching
│   ├── redirect.go             # HTTP → HTTPS redirect
│   ├── trace.go                # OpenTelemetry trace injection
│   └── metrics.go              # Prometheus metrics exposition
│
├── binding/                    # Request data binding
│   ├── binder.go               # Binder interface
│   ├── json.go                 # JSON binding (encoding/json + sonic fallback)
│   ├── xml.go                  # XML binding
│   ├── form.go                 # multipart/form-data + url-encoded
│   ├── query.go                # Query string binding
│   ├── path.go                 # Path param binding
│   ├── header.go               # Header binding
│   └── msgpack.go              # MessagePack binding
│
├── render/                     # Response rendering
│   ├── render.go               # Render interface
│   ├── json.go                 # JSON renderer (zero-copy)
│   ├── xml.go                  # XML renderer
│   ├── html.go                 # HTML template renderer
│   ├── text.go                 # Plain text renderer
│   ├── blob.go                 # Binary/file renderer
│   ├── stream.go               # Chunked transfer stream renderer
│   └── msgpack.go              # MessagePack renderer
│
├── ws/                         # WebSocket engine
│   ├── upgrader.go             # HTTP → WS upgrade
│   ├── conn.go                 # WebSocket connection wrapper
│   ├── hub.go                  # Connection hub (broadcast/rooms)
│   ├── message.go              # Message types (Text, Binary, Ping, Pong)
│   └── pool.go                 # Write buffer pool
│
├── sse/                        # Server-Sent Events engine
│   ├── broker.go               # SSE event broker
│   ├── event.go                # Event struct (id, event, data, retry)
│   ├── client.go               # Per-client writer
│   └── stream.go               # Flusher-aware streaming writer
│
├── validator/                  # Input validation
│   ├── validator.go            # Validator interface
│   ├── builtin.go              # Built-in rules (required, min, max, email, url, uuid…)
│   ├── struct.go               # Struct tag-based validation
│   └── custom.go               # Custom rule registration
│
├── config/                     # Configuration
│   ├── config.go               # Config struct with defaults
│   ├── env.go                  # .env / OS env loader
│   └── yaml.go                 # YAML config file loader
│
├── errors/                     # Error handling
│   ├── errors.go               # RudraError type
│   ├── http.go                 # HTTP error constructors (400, 401, 403, 404, 500…)
│   └── handler.go              # Global error handler
│
├── testutil/                   # Testing utilities
│   ├── recorder.go             # Response recorder
│   ├── request.go              # Test request builder
│   └── assert.go               # HTTP assertion helpers
│
├── cmd/
│   └── rudra/                  # CLI scaffolding tool (future)
│
├── examples/
│   ├── hello/                  # Minimal hello world
│   ├── rest-api/               # Full REST API example
│   ├── websocket/              # WebSocket chat example
│   ├── sse/                    # SSE live feed example
│   ├── http2/                  # HTTP/2 push example
│   ├── middleware/             # Custom middleware example
│   └── grpc-web/               # gRPC-Web bridge example
│
├── benchmarks/
│   ├── router_bench_test.go    # Router micro-benchmark
│   ├── context_bench_test.go   # Context alloc benchmark
│   ├── framework_bench_test.go # Cross-framework comparison
│   └── scripts/
│       ├── wrk.sh              # wrk benchmark script
│       └── ab.sh               # Apache Bench script
│
├── docs/
│   ├── ARCHITECTURE.md         # This file
│   ├── ROADMAP.md
│   ├── CONTRIBUTING.md
│   ├── SECURITY.md
│   └── CODE_OF_CONDUCT.md
│
├── CHANGELOG.md
├── README.md
├── LICENSE-MIT
└── LICENSE-APACHE
```

---

## 4. Core Architecture

### 4.1 Engine Lifecycle

```
                        rudra.New()
                            │
                            ▼
                     ┌─────────────┐
                     │   Engine    │◄─── functional options
                     │             │     (WithTLS, WithH2, WithLogger…)
                     └──────┬──────┘
                            │
              ┌─────────────┼──────────────┐
              ▼             ▼              ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐
        │  Router  │  │  Server  │  │  Config  │
        │ (radix)  │  │(h1+h2+ws)│  │  Store   │
        └──────────┘  └──────────┘  └──────────┘
              │             │
              ▼             ▼
        Route Table    net/http.Server
        (compiled)     + h2c Handler
                            │
                ┌───────────┴───────────┐
                ▼                       ▼
          HTTP/1.1 conn           HTTP/2 conn
          (keep-alive)           (multiplexed)
                │                       │
                └───────────┬───────────┘
                            ▼
                    Middleware Chain
                            │
                            ▼
                    Router.ServeHTTP()
                            │
                  ┌─────────┴──────────┐
                  ▼                    ▼
            Static match         Param/Wildcard
            (O(1) map)           (radix tree)
                  │                    │
                  └─────────┬──────────┘
                            ▼
                    Context (from pool)
                            │
                            ▼
                     Handler func
                            │
                  ┌─────────┴──────────┐
                  ▼                    ▼
           Normal response        Upgrade path
           (JSON/HTML/etc)     (WS / SSE / h2push)
                  │
                  ▼
           Context.Release()
           (return to pool)
```

### 4.2 Engine Struct

```go
// core/engine.go

package core

import (
    "context"
    "crypto/tls"
    "net"
    "net/http"
    "sync"
    "time"

    "github.com/AarambhDevHub/rudra/router"
    rudraContext "github.com/AarambhDevHub/rudra/context"
    "github.com/AarambhDevHub/rudra/errors"
    "golang.org/x/net/http2"
    "golang.org/x/net/http2/h2c"
)

// HandlerFunc is the core handler type in Rudra.
// Every route handler and middleware implements this signature.
type HandlerFunc func(*rudraContext.Context) error

// Engine is the heart of Rudra.
// It owns the router, middleware chain, server configuration,
// and the context pool.
type Engine struct {
    // Router holds the radix tree and all registered routes.
    router *router.Router

    // pool recycles Context objects to eliminate per-request allocations.
    pool sync.Pool

    // middleware is the global middleware chain applied before routing.
    middleware []HandlerFunc

    // errorHandler is called when a HandlerFunc returns a non-nil error.
    errorHandler errors.ErrorHandler

    // opts holds all server-level configuration.
    opts *Options

    // server is the underlying net/http server.
    server *http.Server

    // h2server manages the HTTP/2 protocol layer.
    h2server *http2.Server

    // shutdownCh receives OS signals for graceful shutdown.
    shutdownCh chan struct{}

    // once ensures the server is started only once.
    once sync.Once

    mu sync.RWMutex
}

// New creates a new Rudra Engine with default options.
// Use functional options to customize behavior.
//
// Example:
//
//   app := core.New(
//       core.WithReadTimeout(5 * time.Second),
//       core.WithTLS("cert.pem", "key.pem"),
//       core.WithHTTP2(),
//   )
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

    // Context pool: recycle RudraContext objects per-request.
    // This is the single most impactful allocation optimization.
    e.pool = sync.Pool{
        New: func() any {
            return rudraContext.New()
        },
    }

    return e
}

// ServeHTTP implements http.Handler.
// This is the entry point for every incoming HTTP request.
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Acquire a context from the pool — zero allocation.
    c := e.pool.Get().(*rudraContext.Context)
    c.Reset(w, r)

    defer func() {
        // Always release the context back to the pool.
        c.Release()
        e.pool.Put(c)
    }()

    // Run global middleware chain, then route to handler.
    if err := e.handleRequest(c); err != nil {
        e.errorHandler(c, err)
    }
}

// handleRequest applies the middleware chain and dispatches to the matched handler.
func (e *Engine) handleRequest(c *rudraContext.Context) error {
    // Apply global middleware, building a composed handler chain.
    h := e.router.Find(c.Method(), c.Path(), c)

    if h == nil {
        return errors.NotFound("route not found: " + c.Path())
    }

    // Chain global middleware around the matched handler.
    composed := applyMiddleware(h, e.middleware...)
    return composed(c)
}

// applyMiddleware wraps a handler with a slice of middleware in reverse order,
// so the first middleware in the slice is the outermost wrapper.
func applyMiddleware(h HandlerFunc, middleware ...HandlerFunc) HandlerFunc {
    for i := len(middleware) - 1; i >= 0; i-- {
        mw := middleware[i]
        next := h
        h = func(c *rudraContext.Context) error {
            c.SetNext(func() error { return next(c) })
            return mw(c)
        }
    }
    return h
}

// Run starts the HTTP/1.1 server on the given address.
func (e *Engine) Run(addr string) error {
    e.server = &http.Server{
        Addr:              addr,
        Handler:           e,
        ReadTimeout:       e.opts.ReadTimeout,
        WriteTimeout:      e.opts.WriteTimeout,
        IdleTimeout:       e.opts.IdleTimeout,
        ReadHeaderTimeout: e.opts.ReadHeaderTimeout,
        MaxHeaderBytes:    e.opts.MaxHeaderBytes,
    }

    go e.waitForShutdown(e.server)
    return e.server.ListenAndServe()
}

// RunTLS starts the HTTPS server with HTTP/2 support (when enabled).
func (e *Engine) RunTLS(addr, certFile, keyFile string) error {
    tlsCfg := &tls.Config{
        MinVersion:               tls.VersionTLS12,
        PreferServerCipherSuites: true,
        CurvePreferences: []tls.CurveID{
            tls.X25519,
            tls.CurveP256,
        },
    }

    e.server = &http.Server{
        Addr:         addr,
        Handler:      e,
        TLSConfig:    tlsCfg,
        ReadTimeout:  e.opts.ReadTimeout,
        WriteTimeout: e.opts.WriteTimeout,
        IdleTimeout:  e.opts.IdleTimeout,
    }

    // Configure HTTP/2 on the TLS server when enabled.
    if e.opts.HTTP2Enabled {
        e.h2server = &http2.Server{
            MaxHandlers:                  e.opts.HTTP2MaxHandlers,
            MaxConcurrentStreams:          e.opts.HTTP2MaxConcurrentStreams,
            MaxReadFrameSize:             e.opts.HTTP2MaxReadFrameSize,
            IdleTimeout:                  e.opts.HTTP2IdleTimeout,
        }
        if err := http2.ConfigureServer(e.server, e.h2server); err != nil {
            return err
        }
    }

    go e.waitForShutdown(e.server)
    return e.server.ListenAndServeTLS(certFile, keyFile)
}

// RunH2C starts a plaintext HTTP/2 (h2c) server — no TLS required.
// Ideal for internal microservice communication.
func (e *Engine) RunH2C(addr string) error {
    h2s := &http2.Server{
        MaxConcurrentStreams: e.opts.HTTP2MaxConcurrentStreams,
    }

    e.server = &http.Server{
        Addr:        addr,
        Handler:     h2c.NewHandler(e, h2s),
        ReadTimeout: e.opts.ReadTimeout,
        IdleTimeout: e.opts.IdleTimeout,
    }

    go e.waitForShutdown(e.server)
    return e.server.ListenAndServe()
}

// Shutdown performs a graceful shutdown, draining active connections.
func (e *Engine) Shutdown(ctx context.Context) error {
    close(e.shutdownCh)
    return e.server.Shutdown(ctx)
}

func (e *Engine) waitForShutdown(srv *http.Server) {
    <-e.shutdownCh
    ctx, cancel := context.WithTimeout(context.Background(), e.opts.ShutdownTimeout)
    defer cancel()
    _ = srv.Shutdown(ctx)
}

// Listener starts the server on a custom net.Listener.
// Useful for testing or Unix domain sockets.
func (e *Engine) RunListener(l net.Listener) error {
    e.server = &http.Server{Handler: e}
    return e.server.Serve(l)
}
```

---

## 5. HTTP/1.1 Engine

### 5.1 Connection Handling

Rudra's HTTP/1.1 layer uses Go's standard `net/http` server with aggressive tuning:

```go
// core/options.go

package core

import "time"

type Options struct {
    // Timeouts — tuned for low-latency APIs
    ReadTimeout       time.Duration // default: 5s
    WriteTimeout      time.Duration // default: 10s
    IdleTimeout       time.Duration // default: 120s  (keep-alive)
    ReadHeaderTimeout time.Duration // default: 2s
    ShutdownTimeout   time.Duration // default: 30s

    // Limits
    MaxHeaderBytes    int           // default: 1MB
    MaxBodyBytes      int64         // default: 32MB

    // HTTP/2 settings
    HTTP2Enabled              bool
    HTTP2MaxHandlers          int
    HTTP2MaxConcurrentStreams uint32  // default: 250
    HTTP2MaxReadFrameSize     uint32  // default: 1MB
    HTTP2IdleTimeout          time.Duration

    // TLS
    TLSCertFile string
    TLSKeyFile  string

    // Behavior
    StrictRouting   bool  // /foo ≠ /foo/
    CaseSensitive   bool  // /Foo ≠ /foo
    UnescapeParams  bool  // URL-decode path params
    TrustProxyIPs   []string
}

func defaultOptions() *Options {
    return &Options{
        ReadTimeout:               5 * time.Second,
        WriteTimeout:              10 * time.Second,
        IdleTimeout:               120 * time.Second,
        ReadHeaderTimeout:         2 * time.Second,
        ShutdownTimeout:           30 * time.Second,
        MaxHeaderBytes:            1 << 20,  // 1MB
        MaxBodyBytes:              32 << 20, // 32MB
        HTTP2MaxConcurrentStreams: 250,
        HTTP2MaxReadFrameSize:     1 << 20,
        HTTP2IdleTimeout:          90 * time.Second,
        StrictRouting:             false,
        CaseSensitive:             false,
    }
}
```

### 5.2 Keep-Alive Optimization

The server is configured with 120s idle timeout. Go's `net/http` manages connection pooling internally. Rudra extends this with:

- `SO_REUSEPORT` socket option (Linux) for multi-core listener distribution
- `TCP_NODELAY` enabled (disables Nagle, reduces latency for small responses)
- `TCP_FASTOPEN` support on Linux 3.7+ for 0-RTT connection establishment

```go
// core/server.go — custom listener with socket tuning

import (
    "net"
    "syscall"
)

func newTCPListener(addr string) (net.Listener, error) {
    lc := net.ListenConfig{
        Control: func(network, address string, c syscall.RawConn) error {
            return c.Control(func(fd uintptr) {
                // TCP_NODELAY — disable Nagle algorithm
                syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
                // SO_REUSEPORT — allow multiple listeners on same port
                syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
            })
        },
    }
    return lc.Listen(context.Background(), "tcp", addr)
}
```

---

## 6. HTTP/2 Engine

### 6.1 HTTP/2 Feature Set

Rudra supports the full HTTP/2 feature set via `golang.org/x/net/http2`:

| Feature                   | Support |
|---------------------------|---------|
| Multiplexed streams       | ✅      |
| Header compression (HPACK)| ✅      |
| Server push               | ✅      |
| Flow control              | ✅      |
| Stream prioritization     | ✅      |
| Plaintext h2c             | ✅      |
| TLS ALPN negotiation      | ✅      |
| GOAWAY / RST_STREAM       | ✅      |

### 6.2 Server Push

```go
// context/response.go

// Push proactively sends a resource to the client over HTTP/2.
// Falls back silently on HTTP/1.1 connections.
func (c *Context) Push(target string, opts *http.PushOptions) error {
    pusher, ok := c.writer.(http.Pusher)
    if !ok {
        return nil // HTTP/1.1 — silently skip, not an error
    }
    if opts == nil {
        opts = &http.PushOptions{
            Header: http.Header{
                "Content-Type": []string{"text/css"},
            },
        }
    }
    return pusher.Push(target, opts)
}

// Example handler using HTTP/2 push:
func handler(c *rudraContext.Context) error {
    // Push stylesheet before sending HTML
    _ = c.Push("/static/style.css", nil)
    _ = c.Push("/static/app.js", nil)
    return c.HTML(200, "<html>...</html>")
}
```

### 6.3 h2c (Plaintext HTTP/2)

```go
// Useful for internal gRPC-web and microservice traffic
// No TLS overhead — high throughput on private networks

app := rudra.New(rudra.WithH2C())
app.GET("/grpc-ping", grpcWebHandler)
app.RunH2C(":8080")
```

---

## 7. WebSocket Engine

### 7.1 Architecture

```
Client                    Rudra Engine                      Hub
  │                            │                             │
  │──── HTTP GET /ws ─────────►│                             │
  │     Upgrade: websocket     │                             │
  │◄─── 101 Switching Proto ───│                             │
  │                            │                             │
  │                     ws.Conn created                      │
  │                     registered in Hub ──────────────────►│
  │                            │                             │
  │──── Text/Binary frame ────►│                             │
  │                     OnMessage handler                    │
  │                            │── Broadcast ───────────────►│
  │                            │                     distribute to all
  │◄─── frames ────────────────│◄────────────────────────────│
  │                            │                             │
  │──── Close frame ──────────►│                             │
  │                     deregister from Hub ────────────────►│
```

### 7.2 WebSocket Upgrader

```go
// ws/upgrader.go

package ws

import (
    "bufio"
    "crypto/sha1"
    "encoding/base64"
    "net"
    "net/http"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// Upgrader handles the HTTP → WebSocket protocol upgrade.
type Upgrader struct {
    // ReadBufferSize and WriteBufferSize specify the I/O buffer sizes in bytes.
    // Larger values reduce syscall overhead at the cost of memory.
    ReadBufferSize  int // default: 4096
    WriteBufferSize int // default: 4096

    // CheckOrigin returns true if the request origin is acceptable.
    // If nil, all origins are accepted (not safe for production).
    CheckOrigin func(r *http.Request) bool

    // EnableCompression enables per-message deflate compression (RFC 7692).
    EnableCompression bool

    // HandshakeTimeout limits the duration of the handshake.
    HandshakeTimeout time.Duration

    // WritePool holds reusable write buffers to reduce allocations.
    writePool sync.Pool
}

// NewUpgrader returns a production-ready Upgrader with sane defaults.
func NewUpgrader() *Upgrader {
    u := &Upgrader{
        ReadBufferSize:  4096,
        WriteBufferSize: 4096,
        CheckOrigin:     func(r *http.Request) bool { return true },
    }
    u.writePool = sync.Pool{
        New: func() any {
            buf := make([]byte, u.WriteBufferSize)
            return &buf
        },
    }
    return u
}

// Upgrade performs the WebSocket handshake.
// Returns a Conn that is safe for concurrent reads and writes.
func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request) (*Conn, error) {
    if r.Method != http.MethodGet {
        return nil, ErrMethodNotAllowed
    }

    if !u.CheckOrigin(r) {
        return nil, ErrForbidden
    }

    key := r.Header.Get("Sec-WebSocket-Key")
    if key == "" {
        return nil, ErrBadHandshake
    }

    // Compute the accept key per RFC 6455 §4.2.2
    h := sha1.New()
    h.Write([]byte(key + wsGUID))
    acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

    // Hijack the connection to take over raw TCP control.
    hj, ok := w.(http.Hijacker)
    if !ok {
        return nil, ErrHijackNotSupported
    }

    netConn, brw, err := hj.Hijack()
    if err != nil {
        return nil, err
    }

    // Send 101 Switching Protocols response.
    resp := "HTTP/1.1 101 Switching Protocols\r\n" +
        "Upgrade: websocket\r\n" +
        "Connection: Upgrade\r\n" +
        "Sec-WebSocket-Accept: " + acceptKey + "\r\n"

    if u.EnableCompression {
        resp += "Sec-WebSocket-Extensions: permessage-deflate; server_no_context_takeover\r\n"
    }

    resp += "\r\n"

    if _, err = brw.WriteString(resp); err != nil {
        netConn.Close()
        return nil, err
    }
    if err = brw.Flush(); err != nil {
        netConn.Close()
        return nil, err
    }

    return newConn(netConn, brw, u), nil
}
```

### 7.3 WebSocket Connection

```go
// ws/conn.go

// Conn represents a live WebSocket connection.
// Read and Write operations are each protected by a dedicated mutex,
// allowing one concurrent reader and one concurrent writer.
type Conn struct {
    conn        net.Conn
    br          *bufio.Reader
    writeMu     sync.Mutex
    readMu      sync.Mutex
    closeOnce   sync.Once
    closed      chan struct{}
    readTimeout time.Duration
    compression bool
}

// ReadMessage blocks until a complete WebSocket frame is received.
// Returns the message type (TextMessage / BinaryMessage) and payload.
// Zero-copy path: returns a slice pointing into the internal read buffer
// when the payload fits within ReadBufferSize.
func (c *Conn) ReadMessage() (MessageType, []byte, error) {
    c.readMu.Lock()
    defer c.readMu.Unlock()
    return c.readFrame()
}

// WriteMessage sends a WebSocket frame of the given type.
// Acquires the write mutex — safe to call from multiple goroutines.
func (c *Conn) WriteMessage(t MessageType, data []byte) error {
    c.writeMu.Lock()
    defer c.writeMu.Unlock()
    return c.writeFrame(t, data)
}

// WriteJSON marshals v to JSON and sends it as a TextMessage frame.
func (c *Conn) WriteJSON(v any) error {
    data, err := json.Marshal(v)
    if err != nil {
        return err
    }
    return c.WriteMessage(TextMessage, data)
}

// Close sends a CloseNormalClosure frame and closes the TCP connection.
func (c *Conn) Close() error {
    var err error
    c.closeOnce.Do(func() {
        close(c.closed)
        _ = c.WriteMessage(CloseMessage, FormatCloseMessage(CloseNormalClosure, ""))
        err = c.conn.Close()
    })
    return err
}
```

### 7.4 WebSocket Hub (Rooms + Broadcast)

```go
// ws/hub.go

// Hub manages all active WebSocket connections.
// Supports rooms (named groups) and global broadcast.
// All operations are goroutine-safe via channel-based dispatch.
type Hub struct {
    // clients maps client ID → *Conn
    clients map[string]*Conn

    // rooms maps room name → set of client IDs
    rooms map[string]map[string]struct{}

    // broadcast delivers a message to every connected client
    broadcast chan *HubMessage

    // roomcast delivers a message to all clients in a room
    roomcast chan *RoomMessage

    // register adds a new client to the hub
    register chan *ClientRegistration

    // unregister removes a client from all rooms and the hub
    unregister chan string

    mu sync.RWMutex
}

// Broadcast sends a message to all connected clients concurrently.
func (h *Hub) Broadcast(msgType MessageType, data []byte) {
    h.broadcast <- &HubMessage{Type: msgType, Data: data}
}

// BroadcastToRoom sends a message to all clients in the named room.
func (h *Hub) BroadcastToRoom(room string, msgType MessageType, data []byte) {
    h.roomcast <- &RoomMessage{Room: room, Type: msgType, Data: data}
}

// JoinRoom adds a client to a named room.
func (h *Hub) JoinRoom(clientID, room string) { ... }

// LeaveRoom removes a client from a room.
func (h *Hub) LeaveRoom(clientID, room string) { ... }

// Run starts the hub's event loop. Call in a goroutine.
func (h *Hub) Run() {
    for {
        select {
        case reg := <-h.register:
            h.clients[reg.ID] = reg.Conn

        case id := <-h.unregister:
            if conn, ok := h.clients[id]; ok {
                conn.Close()
                delete(h.clients, id)
                // Remove from all rooms
                for _, members := range h.rooms {
                    delete(members, id)
                }
            }

        case msg := <-h.broadcast:
            for _, conn := range h.clients {
                go conn.WriteMessage(msg.Type, msg.Data)
            }

        case msg := <-h.roomcast:
            if members, ok := h.rooms[msg.Room]; ok {
                for id := range members {
                    if conn, ok := h.clients[id]; ok {
                        go conn.WriteMessage(msg.Type, msg.Data)
                    }
                }
            }
        }
    }
}
```

---

## 8. Server-Sent Events (SSE)

### 8.1 Architecture

SSE uses chunked transfer encoding over a persistent HTTP connection. No upgrade — pure HTTP/1.1 or HTTP/2.

```
Client                    Rudra SSE Broker
  │                            │
  │──── GET /events ──────────►│
  │◄─── 200 text/event-stream ─│
  │                            │
  │◄─── data: {"t": 1} ────────│  (event pushed)
  │◄─── data: {"t": 2} ────────│
  │◄─── id: 42 ────────────────│
  │◄─── event: alert ──────────│
  │◄─── retry: 3000 ───────────│
  │                            │
  │──── Connection close ──────│  (client disconnects)
  │                     client unregistered
```

### 8.2 SSE Broker

```go
// sse/broker.go

package sse

import (
    "fmt"
    "net/http"
    "sync"
    "time"
)

// Event represents a single SSE event following the spec.
// All fields are optional except Data.
type Event struct {
    ID      string // event id (for Last-Event-ID reconnect)
    Event   string // event type name (default: "message")
    Data    string // event payload (can be multi-line)
    Retry   int    // client reconnect interval in ms (0 = use default)
    Comment string // comment line (starts with :)
}

// String serializes the Event to the SSE wire format.
func (e *Event) String() string {
    var buf strings.Builder
    if e.Comment != "" {
        buf.WriteString(": " + e.Comment + "\n")
    }
    if e.ID != "" {
        buf.WriteString("id: " + e.ID + "\n")
    }
    if e.Event != "" {
        buf.WriteString("event: " + e.Event + "\n")
    }
    // Multi-line data: each line gets its own "data: " prefix
    for _, line := range strings.Split(e.Data, "\n") {
        buf.WriteString("data: " + line + "\n")
    }
    if e.Retry > 0 {
        buf.WriteString(fmt.Sprintf("retry: %d\n", e.Retry))
    }
    buf.WriteString("\n") // blank line = end of event
    return buf.String()
}

// Broker manages all SSE client connections and event distribution.
type Broker struct {
    // clients maps client ID → send channel
    clients   map[string]chan *Event
    mu        sync.RWMutex

    // publish receives events from the application layer
    publish   chan *Event

    // closed signals broker shutdown
    closed    chan struct{}

    // bufSize is the per-client channel buffer depth
    bufSize   int
}

// NewBroker creates a Broker and starts its event loop.
func NewBroker(bufSize int) *Broker {
    b := &Broker{
        clients: make(map[string]chan *Event),
        publish: make(chan *Event, 256),
        closed:  make(chan struct{}),
        bufSize: bufSize,
    }
    go b.run()
    return b
}

// Publish sends an event to all connected SSE clients.
func (b *Broker) Publish(e *Event) {
    b.publish <- e
}

// ServeHTTP makes Broker implement http.Handler.
// Each request is held open and receives a stream of events.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // SSE requires the ResponseWriter to support Flusher.
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "SSE not supported", http.StatusInternalServerError)
        return
    }

    // Set SSE headers.
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

    // Support Last-Event-ID for automatic reconnect.
    lastID := r.Header.Get("Last-Event-ID")
    _ = lastID // replay missed events in production implementation

    // Register this client with the broker.
    clientID := generateClientID()
    ch := make(chan *Event, b.bufSize)

    b.mu.Lock()
    b.clients[clientID] = ch
    b.mu.Unlock()

    defer func() {
        b.mu.Lock()
        delete(b.clients, clientID)
        close(ch)
        b.mu.Unlock()
    }()

    // Heartbeat ticker — prevents proxy timeout disconnects.
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case event, ok := <-ch:
            if !ok {
                return
            }
            fmt.Fprint(w, event.String())
            flusher.Flush()

        case <-ticker.C:
            // Send a comment as heartbeat — invisible to application.
            fmt.Fprint(w, ": heartbeat\n\n")
            flusher.Flush()

        case <-r.Context().Done():
            // Client disconnected.
            return

        case <-b.closed:
            return
        }
    }
}

func (b *Broker) run() {
    for {
        select {
        case event := <-b.publish:
            b.mu.RLock()
            for _, ch := range b.clients {
                // Non-blocking send: drop event if client buffer is full.
                // Prevents one slow client from blocking all others.
                select {
                case ch <- event:
                default:
                }
            }
            b.mu.RUnlock()

        case <-b.closed:
            return
        }
    }
}
```

---

## 9. Radix Tree Router

### 9.1 Design Goals

- **O(log n)** worst-case route matching (not O(n) linear scan like Gin's older versions)
- **Zero heap allocations** during routing (params stored in pre-allocated array on Context)
- Support for `:param`, `*wildcard`, and static segments
- Method-level trees (one radix tree per HTTP method)
- Named route registry for URL generation (`router.URL("user.profile", 42)`)

### 9.2 Tree Structure

```go
// router/tree.go

package router

// node represents a single node in the radix tree.
type node struct {
    // path is the static prefix stored at this node.
    // For param nodes, path = ":paramName".
    // For wildcard nodes, path = "*name".
    path string

    // children holds all child nodes, sorted by first byte for O(1) prefix lookup.
    children []*node

    // handler is the HandlerFunc registered at this node.
    // nil if this is an intermediate (non-terminal) node.
    handler HandlerFunc

    // isParam is true if this node captures a path parameter.
    isParam bool

    // isWildcard is true if this node captures a wildcard (rest of path).
    isWildcard bool

    // paramName is the name of the captured parameter (without leading colon).
    paramName string

    // middleware holds route-level middleware applied before handler.
    middleware []HandlerFunc

    // indices is a fast lookup: maps first byte of each child's path to child index.
    // Avoids iterating all children on every character comparison.
    indices []byte
}

// Router holds one radix tree per HTTP method.
type Router struct {
    // trees maps HTTP method (GET, POST, ...) to its radix tree root.
    trees map[string]*node

    // namedRoutes maps route name → pattern for URL generation.
    namedRoutes map[string]string

    mu sync.RWMutex
}

// Find matches the given method and path against the radix tree.
// Path parameters are stored directly into the Context's pre-allocated params slice.
// Returns nil if no match is found.
func (r *Router) Find(method, path string, c *rudraContext.Context) HandlerFunc {
    root, ok := r.trees[method]
    if !ok {
        return nil
    }
    return root.search(path, c)
}

// search traverses the tree for the given path, populating params into the context.
func (n *node) search(path string, c *rudraContext.Context) HandlerFunc {
    if n.isWildcard {
        // Wildcard captures the entire remaining path.
        c.SetParam(n.paramName, path)
        return n.handler
    }

    if n.isParam {
        // Param node: capture until next '/' or end of path.
        end := strings.IndexByte(path, '/')
        if end == -1 {
            end = len(path)
        }
        c.SetParam(n.paramName, path[:end])
        path = path[end:]
        if path == "" {
            return n.handler
        }
    } else {
        // Static node: path must start with node.path.
        if !strings.HasPrefix(path, n.path) {
            return nil
        }
        path = path[len(n.path):]
        if path == "" {
            return n.handler
        }
    }

    // Look up the child by the next byte.
    if len(n.indices) > 0 {
        c := path[0]
        for i, idx := range n.indices {
            if idx == c {
                if h := n.children[i].search(path, ctx); h != nil {
                    return h
                }
            }
        }
    }

    // Try param or wildcard children last.
    for _, child := range n.children {
        if child.isParam || child.isWildcard {
            if h := child.search(path, ctx); h != nil {
                return h
            }
        }
    }

    return nil
}
```

### 9.3 Route Groups

```go
// router/group.go

// Group represents a set of routes sharing a common prefix and middleware.
type Group struct {
    prefix     string
    middleware []HandlerFunc
    router     *Router
}

// Group creates a sub-group with an additional prefix and optional middleware.
func (g *Group) Group(prefix string, middleware ...HandlerFunc) *Group {
    return &Group{
        prefix:     g.prefix + prefix,
        middleware: append(g.middleware, middleware...),
        router:     g.router,
    }
}

// GET registers a GET handler on this group.
func (g *Group) GET(path string, h HandlerFunc, mw ...HandlerFunc) {
    g.router.add(http.MethodGet, g.prefix+path, h, append(g.middleware, mw...)...)
}
// POST, PUT, PATCH, DELETE, OPTIONS, HEAD, CONNECT, TRACE follow the same pattern.
```

---

## 10. Context System

### 10.1 Context Design

The `Context` is the most performance-critical struct in Rudra. Every request gets one, and it must be as cheap as possible to acquire, use, and release.

```go
// context/context.go

package context

import (
    "net/http"
    "sync"
)

// maxParams is the maximum number of URL parameters supported.
// Stored as a fixed-size array on the Context to avoid heap allocation.
const maxParams = 16

// Param is a single key-value pair from the URL path.
type Param struct {
    Key   string
    Value string
}

// Context is Rudra's per-request state container.
// It is pooled via sync.Pool — never create one directly.
type Context struct {
    // writer is the underlying http.ResponseWriter.
    writer  http.ResponseWriter

    // request is the incoming *http.Request.
    request *http.Request

    // params holds zero-allocation URL path parameters.
    // Pre-allocated array — no heap allocation for ≤16 params.
    params     [maxParams]Param
    paramCount int

    // store is a per-request key-value store for middleware communication.
    // Lazily initialized to avoid allocation when unused.
    store map[string]any

    // statusCode tracks the HTTP status set for this response.
    statusCode int

    // written tracks whether the response has been written.
    written bool

    // next is the next handler in the middleware chain.
    next func() error

    // requestID is the X-Request-ID for this request.
    requestID string

    // body is the lazily-read request body bytes.
    body []byte

    // errors accumulates non-fatal errors during handler execution.
    errors []error

    // index tracks position in a locally-composed middleware chain.
    index int8

    // aborted signals that the handler chain should stop.
    aborted bool
}

// Reset resets the Context for reuse from the pool.
// Called at the start of every request — must be fast.
func (c *Context) Reset(w http.ResponseWriter, r *http.Request) {
    c.writer     = w
    c.request    = r
    c.paramCount = 0
    c.statusCode = 200
    c.written    = false
    c.aborted    = false
    c.index      = 0
    c.next       = nil
    c.requestID  = ""
    c.body       = c.body[:0]  // reuse backing array
    c.errors     = c.errors[:0]

    // Reuse the store map if it exists; clear its contents.
    if c.store != nil {
        for k := range c.store {
            delete(c.store, k)
        }
    }
}

// Release clears sensitive fields before returning to the pool.
func (c *Context) Release() {
    c.writer  = nil
    c.request = nil
}

// SetParam adds a URL path parameter. Called by the router during matching.
// Uses a fixed-size pre-allocated array — zero heap allocation.
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

// JSON writes a JSON response with the given status code.
// Uses a pooled encoder for zero-allocation serialization.
func (c *Context) JSON(code int, v any) error {
    return render.JSON(c.writer, code, v)
}

// String writes a plain text response.
func (c *Context) String(code int, s string) error {
    return render.Text(c.writer, code, s)
}

// HTML writes an HTML response.
func (c *Context) HTML(code int, html string) error {
    return render.HTML(c.writer, code, html)
}

// NoContent sends a 204 No Content response.
func (c *Context) NoContent() error {
    c.writer.WriteHeader(http.StatusNoContent)
    return nil
}

// Redirect sends a redirect response.
func (c *Context) Redirect(code int, url string) error {
    http.Redirect(c.writer, c.request, url, code)
    return nil
}

// Abort stops the middleware chain from continuing.
func (c *Context) Abort() {
    c.aborted = true
}

// AbortWithError stops the chain and returns an error.
func (c *Context) AbortWithError(code int, err error) error {
    c.Abort()
    return errors.NewHTTPError(code, err.Error())
}

// Next calls the next handler in the middleware chain.
func (c *Context) Next() error {
    if c.next != nil && !c.aborted {
        return c.next()
    }
    return nil
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
```

---

## 11. Middleware Chain

### 11.1 Chain Execution Model

Rudra uses the **onion model**: each middleware wraps the next. Middleware calls `c.Next()` to invoke the inner layer.

```
Request
   │
   ▼
┌──────────────────────────────────────────┐
│  Logger (start timer)                    │
│  ┌────────────────────────────────────┐  │
│  │  Recovery (defer panic catch)      │  │
│  │  ┌──────────────────────────────┐  │  │
│  │  │  Auth (verify JWT)           │  │  │
│  │  │  ┌────────────────────────┐  │  │  │
│  │  │  │  RateLimit (check)     │  │  │  │
│  │  │  │  ┌──────────────────┐  │  │  │  │
│  │  │  │  │  Handler (exec)  │  │  │  │  │
│  │  │  │  └──────────────────┘  │  │  │  │
│  │  │  └────────────────────────┘  │  │  │
│  │  └──────────────────────────────┘  │  │
│  └────────────────────────────────────┘  │
│  Logger (log duration + status)          │
└──────────────────────────────────────────┘
   │
   ▼
Response
```

### 11.2 Writing Custom Middleware

```go
// A Rudra middleware wraps the next handler using c.Next().
func MyMiddleware() core.HandlerFunc {
    return func(c *context.Context) error {
        // Before handler
        start := time.Now()

        // Call the next layer
        if err := c.Next(); err != nil {
            return err
        }

        // After handler
        log.Printf("%s %s %v", c.Method(), c.Path(), time.Since(start))
        return nil
    }
}

// Register globally
app.Use(MyMiddleware())

// Register on a group
api := app.Group("/api", MyMiddleware())
```

---

## 12. Binding & Validation

### 12.1 Binding

```go
// binding/binder.go

// Binder is the interface for all request data binders.
type Binder interface {
    Bind(c *context.Context, v any) error
}

// BindJSON reads and decodes the request body as JSON into v.
// Uses sonic (bytedance) when available, falls back to encoding/json.
func BindJSON(c *context.Context, v any) error {
    body, err := io.ReadAll(io.LimitReader(c.Request().Body, c.Engine().MaxBodyBytes()))
    if err != nil {
        return errors.BadRequest(err.Error())
    }
    return jsonBinding{}.bind(body, v)
}

// ShouldBind automatically selects the binder based on Content-Type.
func (c *Context) ShouldBind(v any) error {
    switch c.ContentType() {
    case "application/json":
        return BindJSON(c, v)
    case "application/xml", "text/xml":
        return BindXML(c, v)
    case "application/x-www-form-urlencoded":
        return BindForm(c, v)
    case "multipart/form-data":
        return BindMultipart(c, v)
    case "application/msgpack":
        return BindMsgpack(c, v)
    default:
        return BindJSON(c, v)
    }
}

// MustBind binds and validates. Aborts with 400 on failure.
func (c *Context) MustBind(v any) error {
    if err := c.ShouldBind(v); err != nil {
        return c.AbortWithError(http.StatusBadRequest, err)
    }
    if err := c.Validate(v); err != nil {
        return c.AbortWithError(http.StatusUnprocessableEntity, err)
    }
    return nil
}
```

### 12.2 Validation

```go
// validator/builtin.go

// Struct-tag based validation using `rudra:"..."` tags.
//
// Supported tags:
//   required           — field must not be zero value
//   min=N              — minimum length (string) / value (number)
//   max=N              — maximum length (string) / value (number)
//   email              — valid email format
//   url                — valid URL
//   uuid               — valid UUID v4
//   len=N              — exact length
//   oneof=a b c        — value must be one of the listed values
//   alphanum           — only alphanumeric characters
//   numeric            — only numeric characters
//   regexp=pattern     — match regular expression

type User struct {
    Name     string `json:"name"  rudra:"required,min=2,max=64"`
    Email    string `json:"email" rudra:"required,email"`
    Age      int    `json:"age"   rudra:"required,min=18,max=120"`
    Role     string `json:"role"  rudra:"required,oneof=admin user guest"`
}

// Custom validator registration:
validator.Register("indianphone", func(v string) bool {
    matched, _ := regexp.MatchString(`^[6-9]\d{9}$`, v)
    return matched
})
```

---

## 13. Rendering Engine

### 13.1 JSON Renderer (Zero-Copy)

```go
// render/json.go

package render

import (
    "net/http"
    "sync"
    "github.com/bytedance/sonic"  // drop-in encoding/json replacement, 2-4x faster
)

var encoderPool = sync.Pool{
    New: func() any { return sonic.ConfigFastest.NewEncoder(nil) },
}

// JSON writes a JSON response.
// Uses sonic for SIMD-accelerated encoding.
// Writes directly to the ResponseWriter — zero intermediate buffer.
func JSON(w http.ResponseWriter, code int, v any) error {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    enc := sonic.ConfigFastest.NewEncoder(w)
    if err := enc.Encode(v); err != nil {
        return err
    }
    return nil
}
```

### 13.2 Streaming Renderer

```go
// render/stream.go

// Stream writes chunked data incrementally.
// Used for large file downloads, live data feeds, and NDJSON streams.
func Stream(w http.ResponseWriter, code int, contentType string, fn func(w io.Writer) error) error {
    flusher, ok := w.(http.Flusher)
    if !ok {
        return errors.New("streaming not supported")
    }

    w.Header().Set("Content-Type", contentType)
    w.Header().Set("Transfer-Encoding", "chunked")
    w.WriteHeader(code)

    return fn(&flushWriter{w: w, flusher: flusher})
}

type flushWriter struct {
    w       http.ResponseWriter
    flusher http.Flusher
}

func (fw *flushWriter) Write(p []byte) (int, error) {
    n, err := fw.w.Write(p)
    fw.flusher.Flush()
    return n, err
}
```

---

## 14. Error Handling

### 14.1 Error Types

```go
// errors/errors.go

// RudraError is the standard error type for all HTTP errors in Rudra.
// It carries both a machine-readable code and a human-readable message.
type RudraError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Detail  any    `json:"detail,omitempty"`
    Cause   error  `json:"-"` // internal cause, not exposed to clients
}

func (e *RudraError) Error() string {
    return fmt.Sprintf("rudra error %d: %s", e.Code, e.Message)
}

// Constructor helpers
func BadRequest(msg string, detail ...any) *RudraError    { return newErr(400, msg, detail...) }
func Unauthorized(msg string, detail ...any) *RudraError  { return newErr(401, msg, detail...) }
func Forbidden(msg string, detail ...any) *RudraError     { return newErr(403, msg, detail...) }
func NotFound(msg string, detail ...any) *RudraError      { return newErr(404, msg, detail...) }
func Conflict(msg string, detail ...any) *RudraError      { return newErr(409, msg, detail...) }
func UnprocessableEntity(msg string, detail ...any) *RudraError { return newErr(422, msg, detail...) }
func TooManyRequests(msg string, detail ...any) *RudraError     { return newErr(429, msg, detail...) }
func InternalServerError(msg string, detail ...any) *RudraError { return newErr(500, msg, detail...) }
```

### 14.2 Global Error Handler

```go
// errors/handler.go

// DefaultErrorHandler is used when no custom error handler is set.
// Sends a JSON error response and logs the error.
func DefaultErrorHandler(c *context.Context, err error) {
    var re *RudraError
    if errors.As(err, &re) {
        _ = c.JSON(re.Code, re)
        return
    }
    // Unknown error — don't leak internals to client
    _ = c.JSON(http.StatusInternalServerError, &RudraError{
        Code:    500,
        Message: "internal server error",
    })
}

// Usage: custom error handler
app.SetErrorHandler(func(c *context.Context, err error) {
    // Custom structured error response
    c.JSON(500, map[string]any{"error": err.Error(), "requestId": c.RequestID()})
})
```

---

## 15. Configuration System

```go
// config/config.go

// Config is the top-level application configuration.
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    JWT      JWTConfig      `yaml:"jwt"`
    CORS     CORSConfig     `yaml:"cors"`
    Log      LogConfig      `yaml:"log"`
}

// Load reads config from YAML file and overlays environment variables.
// Environment variables override YAML values.
// Format: RUDRA_SERVER_PORT overrides server.port
func Load(path string) (*Config, error) {
    cfg := &Config{}

    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
        return nil, err
    }

    // Overlay environment variables using reflection.
    overlayEnv(cfg, "RUDRA")
    return cfg, nil
}

// Example config.yaml:
//
// server:
//   port: 8080
//   read_timeout: 5s
//   http2: true
//
// cors:
//   origins: ["https://example.com"]
//   methods: ["GET", "POST", "PUT", "DELETE"]
//
// jwt:
//   secret: ${JWT_SECRET}
//   expiry: 24h
```

---

## 16. Built-in Middleware Catalog

### 16.1 Logger

```go
// Structured access logger with zerolog / slog backend.
// Logs: method, path, status, latency, ip, request_id, user_agent, bytes_out.
app.Use(middleware.Logger(middleware.LoggerConfig{
    Format:     "json",           // "json" | "text" | "common"
    Output:     os.Stdout,
    TimeFormat: time.RFC3339,
    SkipPaths:  []string{"/health", "/metrics"},
}))
```

### 16.2 Recovery

```go
// Catches panics, logs the stack trace, returns 500 to client.
// Never exposes stack traces to clients in production.
app.Use(middleware.Recovery(middleware.RecoveryConfig{
    LogStackTrace: true,
    PrintStack:    false, // never print to client
}))
```

### 16.3 CORS

```go
app.Use(middleware.CORS(middleware.CORSConfig{
    AllowOrigins:     []string{"https://app.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Authorization", "Content-Type"},
    ExposeHeaders:    []string{"X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           86400, // preflight cache: 24h
}))
```

### 16.4 Rate Limiter

```go
// Token bucket rate limiter, keyed by IP or custom key.
app.Use(middleware.RateLimit(middleware.RateLimitConfig{
    Rate:      100,            // 100 requests
    Burst:     20,             // burst of 20
    Window:    time.Minute,
    KeyFunc:   func(c *context.Context) string { return c.RealIP() },
    OnLimit: func(c *context.Context) error {
        return errors.TooManyRequests("slow down")
    },
}))
```

### 16.5 JWT Authentication

```go
app.Use(middleware.JWT(middleware.JWTConfig{
    Secret:     []byte(os.Getenv("JWT_SECRET")),
    SignMethod: jwt.SigningMethodHS256,
    TokenLookup: "header:Authorization", // "cookie:token" | "query:token"
    AuthScheme:  "Bearer",
    Claims:      &CustomClaims{},
    SkipRoutes:  []string{"/auth/login", "/auth/register"},
    ContextKey:  "user", // stored as c.Get("user")
}))
```

### 16.6 Compression

```go
// Supports gzip, brotli (brotli package), and zstd (klauspost/compress).
// Selects algorithm based on Accept-Encoding header.
app.Use(middleware.Compress(middleware.CompressConfig{
    Level:     middleware.BestSpeed,
    MinLength: 1024, // skip compression for small responses
    Algorithms: []string{"br", "zstd", "gzip"}, // preference order
}))
```

### 16.7 Request ID

```go
// Injects X-Request-ID header. Uses UUIDv4 or custom generator.
app.Use(middleware.RequestID(middleware.RequestIDConfig{
    Generator: func() string { return uuid.NewString() },
    Header:    "X-Request-ID",
}))
```

### 16.8 Timeout

```go
// Cancels the request context if the handler exceeds the deadline.
app.Use(middleware.Timeout(middleware.TimeoutConfig{
    Timeout: 30 * time.Second,
    OnTimeout: func(c *context.Context) error {
        return errors.NewHTTPError(http.StatusGatewayTimeout, "request timeout")
    },
}))
```

### 16.9 CSRF

```go
// Double-submit cookie CSRF protection.
app.Use(middleware.CSRF(middleware.CSRFConfig{
    TokenLength: 32,
    CookieName:  "_csrf",
    HeaderName:  "X-CSRF-Token",
    Secure:      true,
    SameSite:    http.SameSiteStrictMode,
}))
```

### 16.10 Security Headers

```go
// Sets: HSTS, X-Frame-Options, X-Content-Type-Options, CSP, Referrer-Policy.
app.Use(middleware.Secure(middleware.SecureConfig{
    HSTS:               true,
    HSTSMaxAge:         31536000,
    FrameDeny:          true,
    ContentTypeNosniff: true,
    CSP:                "default-src 'self'",
    ReferrerPolicy:     "strict-origin-when-cross-origin",
}))
```

---

## 17. Graceful Shutdown

```go
// core/signals.go

// ListenForShutdown blocks until SIGINT or SIGTERM is received,
// then performs a graceful drain of active connections within the timeout.
func (e *Engine) ListenForShutdown() {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("rudra: shutdown signal received, draining connections...")

    ctx, cancel := context.WithTimeout(context.Background(), e.opts.ShutdownTimeout)
    defer cancel()

    if err := e.Shutdown(ctx); err != nil {
        log.Printf("rudra: forced shutdown after timeout: %v", err)
    } else {
        log.Println("rudra: clean shutdown complete")
    }
}

// Usage pattern:
func main() {
    app := rudra.New()
    app.GET("/", handler)

    go func() {
        if err := app.Run(":8080"); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    app.ListenForShutdown()
}
```

---

## 18. Zero-Allocation Optimizations

This section documents every deliberate allocation-reduction technique used throughout Rudra. These are the techniques that make the difference between 200k req/sec and 500k req/sec.

### 18.1 Context Pool

```go
// The single biggest win: reuse Context objects across requests.
// Without pooling: 1 alloc per request for the context struct alone.
// With pooling: 0 allocs for the context struct on the hot path.

e.pool = sync.Pool{
    New: func() any { return rudraContext.New() },
}

// Acquire (zero alloc after warmup)
c := e.pool.Get().(*rudraContext.Context)

// Release
e.pool.Put(c)
```

### 18.2 Param Storage (Fixed Array)

```go
// Bad: []Param slice — requires heap allocation for every routed request
type Context struct {
    params []Param // ❌ heap alloc
}

// Good: fixed-size array on the struct — stays on the stack
type Context struct {
    params     [16]Param // ✅ zero alloc for ≤16 params
    paramCount int
}
```

### 18.3 JSON Encoding (sonic + direct writer)

```go
// Bad: encode to intermediate buffer, then copy to ResponseWriter
buf := bytes.NewBuffer(nil)     // alloc
json.NewEncoder(buf).Encode(v)  // alloc
w.Write(buf.Bytes())            // copy

// Good: encode directly to ResponseWriter — zero intermediate buffer
sonic.ConfigFastest.NewEncoder(w).Encode(v) // ✅ zero copy
```

### 18.4 Write Buffer Pool

```go
// For WebSocket writes: pool write buffers instead of allocating per-send.
var writeBufPool = sync.Pool{
    New: func() any {
        buf := make([]byte, 4096)
        return &buf
    },
}

buf := writeBufPool.Get().(*[]byte)
defer writeBufPool.Put(buf)
// use *buf for write operations
```

### 18.5 String ↔ Byte Slice (Unsafe Conversion)

```go
// For zero-copy string-to-bytes conversion on hot paths (read-only use only!).
// Avoids the allocation that `[]byte(str)` triggers.

//go:nosplit
func unsafeBytes(s string) []byte {
    return unsafe.Slice(unsafe.StringData(s), len(s))
}

//go:nosplit
func unsafeString(b []byte) string {
    return unsafe.String(unsafe.SliceData(b), len(b))
}
```

### 18.6 Route Index (Static Routes Map)

```go
// For routes with no params (the majority): bypass the radix tree entirely.
// Store a direct map[string]HandlerFunc for O(1) static lookups.
type Router struct {
    staticRoutes map[string]map[string]HandlerFunc // method → path → handler
    dynamicTree  map[string]*node                  // method → radix tree
}

// Find checks static routes first, then falls back to radix tree.
func (r *Router) Find(method, path string, c *context.Context) HandlerFunc {
    if h, ok := r.staticRoutes[method][path]; ok {
        return h // O(1) map lookup — fastest possible path
    }
    return r.dynamicTree[method].search(path, c)
}
```

### 18.7 Header Canonicalization Cache

```go
// net/http canonicalizes header names on every access (allocates a string).
// Cache frequently-read headers in a sync.Map to amortize cost.

var headerCache sync.Map

func canonicalHeader(key string) string {
    if v, ok := headerCache.Load(key); ok {
        return v.(string)
    }
    canonical := textproto.CanonicalMIMEHeaderKey(key)
    headerCache.Store(key, canonical)
    return canonical
}
```

### 18.8 Allocation Budget Targets

| Operation                    | Target Allocs | Target Bytes |
|------------------------------|---------------|--------------|
| Route match (static)         | 0             | 0 B          |
| Route match (with params)    | 0             | 0 B          |
| Context acquire from pool    | 0             | 0 B          |
| JSON response (≤4KB)         | 1             | ≤ 4 KB       |
| WebSocket frame write        | 0             | 0 B          |
| SSE event write              | 1             | ≤ 256 B      |
| Middleware chain (3 layers)  | 0             | 0 B          |

---

## 19. Benchmarking Strategy

### 19.1 Micro-Benchmarks

```go
// benchmarks/router_bench_test.go

// BenchmarkRouterStatic — pure static route match, no params
func BenchmarkRouterStatic(b *testing.B) {
    r := router.New()
    r.Add(http.MethodGet, "/api/v1/users", nopHandler)

    req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
    c := context.New()

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        h := r.Find(http.MethodGet, "/api/v1/users", c)
        _ = h
    }
}

// BenchmarkRouterParams — route with 3 URL parameters
func BenchmarkRouterParams(b *testing.B) {
    r := router.New()
    r.Add(http.MethodGet, "/api/v1/orgs/:org/repos/:repo/commits/:sha", nopHandler)

    c := context.New()
    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        c.Reset(nil, nil)
        h := r.Find(http.MethodGet, "/api/v1/orgs/rudra/repos/core/commits/abc123", c)
        _ = h
    }
}
```

### 19.2 Framework Comparison Benchmarks (wrk)

```bash
#!/bin/bash
# benchmarks/scripts/wrk.sh

DURATION="30s"
CONNECTIONS=100
THREADS=4

echo "=== Rudra ==="
wrk -t$THREADS -c$CONNECTIONS -d$DURATION http://localhost:8080/bench

echo "=== Gin ==="
wrk -t$THREADS -c$CONNECTIONS -d$DURATION http://localhost:8081/bench

echo "=== Echo ==="
wrk -t$THREADS -c$CONNECTIONS -d$DURATION http://localhost:8082/bench

echo "=== Fiber ==="
wrk -t$THREADS -c$CONNECTIONS -d$DURATION http://localhost:8083/bench

echo "=== net/http stdlib ==="
wrk -t$THREADS -c$CONNECTIONS -d$DURATION http://localhost:8084/bench
```

### 19.3 Target Benchmark Goals

| Framework    | Target (req/sec) | Rudra Goal    |
|-------------|-----------------|---------------|
| net/http     | ~120,000        | +40% above    |
| Gin          | ~160,000        | +25% above    |
| Echo         | ~155,000        | +30% above    |
| Fiber        | ~220,000        | Match or beat |

---

## 20. Security Architecture

### 20.1 Security Checklist

| Threat                     | Mitigation                                     |
|----------------------------|------------------------------------------------|
| SQL Injection              | Not in scope (handled by DB layer)             |
| XSS                        | CSP headers via `middleware.Secure`            |
| CSRF                       | Double-submit cookie via `middleware.CSRF`     |
| Clickjacking               | X-Frame-Options: DENY                          |
| Path Traversal             | URL path sanitization in router                |
| Request Smuggling          | Strict Content-Length enforcement             |
| Slowloris                  | ReadHeaderTimeout + ReadTimeout                |
| ReDoS                      | Validator regexps pre-compiled at startup      |
| Panic → 500 leak           | Recovery middleware hides internals            |
| JWT algorithm confusion    | Explicit algorithm whitelist in JWT middleware |
| Brute force                | Rate limiter per IP                            |
| TLS downgrade              | MinVersion = TLS 1.2, curve preferences set    |
| HSTS missing               | Secure middleware adds HSTS header             |

### 20.2 TLS Hardening

```go
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
    CurvePreferences: []tls.CurveID{
        tls.X25519,   // preferred: fast, secure
        tls.CurveP256,
    },
    CipherSuites: []uint16{
        tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
        tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
    },
    PreferServerCipherSuites: true,
    SessionTicketsDisabled:   false, // enable TLS session resumption
}
```

---

## 21. Testing Architecture

### 21.1 Test Utilities

```go
// testutil/request.go

// TestRequest is a fluent builder for test HTTP requests.
type TestRequest struct {
    method  string
    path    string
    headers map[string]string
    body    io.Reader
    app     *core.Engine
}

func NewTestRequest(app *core.Engine) *TestRequest {
    return &TestRequest{app: app, headers: make(map[string]string)}
}

func (t *TestRequest) GET(path string) *TestRequest {
    t.method, t.path = http.MethodGet, path
    return t
}

func (t *TestRequest) POST(path string, body any) *TestRequest {
    t.method = http.MethodPost
    t.path   = path
    data, _ := json.Marshal(body)
    t.body  = bytes.NewReader(data)
    t.headers["Content-Type"] = "application/json"
    return t
}

func (t *TestRequest) Header(key, value string) *TestRequest {
    t.headers[key] = value
    return t
}

func (t *TestRequest) Do() *TestResponse {
    req := httptest.NewRequest(t.method, t.path, t.body)
    for k, v := range t.headers {
        req.Header.Set(k, v)
    }
    w := httptest.NewRecorder()
    t.app.ServeHTTP(w, req)
    return &TestResponse{recorder: w}
}

// testutil/assert.go

type TestResponse struct {
    recorder *httptest.ResponseRecorder
}

func (r *TestResponse) Status(t *testing.T, code int) *TestResponse {
    t.Helper()
    assert.Equal(t, code, r.recorder.Code)
    return r
}

func (r *TestResponse) JSON(t *testing.T, v any) *TestResponse {
    t.Helper()
    assert.NoError(t, json.Unmarshal(r.recorder.Body.Bytes(), v))
    return r
}

func (r *TestResponse) HasHeader(t *testing.T, key, val string) *TestResponse {
    t.Helper()
    assert.Equal(t, val, r.recorder.Header().Get(key))
    return r
}

// Usage in tests:
func TestUserCreate(t *testing.T) {
    app := setupTestApp()

    var resp UserResponse
    testutil.NewTestRequest(app).
        POST("/api/users", CreateUserRequest{Name: "Arjun", Email: "arjun@example.com"}).
        Header("Authorization", "Bearer "+testJWT).
        Do().
        Status(t, 201).
        JSON(t, &resp)

    assert.Equal(t, "Arjun", resp.Name)
}
```

---

## 22. Feature Comparison Matrix

| Feature                     | Gin  | Echo | Fiber | **Rudra** |
|-----------------------------|------|------|-------|-----------|
| Radix tree router           | ✅   | ✅   | ✅    | ✅        |
| Zero-alloc routing          | ~    | ~    | ✅    | ✅        |
| HTTP/2 native               | ~    | ~    | ❌    | ✅        |
| HTTP/2 server push          | ~    | ~    | ❌    | ✅        |
| h2c (plaintext HTTP/2)      | ❌   | ❌   | ❌    | ✅        |
| WebSocket built-in          | ❌   | ❌   | ✅    | ✅        |
| SSE built-in                | ❌   | ❌   | ✅    | ✅        |
| Context pooling             | ✅   | ✅   | ✅    | ✅        |
| Struct validation           | ❌   | ✅   | ❌    | ✅        |
| Multi-format binding        | ✅   | ✅   | ✅    | ✅        |
| Route groups                | ✅   | ✅   | ✅    | ✅        |
| Middleware chain            | ✅   | ✅   | ✅    | ✅        |
| JWT middleware              | 3rd  | 3rd  | ✅    | ✅        |
| CORS middleware             | 3rd  | ✅   | ✅    | ✅        |
| Rate limiter middleware     | 3rd  | 3rd  | ✅    | ✅        |
| Compression middleware      | 3rd  | ✅   | ✅    | ✅        |
| CSRF middleware             | 3rd  | ✅   | ✅    | ✅        |
| Security headers            | ❌   | ✅   | ✅    | ✅        |
| Graceful shutdown           | ~    | ✅   | ✅    | ✅        |
| net/http compatibility      | ✅   | ✅   | ❌    | ✅        |
| OpenTelemetry integration   | 3rd  | 3rd  | 3rd   | ✅        |
| Prometheus metrics          | 3rd  | 3rd  | 3rd   | ✅        |
| gRPC-Web bridge             | ❌   | ❌   | ❌    | ✅ (v0.4) |
| Named route URL generation  | ✅   | ✅   | ✅    | ✅        |
| File server built-in        | ✅   | ✅   | ✅    | ✅        |
| Multipart upload            | ✅   | ✅   | ✅    | ✅        |
| MessagePack support         | ❌   | ✅   | ❌    | ✅        |
| Sonic JSON (SIMD)           | ❌   | ❌   | ❌    | ✅        |
| CLI scaffolding tool        | ❌   | ❌   | ✅    | ✅ (v0.3) |

---

## 23. Internal Data Flow

### 23.1 Full Request Lifecycle

```
1.  TCP connection accepted by net/http listener
2.  HTTP/1.1 or HTTP/2 frame parsed by net/http
3.  Engine.ServeHTTP(w, r) called
4.  Context acquired from sync.Pool (zero alloc)
5.  Context.Reset(w, r) — O(1) field reset
6.  Global middleware chain composed (closure chain)
7.  Router.Find(method, path, ctx) called
     └─ Static routes: O(1) map lookup
     └─ Dynamic routes: O(log n) radix tree search
     └─ Params stored in ctx.params[16] fixed array
8.  Handler chain executed (middleware → handler)
9.  Handler calls ctx.JSON/ctx.HTML/etc.
     └─ Response headers set
     └─ Status code written
     └─ Body encoded directly to ResponseWriter (zero copy)
10. Deferred Context.Release() clears sensitive refs
11. Context returned to sync.Pool
12. net/http manages keep-alive or connection close
```

---

## 24. Memory Model

### 24.1 Per-Request Memory Layout

```
Stack frame (per goroutine, ~8KB default):
┌────────────────────────────────────────────┐
│  Engine.ServeHTTP stack frame              │
│  └─ ctx pointer (8 bytes)                  │
│  └─ err (16 bytes interface)               │
└────────────────────────────────────────────┘

Heap (sync.Pool, reused across requests):
┌────────────────────────────────────────────┐
│  Context struct (~2KB)                     │
│  ├─ params [16]Param   (512 bytes)         │
│  ├─ body   []byte      (slice header only) │
│  ├─ errors []error     (slice header only) │
│  └─ store  map (nil until needed)          │
└────────────────────────────────────────────┘

One-time heap allocations (reused for lifecycle):
┌────────────────────────────────────────────┐
│  Router radix tree  (built at startup)     │
│  Middleware chain   (built at startup)     │
│  http.Server        (one per Run() call)   │
└────────────────────────────────────────────┘
```

---

## 25. Future Architecture Targets

| Version | Feature                                           |
|---------|---------------------------------------------------|
| v0.2.x  | File server, static asset serving                 |
| v0.3.x  | `rudra` CLI scaffolding tool (new project, route) |
| v0.4.x  | gRPC-Web bridge, Protocol Buffers support         |
| v0.5.x  | Plugin system (hot-reload middleware registry)    |
| v0.6.x  | QUIC / HTTP/3 via `quic-go`                       |
| v0.7.x  | Distributed rate limiter (Redis backend)          |
| v0.8.x  | OpenTelemetry first-class integration             |
| v0.9.x  | Zero-allocation JSON parser (SIMD, custom impl)   |
| v1.0.0  | Stable API, LTS support, full benchmark suite     |

---

## Appendix A — Dependency List

| Package                        | Purpose                              | Why Chosen             |
|--------------------------------|--------------------------------------|------------------------|
| `golang.org/x/net/http2`       | HTTP/2 + h2c support                 | Official Go team       |
| `github.com/bytedance/sonic`   | SIMD-accelerated JSON encoding       | 2-4x faster than stdlib|
| `github.com/golang-jwt/jwt/v5` | JWT parsing/signing                  | Most maintained        |
| `golang.org/x/crypto`          | bcrypt, argon2 for auth helpers      | Official Go team       |
| `github.com/klauspost/compress`| zstd + brotli compression            | Best Go compression lib|
| `go.opentelemetry.io/otel`     | Distributed tracing                  | CNCF standard          |
| `github.com/prometheus/client_golang` | Metrics exposition          | Industry standard      |
| `gopkg.in/yaml.v3`             | YAML config parsing                  | Most complete YAML lib |

All other functionality is implemented from scratch on top of the Go standard library.

---

## Appendix B — Quick Start

```go
package main

import (
    "github.com/AarambhDevHub/rudra"
    "github.com/AarambhDevHub/rudra/context"
    "github.com/AarambhDevHub/rudra/middleware"
)

func main() {
    app := rudra.New(
        rudra.WithHTTP2(),
        rudra.WithReadTimeout(5 * time.Second),
    )

    // Global middleware
    app.Use(
        middleware.Recovery(),
        middleware.Logger(),
        middleware.RequestID(),
        middleware.CORS(middleware.DefaultCORSConfig()),
        middleware.Compress(),
    )

    // Routes
    app.GET("/", func(c *context.Context) error {
        return c.JSON(200, map[string]string{"framework": "Rudra", "status": "fierce"})
    })

    // Route groups
    api := app.Group("/api/v1", middleware.JWT(jwtConfig))
    api.GET("/users/:id", getUser)
    api.POST("/users", createUser)

    // WebSocket
    app.GET("/ws", wsHandler)

    // SSE
    app.GET("/events", broker.ServeHTTP)

    // Graceful shutdown
    go app.ListenForShutdown()

    app.Run(":8080")
}
```

---

*Rudra (रुद्र) — Fierce. Fast. Fearless.*
*© Aarambh Dev Hub — MIT + Apache 2.0*
