# Changelog

All notable changes to Rudra (रुद्र) are documented here.

This project adheres to [Semantic Versioning](https://semver.org) and [Conventional Commits](https://www.conventionalcommits.org).

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

---

## [Unreleased]

> Phase 2 — Binding & Validation (`0.2.0` → `0.2.9`)

### Planned

- `0.2.0` — `binding.BindJSON` — JSON request body binding
- `0.2.1` — `binding.BindForm` + `binding.BindMultipart` — form data binding
- `0.2.2` — `binding.BindQuery` + `binding.BindPath` + `binding.BindHeader`

---

## [0.1.9] — 2026-04-22

### Added

- `middleware.CSRF(config ...CSRFConfig) HandlerFunc` — CSRF protection
  - Double-submit cookie pattern with `crypto/rand` token generation
  - Constant-time comparison via `crypto/subtle` (timing-attack safe)
  - Configurable: `TokenLength`, `CookieName`, `HeaderName`, `FormField`, `Secure`, `SameSite`, `MaxAge`
  - Safe methods (GET, HEAD, OPTIONS, TRACE) automatically skipped
  - Token stored on context (`c.Set("csrf_token", ...)`) for template rendering
  - Token refreshed after every successful unsafe-method validation
- `middleware.ETag(config ...ETagConfig) HandlerFunc` — ETag generation
  - SHA-256 hash of response body (first 16 bytes) for ETag computation
  - Returns `304 Not Modified` on `If-None-Match` header match
  - Supports weak ETags (`W/"..."`) via `ETagConfig.Weak`
  - Consistent hashing: same content → same ETag across requests
- `Engine.Static(prefix, root string)` — serve files from directory
- `Engine.StaticFile(path, file string)` — serve single file
- `Engine.StaticFS(prefix string, fs http.FileSystem)` — serve from `http.FS`/`embed.FS`
- Directory listing disabled for security (serves `index.html` only)
- Directory traversal protection (blocks `..` paths)
- 13 unit tests covering CSRF + ETag scenarios

---

## [0.1.8] — 2026-04-22

### Added

- `middleware.Compress(config ...CompressConfig) HandlerFunc` — gzip compression
  - Stdlib `compress/gzip` — zero external dependencies
  - `Accept-Encoding` negotiation with `Content-Encoding: gzip` response
  - `CompressConfig.Level` — gzip level (default: `gzip.DefaultCompression`)
  - `CompressConfig.MinLength` — skip small responses (default: 1024 bytes)
  - `CompressConfig.ContentTypes` — prefix-matched content type filter
  - `sync.Pool`’d gzip writers for minimal allocations
  - `Vary: Accept-Encoding` header always set
  - Skips 204/304 responses and pre-compressed content
  - Implements `http.Flusher` interface for SSE compatibility
- 5 unit tests covering compression, skipping, and header behavior

---

## [0.1.7] — 2026-04-22

### Added

- `middleware.RateLimit(config ...RateLimitConfig) HandlerFunc` — token bucket rate limiter
  - Per-key (default: IP) in-memory token bucket algorithm
  - `RateLimitConfig.Rate` — tokens per second (default: 10)
  - `RateLimitConfig.Burst` — bucket capacity (default: 20)
  - `RateLimitConfig.KeyFunc` — custom key extraction (default: `c.RealIP()`)
  - `RateLimitConfig.OnLimit` — custom rate-limited response handler
  - `RateLimitConfig.ExpiresIn` — idle bucket cleanup interval (default: 5m)
  - `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` headers on every response
  - `Retry-After` header on 429 Too Many Requests
  - Background goroutine for periodic expired bucket cleanup
  - `sync.Map` for lock-free per-key bucket lookup
- 5 unit tests covering limits, custom keys, and custom handlers

---

## [0.1.6] — 2026-04-22

### Added

- `middleware.BodyLimit(config ...BodyLimitConfig) HandlerFunc` — request body size limiter
  - Uses `http.MaxBytesReader` (stdlib, zero-alloc)
  - Returns `413 Payload Too Large` when body exceeds limit
  - `BodyLimitConfig.Limit` — max bytes (default: 32MB)
  - `BodyLimitConfig.OnLimit` — custom over-limit handler
  - Skips bodyless methods (GET, HEAD, OPTIONS)
- `middleware.Secure(config ...SecureConfig) HandlerFunc` — security headers
  - `X-Content-Type-Options: nosniff` (default: enabled)
  - `X-Frame-Options: DENY` (configurable: `SAMEORIGIN`, etc.)
  - `X-XSS-Protection: 1; mode=block`
  - `Strict-Transport-Security` with `max-age`, `includeSubDomains`, `preload` (HTTPS only)
  - `Content-Security-Policy` header
  - `Referrer-Policy` header (default: `strict-origin-when-cross-origin`)
  - `Permissions-Policy` header
  - Pre-computed HSTS value — zero per-request string formatting
- `middleware.SecureRedirect(httpsPort int)` — HTTP→HTTPS redirect middleware
- 12 unit tests covering body limits, all security headers, HSTS proxy detection

---

## [0.1.5] — 2026-04-22

### Added

- `middleware.CORS(config ...CORSConfig) HandlerFunc` — full CORS middleware
- Handles both simple requests and preflight `OPTIONS` requests
- `CORSConfig.AllowOrigins []string` — exact match + wildcard `"*"`
- `CORSConfig.AllowMethods []string` — configurable allowed methods
- `CORSConfig.AllowHeaders []string` — configurable allowed headers
- `CORSConfig.ExposeHeaders []string` — headers exposed to browser
- `CORSConfig.AllowCredentials bool` — credential support (blocks wildcard origin)
- `CORSConfig.MaxAge int` — preflight cache duration in seconds (default 24h)
- `CORSConfig.AllowOriginFunc func(origin string) bool` — dynamic origin validation
- `DefaultCORSConfig()` — permissive defaults for development
- Pre-computed header strings for zero per-request string allocations
- Preflight returns `204 No Content` without proceeding to handler
- `Vary: Origin` header set when origin is not wildcard
- 10 unit tests covering all CORS scenarios

---

## [0.1.4] — 2026-04-22

### Added

- `middleware.Timeout(config ...TimeoutConfig) HandlerFunc` — per-request timeout
- Wraps `r.Context()` with `context.WithTimeout` — compatible with DB/HTTP client timeouts
- `TimeoutConfig.Timeout time.Duration` — default 30s
- `TimeoutConfig.OnTimeout func(*Context) error` — custom timeout handler
- Returns `503 Service Unavailable` on timeout (via `errors.NewHTTPError`)
- Goroutine-based handler execution with `select` on done/timeout channels
- Re-panics recovered panics for upstream Recovery middleware to handle
- `Context.SetRequest(r *http.Request)` — added to context for request replacement
- 5 unit tests including fast/slow handlers, custom handler, and context propagation

---

## [0.1.3] — 2026-04-22

### Added

- `middleware.RequestID(config ...RequestIDConfig) HandlerFunc` — unique request ID per request
- UUID v4 generation via `crypto/rand` — cryptographically secure
- Reads `X-Request-ID` from incoming request first (forwarded from upstream proxy)
- Sets `X-Request-ID` on response header
- Stores on context: `c.Set("request_id", id)` + `c.SetRequestID(id)`
- `RequestIDConfig.Generator func() string` — custom ID generator
- `RequestIDConfig.Header string` — custom header name (default `"X-Request-ID"`)
- `DefaultRequestIDConfig()` — sane defaults
- 7 unit tests including uniqueness, forwarding, custom generator, UUID v4 format validation

---

## [0.1.2] — 2026-04-22

### Added

- `middleware.Recovery(config ...RecoveryConfig) HandlerFunc` — panic recovery middleware
- `defer/recover` wraps the inner middleware chain
- Captures panic value + full stack trace via `runtime/debug.Stack()`
- Logs stack trace via `log/slog` JSON handler (never sent to client)
- Returns `500 Internal Server Error` JSON to client
- `RecoveryConfig.LogStackTrace bool` — toggle stack trace logging (default true)
- `RecoveryConfig.Output io.Writer` — stack trace output destination (default os.Stderr)
- `RecoveryConfig.OnPanic func(c, err, stack)` — custom panic hook
- `DefaultRecoveryConfig()` — sane defaults
- 6 unit tests: panic handling, no-panic passthrough, OnPanic hook, disabled logging, error propagation

---

## [0.1.1] — 2026-04-22

### Added

- `middleware.Logger(config ...LoggerConfig) HandlerFunc` — structured access logging
- Logs: method, path, status, latency, IP, request_id, user_agent, bytes_written
- Three formats: `"json"` (slog JSON handler), `"text"` (slog text handler), `"common"` (Apache Combined Log)
- `LoggerConfig.SkipPaths []string` — omit health/metrics routes (O(1) map lookup)
- `LoggerConfig.Output io.Writer` — defaults to `os.Stdout`
- `LoggerConfig.TimeFormat string` — configurable time format
- `LoggerConfig.Level slog.Level` — configurable log level
- Latency measured around `c.Next()` — nanosecond-accurate
- Uses `log/slog` (Go 1.21+) for structured output
- `middleware/response_writer.go` — intercepting `http.ResponseWriter` wrapper
  - Tracks `statusCode` and `bytesWritten` across all Write calls
  - Implements `http.Flusher`, `http.Hijacker`, `http.Pusher` for full compatibility
  - Double `WriteHeader` calls safely ignored
- `Context.SetWriter(w http.ResponseWriter)` — replace response writer for interception
- `Context.UserAgent() string` — convenience accessor for User-Agent header
- `DefaultLoggerConfig()` — sane defaults
- 8 unit tests covering all formats, skip paths, bytes/status tracking, writer interfaces

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

[Unreleased]: https://github.com/AarambhDevHub/rudra/compare/v0.1.9...HEAD
[0.1.9]: https://github.com/AarambhDevHub/rudra/compare/v0.1.8...v0.1.9
[0.1.8]: https://github.com/AarambhDevHub/rudra/compare/v0.1.7...v0.1.8
[0.1.7]: https://github.com/AarambhDevHub/rudra/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/AarambhDevHub/rudra/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/AarambhDevHub/rudra/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/AarambhDevHub/rudra/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/AarambhDevHub/rudra/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/AarambhDevHub/rudra/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/AarambhDevHub/rudra/compare/v0.1.0...v0.1.1
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