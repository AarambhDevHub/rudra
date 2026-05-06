<div align="center">

<img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"/>
<img src="https://img.shields.io/badge/version-0.2.9-brightgreen?style=for-the-badge" alt="Version"/>
<img src="https://img.shields.io/badge/phase_2-complete-brightgreen?style=for-the-badge" alt="Phase 2 Complete"/>
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
HTTP/1.1 · Radix Tree Router · Zero Alloc Hot Path · Full Binding & Validation

**Sister project of [Ajaya](https://github.com/AarambhDevHub/ajaya) (Rust) — by [Aarambh Dev Hub](https://github.com/AarambhDevHub)**

</div>

---

## Why Rudra?

Go already has great frameworks. So why build another one?

Because none of them do everything well in a single package. **Gin** has no built-in validation. **Echo** lacks zero-alloc routing. **Fiber** breaks `net/http` compatibility. Rudra is built to be **the framework you never have to leave** — fast routing, full binding, struct validation, and production middleware all in one `go get`.

---

## Benchmarks

> Hardware: Intel Core i5-1135G7, 8GB RAM, Pop OS, Go 1.22
> Test: 4 threads, 100 connections, 10 seconds (`wrk`)

### Router Micro-Benchmarks

| Operation | ns/op | allocs/op |
|-----------|-------|-----------|
| Static route match | **26** | **0** |
| 3-param route match | **68** | **0** |
| Wildcard route match | **33** | **0** |
| Context acquire/release | **14** | **0** |

### `wrk` Throughput

| Route | Req/sec | Avg Latency |
|-------|---------|-------------|
| `GET /` (static) | **213,817** | 671 µs |
| `GET /hello/:name` (param) | **178,407** | 783 µs |

### JSON Binding (Phase 2)

| Engine | ns/op | allocs |
|--------|-------|--------|
| stdlib (`encoding/json`) | ~620 | 3 |
| Sonic (`-tags sonic`) | ~185 | 1 |

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
    "github.com/AarambhDevHub/rudra/middleware"
    "github.com/AarambhDevHub/rudra/validator"
)

type CreateUserRequest struct {
    Name  string `json:"name"  rudra:"required,min=2,max=64"`
    Email string `json:"email" rudra:"required,email"`
    Age   int    `json:"age"   rudra:"required,min=18,max=120"`
    Role  string `json:"role"  rudra:"required,oneof=admin user guest"`
}

func main() {
    app := core.New()

    app.Use(middleware.Recovery())
    app.Use(middleware.RequestID())
    app.Use(middleware.Logger())
    app.Use(middleware.CORS(middleware.DefaultCORSConfig()))

    app.POST("/users", func(c *rudraContext.Context) error {
        var req CreateUserRequest
        if err := c.MustBind(&req); err != nil {
            return err // already aborted with 400/422
        }
        return c.JSON(http.StatusCreated, map[string]any{
            "id":   42,
            "name": req.Name,
        })
    })

    app.GET("/users/:id", func(c *rudraContext.Context) error {
        type PathParams struct {
            ID int64 `path:"id"`
        }
        var p PathParams
        if err := c.BindPath(&p); err != nil {
            return err
        }
        return c.JSON(http.StatusOK, map[string]any{"id": p.ID})
    })

    go func() {
        log.Println("rudra: listening on :8080")
        if err := app.Run(":8080"); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    app.ListenForShutdown()
}
```

---

## Current Status

**Phase 0 (0.0.1 → 0.0.9)** ✅ Engine, router, context, rendering, middleware chain

**Phase 1 (0.1.0 → 0.1.9)** ✅ Logger, Recovery, RequestID, Timeout, CORS, BodyLimit, Secure, RateLimit, Compress, CSRF, ETag, Static

**Phase 2 (0.2.0 → 0.2.9)** ✅ **Binding & Validation — complete**

| Version | Feature | Status |
|---------|---------|--------|
| 0.2.0 | JSON binding + decoder foundation + struct-field cache | ✅ |
| 0.2.1 | Form + multipart binding | ✅ |
| 0.2.2 | Query + path + header binding; full type coercion | ✅ |
| 0.2.3 | XML binding + rendering | ✅ |
| 0.2.4 | MessagePack binding + rendering (build tag `msgpack`) | ✅ |
| 0.2.5 | Validator core: required, min, max, email, url | ✅ |
| 0.2.6 | Extended rules: uuid, len, oneof, alphanum, regexp, eqfield, … | ✅ |
| 0.2.7 | Custom rule + message registration | ✅ |
| 0.2.8 | `ShouldBind`, `MustBind`, `ShouldBindWith`, body cache | ✅ |
| 0.2.9 | Sonic JSON integration (build tag `sonic`) | ✅ |

**Phase 3 (0.3.0 → 0.3.5)** 🔜 HTTP/2 — up next

---

## Binding

Rudra auto-selects the right binder from the request's `Content-Type`:

```go
// Auto-detection (recommended)
var req CreateUserRequest
if err := c.ShouldBind(&req); err != nil {
    return err
}

// Bind + Validate in one call (aborts 400/422 on failure)
if err := c.MustBind(&req); err != nil {
    return err
}

// Explicit binder
import "github.com/AarambhDevHub/rudra/binding"
if err := c.ShouldBindWith(&req, binding.JSON); err != nil { ... }

// Query parameters: GET /users?page=2&limit=20
type Pagination struct {
    Page  int `query:"page"`
    Limit int `query:"limit"`
}
var p Pagination
_ = c.BindQuery(&p)

// Path parameters: /orgs/:org/repos/:repo
type RepoParams struct {
    Org  string `path:"org"`
    Repo string `path:"repo"`
}
var rp RepoParams
_ = c.BindPath(&rp)

// Request headers
type AuthHeaders struct {
    Authorization string `header:"Authorization"`
    RequestID     string `header:"X-Request-Id"`
}
var h AuthHeaders
_ = c.BindHeader(&h)
```

### Struct Tag Reference

```go
type CreateUserRequest struct {
    // Basic rules
    Name     string   `json:"name"     rudra:"required,min=2,max=64"`
    Email    string   `json:"email"    rudra:"required,email"`
    Age      int      `json:"age"      rudra:"required,min=18,max=120"`

    // Enum validation
    Role     string   `json:"role"     rudra:"required,oneof=admin user guest"`

    // String character rules
    Code     string   `json:"code"     rudra:"alphanum"`
    Phone    string   `json:"phone"    rudra:"regexp=^[6-9]\\d{9}$"`

    // Cross-field rules
    Password string   `json:"password" rudra:"required,min=8"`
    Confirm  string   `json:"confirm"  rudra:"required,eqfield=Password"`

    // Slice with dive (validate each element)
    Tags     []string `json:"tags"     rudra:"min=1,dive,required"`

    // Nested struct (validated recursively)
    Address  Address  `json:"address"`
}
```

### Supported Built-in Rules

| Rule | Description |
|------|-------------|
| `required` | Field must not be zero value (empty string, 0, nil, false, empty slice) |
| `min=N` | Min length (string/slice) or min value (number) |
| `max=N` | Max length (string/slice) or max value (number) |
| `len=N` | Exact length |
| `email` | Valid email address format |
| `url` | Valid http/https URL |
| `uuid` | Valid UUID (any version) |
| `oneof=a b c` | Value must be one of the space-separated list |
| `alphanum` | Only `[a-zA-Z0-9]` characters |
| `alpha` | Only alphabetic characters |
| `numeric` | Only digit characters |
| `regexp=pattern` | Must match the compiled pattern (cached at first use) |
| `eqfield=Field` | Must equal another field (cross-field) |
| `nefield=Field` | Must not equal another field (cross-field) |
| `dive` | Validate each element of a slice or map |

### Custom Rules

```go
import (
    "reflect"
    "regexp"
    "github.com/AarambhDevHub/rudra/validator"
)

var phoneRe = regexp.MustCompile(`^[6-9]\d{9}$`)

func init() {
    validator.Register("indianphone", func(field string, v reflect.Value, _ string, _ reflect.Value) string {
        if v.Kind() != reflect.String || !phoneRe.MatchString(v.String()) {
            return "indianphone"
        }
        return ""
    })
    validator.RegisterMessage("indianphone", "'{field}' must be a valid 10-digit Indian mobile number")
}
```

### Validation Errors

```go
if err := c.Validate(&req); err != nil {
    ve := err.(validator.ValidationErrors)

    // Full error string
    log.Println(ve.Error()) // "'name' is required; 'email' must be a valid email"

    // First error only
    log.Println(ve.First().Field, ve.First().Rule)

    // Errors for a specific field
    for _, fe := range ve.ForField("email") {
        log.Println(fe.Rule, fe.Message)
    }

    // Map for JSON response
    return c.JSON(422, map[string]any{
        "message": "validation failed",
        "errors":  ve.Map(), // {"name": "...", "email": "..."}
    })
}
```

---

## Build Tags

| Feature | Tag | Dependency |
|---------|-----|------------|
| Sonic JSON (2–4× faster) | `sonic` | `github.com/bytedance/sonic` |
| MessagePack binding + rendering | `msgpack` | `github.com/shamaton/msgpack/v2` |

```bash
# Enable Sonic JSON acceleration
go get github.com/bytedance/sonic
go build -tags sonic ./...

# Enable MessagePack
go get github.com/shamaton/msgpack/v2
go build -tags msgpack ./...

# Both
go build -tags "sonic msgpack" ./...
```

---

## Features

### Binding (Phase 2 — complete)
- **Auto-detection** — `ShouldBind` selects binder from Content-Type
- **Bind + Validate** — `MustBind` aborts with correct HTTP status on failure
- **JSON** — stdlib (`encoding/json`) or Sonic SIMD (`-tags sonic`)
- **XML** — `encoding/xml` streaming decoder
- **Form** — `application/x-www-form-urlencoded` + `multipart/form-data`
- **Query** — URL query parameters with slice/comma-split support
- **Path** — URL path parameters typed to struct fields
- **Header** — Request headers with case-insensitive matching
- **MessagePack** — optional via `-tags msgpack`
- **Type coercion** — `string → int, uint, float, bool, []T, *T, time.Time`
- **Struct-field metadata cache** — one reflection walk per type per binder

### Validation (Phase 2 — complete)
- **14 built-in rules** — required, min, max, len, email, url, uuid, oneof, alphanum, alpha, numeric, regexp, eqfield, nefield
- **Cross-field rules** — eqfield, nefield using root struct value
- **Dive** — validate each element of slices/maps
- **Nested struct** — recursive validation
- **Embedded struct** — transparent anonymous field traversal
- **Custom rules** — `validator.Register(name, fn)` globally and thread-safely
- **Custom messages** — `validator.RegisterMessage(rule, template)` with `{field}`, `{param}` placeholders
- **Per-type cache** — `sync.Map` keyed by `reflect.Type`; near-zero cost after warm-up
- **Pre-compiled regexp** — patterns cached per string in `sync.Map`
- **Rich error type** — `ValidationErrors` with `.First()`, `.ForField()`, `.Has()`, `.Map()`

### Router
- Radix tree — O(log n) worst case, O(1) static fast-path
- `:param` and `*wildcard` — zero heap allocation (fixed `[16]Param` array)
- Route groups, named routes, URL generation
- 404 / 405 handling

### Middleware (Phase 1 — complete)
- Logger (slog JSON/text/Apache), Recovery, RequestID, Timeout, CORS, BodyLimit, Secure, RateLimit, Compress, CSRF, ETag

---

## Project Structure

```
rudra/
├── core/         Engine, server, options, graceful shutdown
├── router/       Radix tree + route groups + named routes
├── context/      Request/response context + sync.Pool
├── binding/      Request data binding (JSON, XML, Form, Query, Path, Header, Msgpack)
├── validator/    Struct validation with tag-based rules
├── middleware/   Built-in middleware (Phase 1)
├── render/       Response renderers (JSON, HTML, Stream, Msgpack…)
├── ws/           WebSocket (Phase 4)
├── sse/          Server-Sent Events (Phase 5)
├── config/       YAML + environment config (Phase 7)
├── errors/       HTTP error types + global handler
├── testutil/     Testing utilities + assertion helpers
├── benchmarks/   Router and framework benchmarks
└── examples/     Working examples
```

---

## Comparison

| Feature | Gin | Echo | Fiber | **Rudra** |
|---------|-----|------|-------|-----------|
| Zero-alloc routing | ~ | ~ | ✅ | ✅ |
| net/http compatibility | ✅ | ✅ | ❌ | ✅ |
| Struct validation built-in | ❌ | ✅ | ❌ | ✅ |
| Custom validation rules | ❌ | ✅ | ❌ | ✅ |
| Multi-format binding | ✅ | ✅ | ✅ | ✅ |
| Query/Path/Header binding | ✅ | ✅ | ✅ | ✅ |
| MessagePack | ❌ | ✅ | ❌ | ✅ |
| Sonic JSON | ❌ | ❌ | ❌ | ✅ |
| CORS built-in | 3rd | ✅ | ✅ | ✅ |
| Rate limiter built-in | 3rd | 3rd | ✅ | ✅ |
| CSRF built-in | 3rd | ✅ | ✅ | ✅ |
| HTTP/2 native | ~ | ~ | ❌ | 🔜 (Phase 3) |
| WebSocket built-in | ❌ | ❌ | ✅ | 🔜 (Phase 4) |
| SSE built-in | ❌ | ❌ | ✅ | 🔜 (Phase 5) |
| OpenTelemetry | 3rd | 3rd | 3rd | 🔜 (Phase 7) |

✅ = Complete · 🔜 = Planned · ~ = Partial · 3rd = Third-party required · ❌ = Not available

---

## Contributing

All contributions are welcome. Please read [`CONTRIBUTING.md`](./CONTRIBUTING.md).

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

- **YouTube**: [Aarambh Dev Hub](https://youtube.com/@AarambhDevHub)
- **Discord**: [Join the community](https://discord.com/invite/HDth6PfCnp)
- **GitHub Discussions**: [AarambhDevHub/rudra/discussions](https://github.com/AarambhDevHub/rudra/discussions)

---

## Support the Project

- ⭐ **Star this repository**
- ☕ **Buy Me a Coffee**: [buymeacoffee.com/aarambhdevhub](https://buymeacoffee.com/aarambhdevhub)
- 💖 **GitHub Sponsors**: [github.com/sponsors/aarambh-darshan](https://github.com/sponsors/aarambh-darshan)
- 💼 **Hire for Rust/Go work**: [Fiverr](https://fiverr.com/s/XL1ab4G)

---

## License

Rudra is dual-licensed under **MIT** and **Apache 2.0**. You may choose either.

- [MIT License](./LICENSE-MIT)
- [Apache License 2.0](./LICENSE-APACHE)

---

<div align="center">

Built with ❤️ by [Aarambh Dev Hub](https://github.com/AarambhDevHub)

*Sister project: [Ajaya (अजय)](https://github.com/AarambhDevHub/ajaya) — The Unconquerable Rust Web Framework*

</div>