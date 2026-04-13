<div align="center">

<img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"/>
<img src="https://img.shields.io/badge/version-0.0.9-brightgreen?style=for-the-badge" alt="Version"/>
<img src="https://img.shields.io/badge/phase_0-complete-brightgreen?style=for-the-badge" alt="Phase 0 Complete"/>
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

> Hardware: ASUS VivoBook 14, Intel Core i3-1115G4 (11th Gen), 8GB RAM, Pop OS, Go 1.22
> Test: 4 threads, 100 connections, 10 seconds (`wrk`) / 100 concurrency, 100k requests (`ab`)

### Router Micro-Benchmarks

| Operation | ns/op | allocs/op |
|-----------|-------|-----------|
| Static route match | **25** | **0** |
| 3-param route match | **69** | **0** |
| Wildcard route match | **34** | **0** |
| Context acquire/release | **14** | **0** |
| Context SetParam | **1.3** | **0** |

### `wrk` Throughput (Phase 0 results)

| Route | Req/sec | Avg Latency | Stdev |
|-------|---------|-------------|-------|
| `GET /` (static) | **206,686** | 688 µs | 831 µs |
| `GET /hello/:name` (param) | **189,005** | 708 µs | 870 µs |

### `ab` Results (Phase 0 results)

| Route | Req/sec | Mean Latency | Max Latency | Failed |
|-------|---------|-------------|-------------|--------|
| `GET /` (static) | **26,763** | 3.7 ms | 14 ms | **0** |
| `GET /hello/:name` (param) | **26,796** | 3.7 ms | 20 ms | **0** |

**Key observations:**
- Parameterized routing is only **~8% slower** than static — radix tree overhead is negligible
- **Zero failed requests** across 200,000 total `ab` requests
- Average latency stays under **0.7ms** under sustained `wrk` load
- These numbers are from Phase 0 — **zero middleware, zero optimizations applied yet**

Cross-framework comparison (`wrk` vs Gin, Echo, Fiber) will be published at `v0.9.x` after the full optimization sprint.

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

    "github.com/AarambhDevHub/rudra/core"
    rudraContext "github.com/AarambhDevHub/rudra/context"
)

func main() {
    app := core.New()

    // Static route
    app.GET("/", func(c *rudraContext.Context) error {
        return c.JSON(http.StatusOK, map[string]string{
            "framework": "Rudra",
            "status":    "fierce",
        })
    })

    // Param route
    app.GET("/users/:id", func(c *rudraContext.Context) error {
        return c.JSON(http.StatusOK, map[string]string{
            "id": c.Param("id"),
        })
    })

    // Wildcard route
    app.GET("/files/*filepath", func(c *rudraContext.Context) error {
        return c.String(http.StatusOK, "File: "+c.Param("filepath"))
    })

    // Route groups with middleware
    api := app.Group("/api/v1")
    api.GET("/users", func(c *rudraContext.Context) error {
        return c.JSON(http.StatusOK, []string{"alice", "bob"})
    })

    // Multiple HTTP methods
    app.POST("/users", func(c *rudraContext.Context) error {
        var body map[string]any
        if err := c.BindJSON(&body); err != nil {
            return c.AbortWithError(http.StatusBadRequest, err)
        }
        return c.JSON(http.StatusCreated, body)
    })

    log.Println("rudra: listening on :8080")
    app.Run(":8080")
}
```

---

## Current Status

**Phase 0 (0.0.1 → 0.0.9) — COMPLETE** ✅

| Version | Feature | Status |
|---------|---------|--------|
| 0.0.1 | Project scaffold, module init | ✅ |
| 0.0.2 | HTTP/1.1 server, functional options | ✅ |
| 0.0.3 | Context system + sync.Pool | ✅ |
| 0.0.4 | Static route router (O(1) map) | ✅ |
| 0.0.5 | Radix tree router (`:param`, `*wildcard`) | ✅ |
| 0.0.6 | Response rendering (JSON, HTML, XML, Stream) | ✅ |
| 0.0.7 | Error handling system | ✅ |
| 0.0.8 | Middleware chain (onion model) | ✅ |
| 0.0.9 | Route groups + named routes | ✅ |

**Next: Phase 1 (0.1.0 → 0.1.9) — HTTP/1.1 Hardening + Core Middleware**

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

### Error Handling
- `RudraError` type with HTTP status codes
- Constructors: 400, 401, 403, 404, 405, 409, 422, 429, 500
- Custom error handlers via `Engine.SetErrorHandler()`
- Panic recovery → 500 JSON (never leaks stack traces)

### Server
- Configurable timeouts: read, write, idle, header
- Graceful shutdown on SIGINT/SIGTERM
- TLS 1.2+ with hardened cipher suites
- Custom `net.Listener` support

---

## Zero-Allocation Design

Rudra is engineered to allocate nothing on the hot path. **All numbers below are confirmed from `go test -bench=. -benchmem`.**

```
Route match (static)          →  25 ns/op   0 allocs   0 bytes
Route match (3 params)        →  69 ns/op   0 allocs   0 bytes
Route match (wildcard)        →  34 ns/op   0 allocs   0 bytes
Context acquire (sync.Pool)   →  14 ns/op   0 allocs   0 bytes
Context SetParam              → 1.3 ns/op   0 allocs   0 bytes
Middleware chain (3 layers)   →   0 allocs   0 bytes
JSON response (≤4KB)          →   1 alloc   (the encoding itself, unavoidable)
```

Key techniques: `sync.Pool` context recycling, fixed-size `[16]Param` array, direct-to-`ResponseWriter` JSON encoding, O(1) static route map bypassing the radix tree entirely.

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
| CORS built-in           | 3rd  | ✅   | ✅    | 🔜 (Phase 1) |
| Rate limiter built-in   | 3rd  | 3rd  | ✅    | 🔜 (Phase 1) |
| Compression built-in   | 3rd  | ✅   | ✅    | 🔜 (Phase 1) |
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
| HTTP/1.1 | 0.1.0–0.1.9 | Full HTTP/1.1 + core middleware | 🔜 Next |
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