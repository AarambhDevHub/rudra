<div align="center">

<img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"/>
<img src="https://img.shields.io/badge/version-0.1.9-brightgreen?style=for-the-badge" alt="Version"/>
<img src="https://img.shields.io/badge/phase_1-complete-brightgreen?style=for-the-badge" alt="Phase 1 Complete"/>
<img src="https://img.shields.io/badge/license-MIT%20%2B%20Apache%202.0-blue?style=for-the-badge" alt="License"/>
<img src="https://img.shields.io/badge/status-active-brightgreen?style=for-the-badge" alt="Status"/>

<br/>

```
██████╗ ██╗   ██╗██████╗ ██████╗  █████╗
██╔══██╗██║   ██║██╔══██╗██╔══██╗██╔══██╗
██████╔╝██║   ██║██║  ██║██████╔╝███████║
██╔══██╗██║   ██║██║  ██║██╔══██╗██╔══██║
██║  ██║╚██████╔╝██████╔╝██║  ██║██║  ██║
╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝
```

# Rudra (रुद्र)

### **Fierce. Fast. Fearless.**

A zero-allocation, batteries-included Go web framework built on `net/http`.
HTTP/1.1 · HTTP/2 · WebSocket · SSE · Radix Tree Router · Zero Alloc Hot Path

**Sister project of [Ajaya](https://github.com/AarambhDevHub/ajaya) (Rust) — by [Aarambh Dev Hub](https://github.com/AarambhDevHub)**

</div>

---

## Why Rudra?

Go already has great frameworks. So why build another one?

Because none of them do everything well in a single package. **Gin** is fast but has no HTTP/2 push, no WebSocket, no SSE, no validation. **Echo** is good but Fiber breaks `net/http` compatibility. **Fiber** has the features but you lose the entire stdlib ecosystem.

Rudra is built with one goal: **be the framework you never have to leave**. Everything you need — HTTP/1.1 hardened, HTTP/2 native, WebSocket with rooms, SSE with backpressure, zero-allocation routing, built-in JWT/CORS/RateLimit/Compression — ships in one `go get`.

And it stays compatible with `net/http`. Your existing middleware works. Your existing tools work.

---

## Benchmarks

> Hardware: Intel Core i5-1135G7 (11th Gen), 8GB RAM, Pop OS, Go 1.22
> Test: 4 threads, 100 connections, 10 seconds (`wrk`)

### Router Micro-Benchmarks

| Operation | ns/op | allocs/op |
|-----------|-------|-----------|
| Static route match | **26** | **0** |
| 3-param route match | **68** | **0** |
| Wildcard route match | **33** | **0** |
| Context acquire/release | **14** | **0** |
| Context SetParam | **1.3** | **0** |

### `wrk` Throughput (Phase 1 results)

| Route | Req/sec | Avg Latency | Stdev |
|-------|---------|-------------|-------|
| `GET /` (static) | **213,817** | 671 µs | 809 µs |
| `GET /hello/:name` (param) | **178,407** | 783 µs | 920 µs |

### `ab` Results (Phase 1 results)

| Route | Req/sec | Mean Latency | Max Latency | Failed |
|-------|---------|-------------|-------------|--------|
| `GET /` (static) | **27,158** | 3.68 ms | 8 ms | **0** |
| `GET /hello/:name` (param) | **27,833** | 3.59 ms | 9 ms | **0** |

**Key observations:**
- Static route throughput **+3.5%** over Phase 0 — TCP tuning (`TCP_NODELAY` + `SO_REUSEPORT`) delivers measurable improvement
- Parameterized routing is only **~16% slower** than static — radix tree overhead remains negligible
- **Zero failed requests** across all benchmark runs
- Average latency stays under **0.8ms** under sustained `wrk` load
- All micro-benchmarks: **0 allocs/op** — zero-allocation hot path preserved

---

## Quick Start

```bash
go get github.com/AarambhDevHub/rudra
```

```go
package main

import (
    "log"
    "net/http"
    "time"

    "github.com/AarambhDevHub/rudra/core"
    rudraContext "github.com/AarambhDevHub/rudra/context"
    "github.com/AarambhDevHub/rudra/middleware"
)

func main() {
    app := core.New()

    // Production middleware stack
    app.Use(middleware.Recovery())                    // panic recovery → 500 JSON
    app.Use(middleware.RequestID())                   // UUID v4 per request
    app.Use(middleware.Logger())                      // structured JSON access logs
    app.Use(middleware.Timeout(middleware.TimeoutConfig{
        Timeout: 10 * time.Second,                   // per-request deadline
    }))
    app.Use(middleware.CORS())                        // CORS with permissive defaults

    // Static route
    app.GET("/", func(c *rudraContext.Context) error {
        return c.JSON(http.StatusOK, map[string]string{
            "framework":  "Rudra",
            "request_id": c.RequestID(),
        })
    })

    // Param route
    app.GET("/users/:id", func(c *rudraContext.Context) error {
        return c.JSON(http.StatusOK, map[string]string{
            "id": c.Param("id"),
        })
    })

    // Route groups
    api := app.Group("/api/v1")
    api.GET("/users", func(c *rudraContext.Context) error {
        return c.JSON(http.StatusOK, []string{"alice", "bob"})
    })

    go func() {
        log.Println("rudra: listening on :8080")
        if err := app.Run(":8080"); err != nil && err != http.ErrServerClosed {
            log.Fatalf("rudra: server error: %v", err)
        }
    }()

    app.ListenForShutdown()
}
```

---

## Current Status

**Phase 0 (0.0.1 → 0.0.9) — COMPLETE** ✅ — Engine, router, context, rendering, error handling, middleware chain, route groups

**Phase 1 (0.1.0 → 0.1.9) — COMPLETE** ✅

| Version | Feature | Status |
|---------|---------|--------|
| 0.1.0 | TCP tuning, graceful shutdown, TLS hardening | ✅ |
| 0.1.1 | Logger middleware (slog JSON/text/Apache) | ✅ |
| 0.1.2 | Recovery middleware (panic → 500 JSON) | ✅ |
| 0.1.3 | RequestID middleware (UUID v4, forwarding) | ✅ |
| 0.1.4 | Timeout middleware (per-request deadline) | ✅ |
| 0.1.5 | CORS middleware (preflight, credentials, dynamic origins) | ✅ |
| 0.1.6 | BodyLimit + Secure headers (HSTS, CSP, XSS, X-Frame) | ✅ |
| 0.1.7 | Rate limiter (token bucket, per-IP, X-RateLimit-*) | ✅ |
| 0.1.8 | Compression (gzip, sync.Pool'd writers) | ✅ |
| 0.1.9 | CSRF + ETag + Static file server | ✅ |

**Phase 2 (0.2.0 → 0.2.9) — UP NEXT** 🔄

| Version | Feature | Status |
|---------|---------|--------|
| 0.2.0 | JSON binding (`BindJSON`) | 🔜 |
| 0.2.1 | Form + Multipart binding | 🔜 |
| 0.2.2 | Query + Path + Header binding | 🔜 |
| 0.2.3 | XML binding + rendering | 🔜 |
| 0.2.4 | MessagePack binding + rendering | 🔜 |
| 0.2.5 | Validator core (required, min, max, email, url) | 🔜 |
| 0.2.6 | Validator extended rules (uuid, oneof, dive, cross-field) | 🔜 |
| 0.2.7 | Custom validator rules | 🔜 |
| 0.2.8 | ShouldBind + MustBind auto-detection | 🔜 |
| 0.2.9 | Binding benchmarks + optimization | 🔜 |

---

## Features

### Router
- Radix tree router — O(log n) worst case
- O(1) fast path for static routes (map lookup)
- `:param` and `*wildcard` with zero heap allocation
- Route groups with prefix + shared middleware
- Named routes + URL generation: `router.URL("user.profile", "42")`
- 404 / 405 error handling
- Route conflict detection (panic on duplicate)

### Context
- `sync.Pool` — zero allocation context acquisition
- `[16]Param` fixed array — no heap alloc for path params
- Per-request key-value store (lazy initialized)
- `c.JSON()`, `c.HTML()`, `c.String()`, `c.Blob()`, `c.Stream()`, `c.XML()`, `c.JSONP()`
- `c.BindJSON()`, `c.BindXML()`
- `c.Redirect()`, `c.NoContent()`

### Middleware
- Onion model middleware chain
- `Engine.Use()` for global middleware
- Route-level middleware: `app.GET("/path", handler, mw1, mw2)`
- `c.Next()` / `c.Abort()` chain control

### Built-in Middleware (v0.1.9)
- **Logger** — structured JSON/text/Apache access logs via `log/slog`, latency + bytes tracking, skip paths
- **Recovery** — `defer/recover`, stack trace in logs (never in response), `OnPanic` hook
- **RequestID** — UUID v4 via `crypto/rand`, forwards upstream `X-Request-ID`, custom generators
- **Timeout** — `context.WithTimeout` per-request, custom timeout handler, race-free goroutine design
- **CORS** — preflight + simple requests, `AllowOriginFunc`, credentials, `MaxAge`, pre-computed headers
- **BodyLimit** — `http.MaxBytesReader` wrapping, 413 on exceed, configurable limit (default 32MB)
- **Secure** — HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, Permissions-Policy
- **RateLimit** — token bucket per IP, `X-RateLimit-*` headers, `Retry-After`, background cleanup
- **Compress** — gzip with `sync.Pool`’d writers, min-length skip, content-type filter
- **CSRF** — double-submit cookie, `crypto/rand` tokens, constant-time comparison
- **ETag** — SHA-256 body hash, 304 Not Modified, weak ETag support

### Static Files
- `Engine.Static(prefix, root)` — directory serving
- `Engine.StaticFile(path, file)` — single file serving
- `Engine.StaticFS(prefix, fs)` — `http.FileSystem` / `embed.FS` support
- Directory listing disabled, traversal protection

### Error Handling
- `RudraError` type with HTTP status codes
- Constructors: 400, 401, 403, 404, 405, 409, 422, 429, 500
- Custom error handlers via `Engine.SetErrorHandler()`
- Panic recovery → 500 JSON (never leaks stack traces)

### Server
- Configurable timeouts: read, write, idle, header
- TCP socket tuning: `TCP_NODELAY` (default), `SO_REUSEPORT`, `TCP_FASTOPEN` (Linux)
- Graceful shutdown on SIGINT/SIGTERM (idempotent, race-free)
- TLS 1.2+ with hardened AEAD cipher suites + session resumption
- Custom `net.Listener` support

---

## Zero-Allocation Design

Rudra is engineered to allocate nothing on the hot path. **All numbers below are confirmed from `go test -bench=. -benchmem`.**

```
Route match (static)          →  26 ns/op   0 allocs   0 bytes
Route match (3 params)        →  68 ns/op   0 allocs   0 bytes
Route match (wildcard)        →  33 ns/op   0 allocs   0 bytes
Context acquire (sync.Pool)   →  14 ns/op   0 allocs   0 bytes
Context SetParam              → 1.3 ns/op   0 allocs   0 bytes
Middleware chain (3 layers)   →   0 allocs   0 bytes
JSON response (≤4KB)          →   1 alloc   (the encoding itself, unavoidable)
```

Key techniques: `sync.Pool` context recycling, fixed-size `[16]Param` array, direct-to-`ResponseWriter` JSON encoding, O(1) static route map bypassing the radix tree entirely, `TCP_NODELAY` for reduced latency on small responses.

---

## Project Structure

```
rudra/
├── core/         Engine, server, options, graceful shutdown
├── router/       Radix tree + route groups + named routes
├── context/      Request/response context + sync.Pool
├── middleware/   Built-in middleware (Phase 1)
├── binding/      Request data binding (Phase 2)
├── render/       Response renderers (JSON, HTML, Stream…)
├── ws/           WebSocket (Phase 4)
├── sse/          Server-Sent Events (Phase 5)
├── validator/    Struct validation (Phase 2)
├── config/       YAML + environment config (Phase 7)
├── errors/       HTTP error types + global handler
├── testutil/     Testing utilities + assertion helpers
├── benchmarks/   Router and framework benchmarks
└── examples/      Working examples
```

---

## API Reference

### Engine

```go
app := core.New(
    core.WithReadTimeout(5 * time.Second),
    core.WithWriteTimeout(10 * time.Second),
    core.WithTCPNoDelay(true),
    core.WithSOReusePort(true),
    core.WithTCPFastOpen(true),
    core.WithHTTP2(),
)

app.GET("/", handler)
app.POST("/", handler)
app.PUT("/:id", handler)
app.PATCH("/:id", handler)
app.DELETE("/:id", handler)
app.OPTIONS("/", handler)
app.HEAD("/", handler)
app.Any("/", handler)

app.Use(middleware1, middleware2)

api := app.Group("/api/v1")
api.GET("/users", listUsers)
```

### Middleware

```go
import "github.com/AarambhDevHub/rudra/middleware"

// Recommended production stack (order matters — outermost first)
app.Use(middleware.Recovery())                                 // catches panics
app.Use(middleware.RequestID())                                // generates X-Request-ID
app.Use(middleware.Logger(middleware.LoggerConfig{              // structured access logs
    Format:    "json",
    SkipPaths: []string{"/health"},
}))
app.Use(middleware.Timeout(middleware.TimeoutConfig{            // per-request deadline
    Timeout: 10 * time.Second,
}))
app.Use(middleware.CORS(middleware.CORSConfig{                  // CORS headers
    AllowOrigins:     []string{"https://app.example.com"},
    AllowCredentials: true,
    MaxAge:           3600,
}))
```

### Context

```go
func handler(c *rudraContext.Context) error {
    // Path params
    id := c.Param("id")

    // Query params
    page := c.QueryDefault("page", "1")

    // Headers
    auth := c.Header("Authorization")
    c.SetHeader("X-Custom", "value")

    // Request body
    var body MyStruct
    if err := c.BindJSON(&body); err != nil {
        return c.AbortWithError(400, err)
    }

    // Store values for middleware
    c.Set("user", user)

    // Responses
    return c.JSON(200, data)
}
```

### Error Handling

```go
// Built-in error constructors
err := errors.NotFound("user not found")
err := errors.BadRequest("invalid input")
err := errors.InternalServerError("something broke")

// Custom error handler
app.SetErrorHandler(func(c *rudraContext.Context, err error) {
    var re *errors.RudraError
    if errors.As(err, &re) {
        c.JSON(re.Code, re)
        return
    }
    c.JSON(500, map[string]string{"error": "internal server error"})
})
```

---

## Comparison

| Feature                 | Gin  | Echo | Fiber | **Rudra** |
|-------------------------|------|------|-------|-----------|
| Zero-alloc routing      | ~    | ~    | ✅    | ✅        |
| HTTP/2 native           | ~    | ~    | ❌    | 🔜 (Phase 3) |
| h2c plaintext HTTP/2    | ❌   | ❌   | ❌    | 🔜 (Phase 3) |
| HTTP/2 server push      | ~    | ~    | ❌    | 🔜 (Phase 3) |
| WebSocket built-in      | ❌   | ❌   | ✅    | 🔜 (Phase 4) |
| SSE built-in            | ❌   | ❌   | ✅    | 🔜 (Phase 5) |
| net/http compatibility  | ✅   | ✅   | ❌    | ✅        |
| Struct validation       | ❌   | ✅   | ❌    | 🔜 (Phase 2) |
| JWT built-in            | 3rd  | 3rd  | ✅    | 🔜 (Phase 6) |
| CORS built-in           | 3rd  | ✅   | ✅    | ✅            |
| Rate limiter built-in   | 3rd  | 3rd  | ✅    | ✅            |
| Compression built-in   | 3rd  | ✅   | ✅    | ✅            |
| CSRF built-in           | 3rd  | ✅   | ✅    | ✅            |
| Static file server      | 3rd  | ✅   | ✅    | ✅            |
| OpenTelemetry           | 3rd  | 3rd  | 3rd   | 🔜 (Phase 7) |
| Prometheus metrics      | 3rd  | 3rd  | 3rd   | 🔜 (Phase 7) |
| Sonic JSON (SIMD)      | ❌   | ❌   | ❌    | 🔜 (Phase 2) |
| gRPC-Web bridge         | ❌   | ❌   | ❌    | 🔜 (Phase 9) |

✅ = Complete · 🔜 = Planned · ~ = Partial · 3rd = Third-party required · ❌ = Not available

---

## Roadmap

See [`ROADMAP.md`](./ROADMAP.md) for the full version-by-version plan from `0.0.1` to `0.9.9`.

| Phase | Versions | Theme | Status |
|-------|----------|-------|--------|
| Foundation | 0.0.1–0.0.9 | Engine, router, context, render | ✅ Complete |
| HTTP/1.1 | 0.1.0–0.1.9 | Full HTTP/1.1 + core middleware | ✅ Complete |
| Binding | 0.2.0–0.2.9 | Request binding + validation | 📋 Planned |
| HTTP/2 | 0.3.0–0.3.5 | TLS, h2c, server push | 📋 Planned |
| WebSocket | 0.4.0–0.4.6 | WS, hub, rooms, compression | 📋 Planned |
| SSE | 0.5.0–0.5.5 | SSE, heartbeat, reconnect | 📋 Planned |
| Auth | 0.6.0–0.6.4 | JWT, Basic, API Key | 📋 Planned |
| Observability | 0.7.0–0.7.5 | Config, logging, tracing, metrics | 📋 Planned |
| DX + CLI | 0.8.0–0.8.9 | Testing, CLI, templates, docs | 📋 Planned |
| Launch Prep | 0.9.0–0.9.9 | Hardening, perf, RC1, RC2 | 📋 Planned |

---

## Contributing

All contributions are welcome. Please read [`CONTRIBUTING.md`](./CONTRIBUTING.md) before opening a PR.

```bash
git clone https://github.com/AarambhDevHub/rudra
cd rudra
go mod tidy
go test ./...
go test -race ./...
go test -bench=. -benchmem ./...
```

---

## Community

- **YouTube**: [Aarambh Dev Hub](https://youtube.com/@AarambhDevHub) — tutorials, deep dives, build-in-public
- **Discord**: [Join the community](https://discord.com/invite/HDth6PfCnp)
- **GitHub Discussions**: [AarambhDevHub/rudra/discussions](https://github.com/AarambhDevHub/rudra/discussions)
- **Issues**: [AarambhDevHub/rudra/issues](https://github.com/AarambhDevHub/rudra/issues)

---

## Support the Project

If Rudra saves you time, consider supporting Aarambh Dev Hub:

- ⭐ **Star this repository** — it genuinely helps
- ☕ **Buy Me a Coffee**: [buymeacoffee.com/aarambhdevhub](https://buymeacoffee.com/aarambhdevhub)
- 💖 **GitHub Sponsors**: [github.com/sponsors/aarambh-darshan](https://github.com/sponsors/aarambh-darshan)
- 💼 **Hire for Rust/Go work**: [Fiverr](https://fiverr.com/s/XL1ab4G)

---

## License

Rudra is dual-licensed under **MIT** and **Apache 2.0**. You may choose either license.

- [MIT License](./LICENSE-MIT)
- [Apache License 2.0](./LICENSE-APACHE)

---

<div align="center">

Built with ❤️ by [Aarambh Dev Hub](https://github.com/AarambhDevHub)

*Sister project: [Ajaya (अजय)](https://github.com/AarambhDevHub/ajaya) — The Unconquerable Rust Web Framework*

</div>