# Changelog

All notable changes to Rudra (रुद्र) are documented here.

This project adheres to [Semantic Versioning](https://semver.org) and [Conventional Commits](https://www.conventionalcommits.org).

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

---

## [Unreleased]

> Phase 1 — HTTP/1.1 Hardening + Core Middleware (`0.1.0` → `0.1.9`)

### Planned

- `0.1.1` — `middleware.Logger` — structured access log (slog JSON)
- `0.1.2` — `middleware.Recovery` — panic recovery, stack in logs never in response
- `0.1.3` — `middleware.RequestID` — UUID v4 per request, `X-Request-ID` header
- `0.1.4` — `middleware.Timeout` — per-request context deadline
- `0.1.5` — `middleware.CORS` — full CORS with preflight cache
- `0.1.6` — `middleware.BodyLimit` + `middleware.Secure` — security headers
- `0.1.7` — `middleware.RateLimit` — token bucket, per-IP, `X-RateLimit-*` headers
- `0.1.8` — `middleware.Compress` — gzip + brotli + zstd, Accept-Encoding negotiated
- `0.1.9` — `middleware.CSRF` + `middleware.ETag` + `Engine.Static`

---

## [0.1.0] — 2026-04-14

### Added

- Custom TCP listener with `SO_REUSEPORT` + `TCP_NODELAY` + `TCP_FASTOPEN` (Linux)
- Platform-specific build tags: `server_linux.go` / `server_other.go`
- `core.WithTCPNoDelay(bool)` — functional option for `TCP_NODELAY` (default: true)
- `core.WithSOReusePort(bool)` — functional option for `SO_REUSEPORT` (default: false, Linux only)
- `core.WithTCPFastOpen(bool)` — functional option for `TCP_FASTOPEN` (default: false, Linux 3.7+)
- `core.WithShutdownTimeout(d)` — functional option for graceful shutdown timeout
- `Options.TCPNoDelay`, `Options.SOReusePort`, `Options.TCPFastOpen` config fields
- `Engine.Run(addr)` refactored to use custom `newTCPListener()` instead of `ListenAndServe`
- `Engine.RunTLS()` hardened with explicit AEAD cipher suites (AES-GCM + ChaCha20-Poly1305)
- TLS session resumption enabled via `SessionTicketsDisabled: false`
- Removed deprecated `PreferServerCipherSuites` (Go 1.22+)
- `Engine.Shutdown()` made idempotent via `sync.Once` — safe to call multiple times
- `Engine.server` field protected by `sync.RWMutex` — race-free concurrent `Run`/`Shutdown`
- `examples/hello/main.go` updated with graceful shutdown pattern (`go app.Run()` + `app.ListenForShutdown()`)
- `core/server_test.go` — 11 new tests: custom listener, socket options, graceful shutdown, idempotent shutdown

### Changed

- `Engine.Run()` now creates listener via `newTCPListener()` before `http.Server`, enabling socket tuning
- `Engine.RunTLS()` now accepts custom listener + explicit cipher suite list instead of Go defaults
- `Engine.RunListener()` sets `e.server` under mutex protection
- `Engine.Shutdown()` uses `sync.Once` to prevent double-close of `shutdownCh`

### Performance

- `wrk` throughput (4 threads, 100 conn, 10s, i5-1135G7):
  - Static route (`GET /`): **213,817 req/sec** @ 671µs avg latency (+3.5% over v0.0.9)
  - Param route (`GET /hello/:name`): **178,407 req/sec** @ 783µs avg latency
- Router micro-benchmarks (zero regression):
  - Static route match: **26 ns/op, 0 allocs/op**
  - 3-param route match: **68 ns/op, 0 allocs/op**
  - Wildcard route match: **33 ns/op, 0 allocs/op**
  - Context acquire/release: **14 ns/op, 0 allocs/op**

---

## [0.0.9] — 2026-04-13

### Added

- `Engine.Group(prefix, middleware...)` — route groups with prefix and shared middleware
- `Group.Group(prefix, middleware...)` — nested route groups
- `Group.Use(middleware...)` — add middleware to existing group
- `Group.GET/POST/PUT/PATCH/DELETE/OPTIONS/HEAD` — method registration on groups
- `router.Name(method, path, name)` — named route registration
- `router.URL(name, params...)` — URL generation from named routes with param substitution
- `examples/rest-api/main.go` — full REST API example with CRUD operations

---

## [0.0.8] — 2026-04-13

### Added

- `HandlerFunc` type: `func(*context.Context) error`
- `Engine.Use(middleware...)` — global middleware registration
- Onion-model middleware chain: `applyMiddleware` composes handlers in reverse order
- `Context.Next()` — calls next handler in chain
- `Context.SetNext(fn)` — sets the next handler
- `Context.Abort()` — stops middleware chain propagation
- `Context.IsAborted()` — checks if chain was aborted
- `Context.AbortWithError(code, err)` — aborts with HTTP error
- Route-level middleware: `app.GET("/path", handler, mw1, mw2)`
- Middleware executes in correct onion order: outermost first, innermost last

---

## [0.0.7] — 2026-04-13

### Added

- `errors.RudraError` struct with `Code`, `Message`, `Detail`, `Cause` fields
- HTTP error constructors: `BadRequest(400)`, `Unauthorized(401)`, `Forbidden(403)`, `NotFound(404)`, `MethodNotAllowed(405)`, `Conflict(409)`, `UnprocessableEntity(422)`, `TooManyRequests(429)`, `InternalServerError(500)`
- `errors.NewHTTPError(code, msg)` — generic HTTP error constructor
- `errors.DefaultErrorHandler(c, err)` — JSON error response, never leaks internals
- `errors.As` / `errors.Is` — stdlib-compatible error comparison
- `Engine.SetErrorHandler(fn)` — custom error handler override
- `Context.Abort()` / `Context.AbortWithError(code, err)` / `Context.IsAborted()`
- Panic recovery in `Engine.ServeHTTP` — returns 500 JSON, never exposes stack traces
- All error paths return valid JSON responses

---

## [0.0.6] — 2026-04-13

### Added

- `render.JSON(w, code, v)` — JSON response with `Content-Type: application/json; charset=utf-8`
- `render.Text(w, code, s)` — plain text response
- `render.HTML(w, code, html)` — HTML response
- `render.Blob(w, code, contentType, data)` — binary response
- `render.XML(w, code, v)` — XML response with prolog
- `render.Stream(w, code, contentType, fn)` — chunked transfer streaming
- `render.JSONP(w, code, callback, v)` — JSONP response
- `Context.JSON(code, v)` — JSON response shortcut
- `Context.String(code, s)` — text response shortcut
- `Context.HTML(code, html)` — HTML response shortcut
- `Context.Blob(code, contentType, data)` — binary response shortcut
- `Context.XML(code, v)` — XML response shortcut
- `Context.Stream(code, contentType, fn)` — streaming response shortcut
- `Context.JSONP(code, callback, v)` — JSONP response shortcut
- `Context.NoContent()` — 204 No Content
- `Context.Redirect(code, url)` — HTTP redirect
- `Context.BindJSON(v)` — JSON request body binding
- `Context.BindXML(v)` — XML request body binding
- `Context.FormFile(name)` — multipart file upload
- `Context.FormValue(name)` — form value accessor
- `Context.MultipartForm()` — multipart form accessor

### Performance

- JSON renders directly to `http.ResponseWriter` — zero intermediate buffer

---

## [0.0.5] — 2026-04-13

### Added

- Full radix tree router (`router/tree.go`, `router/router.go`)
- `:param` URL parameter capture — zero heap allocation via `[16]Param` fixed array
- `*wildcard` path capture — captures entire remaining path
- Priority ordering: static > param > wildcard at each node level
- Common prefix splitting for efficient tree structure
- Static route fast path: O(1) `map[string]map[string]HandlerFunc` before radix tree search
- `Router.Find(method, path, ctx)` — route matching with param population
- `Router.Add(method, path, handler, middleware...)` — route registration with conflict detection
- `Router.Name(method, path, name)` — named route registration
- `Router.URL(name, params...)` — URL generation from named routes
- `Router.HasMethod(method)` — check if any route exists for method
- `Router.AllMethods()` — list all registered HTTP methods

### Performance

- Router static lookup: **25 ns/op, 0 allocs/op** (`go test -bench -benchmem`)
- Router 3-param lookup: **69 ns/op, 0 allocs/op**
- Router wildcard lookup: **34 ns/op, 0 allocs/op**
- Live server `wrk` (4 threads, 100 conn, 10s): static **206,686 req/sec** @ 688µs avg latency
- Live server `wrk` param route: **189,005 req/sec** @ 708µs avg latency
- Live server `ab` (100 concurrency, 100k req): static **26,763 req/sec**, 0 failed requests
- Parameterized routing only ~8% slower than static — radix tree overhead is negligible

---

## [0.0.4] — 2026-04-13

### Added

- Static route map (`map[string]map[string]HandlerFunc`) — O(1) lookup
- `Engine.GET`, `Engine.POST`, `Engine.PUT`, `Engine.PATCH`, `Engine.DELETE`, `Engine.OPTIONS`, `Engine.HEAD` — route registration
- `Engine.Any` — register handler for all HTTP methods
- 404 handling — returns `errors.NotFound` via global error handler
- Route conflict detection — panics on duplicate registration
- Static routes bypass radix tree entirely for maximum speed

---

## [0.0.3] — 2026-04-13

### Added

- `context.Context` struct — per-request state container
- `[16]Param` fixed array — zero heap allocation for path parameters
- `sync.Pool` context recycling in Engine — zero-alloc context acquisition
- `Context.Reset(w, r)` — O(1) field reset for pool reuse
- `Context.Release()` — clears sensitive references before pool return
- `Context.Method()`, `Context.Path()` — request accessors
- `Context.Header(key)`, `Context.SetHeader(key, value)` — header accessors
- `Context.RealIP()` — X-Forwarded-For aware IP detection
- `Context.Set(key, val)`, `Context.Get(key)`, `Context.MustGet(key)` — lazy store map
- `Context.Query(key)`, `Context.QueryDefault(key, def)` — query string accessors
- `Context.ContentType()` — Content-Type without parameters
- `Context.Param(key)`, `Context.SetParam(key, value)` — URL parameter accessors
- `Context.Params()` — all captured parameters
- `Context.Request()`, `Context.Writer()` — raw net/http accessors
- `context.Pool` — exported pool management for external use

### Performance

- Context pool acquire/release: **14 ns/op, 0 allocs/op** (`go test -bench -benchmem`)
- Context SetParam: **1.3 ns/op, 0 allocs/op**
- `[16]Param` fixed array eliminates heap allocation for all routes with ≤16 path parameters

---

## [0.0.2] — 2026-04-13

### Added

- `Engine.Run(addr)` — starts `net/http.Server` with configured timeouts
- `Engine.RunTLS(addr, certFile, keyFile)` — HTTPS with TLS 1.2+ hardening
- `Engine.RunListener(l)` — custom `net.Listener` support
- `Engine.Shutdown(ctx)` — graceful connection drain
- `Engine.ListenForShutdown()` — SIGINT/SIGTERM graceful shutdown
- `core.Options` struct with sane defaults:
  - `ReadTimeout: 5s`, `WriteTimeout: 10s`, `IdleTimeout: 120s`, `ReadHeaderTimeout: 2s`
  - `ShutdownTimeout: 30s`, `MaxHeaderBytes: 1MB`, `MaxBodyBytes: 32MB`
- Functional options pattern: `WithReadTimeout`, `WithWriteTimeout`, `WithIdleTimeout`, `WithReadHeaderTimeout`, `WithShutdownTimeout`, `WithMaxHeaderBytes`, `WithMaxBodyBytes`, `WithHTTP2`, `WithStrictRouting`, `WithCaseSensitive`, `WithTLS`
- `Engine` implements `http.Handler` via `ServeHTTP`
- Default 404 error response for unmatched routes
- `examples/hello/main.go` — minimal hello world example

---

## [0.0.1] — 2026-04-13

### Added

- Initialize Go module: `github.com/AarambhDevHub/rudra`
- 10-package workspace layout per `ARCHITECTURE.md §3`
- `go.mod`, `go.sum`, `.gitignore`, `.github/workflows/ci.yml`
- `rudra.go` public API surface with `New() *Engine`
- `core/` — Engine struct, server lifecycle, options, signals
- `router/` — Radix tree router, route groups, named routes
- `context/` — Request/response context with sync.Pool
- `render/` — JSON, Text, HTML, Blob, XML, Stream renderers
- `errors/` — HTTP error types and global error handler
- `binding/` — Binder interface stub
- `validator/` — Validator interface stub
- `config/` — Config struct stub
- `ws/` — WebSocket stub
- `sse/` — SSE stub
- `middleware/` — Middleware stub
- `testutil/` — Request builder, response assertions, recorder
- `benchmarks/` — Router and framework benchmarks
- `examples/hello/`, `examples/rest-api/`
- `LICENSE-MIT`, `LICENSE-APACHE` dual license
- `ARCHITECTURE.md`, `ROADMAP.md`, `CHANGELOG.md`, `README.md`
- `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`

---

[Unreleased]: https://github.com/AarambhDevHub/rudra/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/AarambhDevHub/rudra/compare/v0.0.9...v0.1.0
[0.0.9]: https://github.com/AarambhDevHub/rudra/compare/v0.0.8...v0.0.9
[0.0.8]: https://github.com/AarambhDevHub/rudra/compare/v0.0.7...v0.0.8
[0.0.7]: https://github.com/AarambhDevHub/rudra/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/AarambhDevHub/rudra/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/AarambhDevHub/rudra/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/AarambhDevHub/rudra/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/AarambhDevHub/rudra/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/AarambhDevHub/rudra/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/AarambhDevHub/rudra/releases/tag/v0.0.1