# Rudra (रुद्र) — Development Roadmap

> **"Fierce. Fast. Fearless."**
> Full development roadmap from `0.0.1` to `0.9.9`
> All features sourced from `ARCHITECTURE.md`

---

## Versioning Philosophy

```
0 . X . Y
│   │   └─ Patch: bug fixes, performance tweaks, doc updates
│   └───── Minor: new feature group or subsystem complete
└───────── Major: breaking API change (staying at 0 until stable)
```

- Every `0.0.x` version is a single, shippable atomic unit of work
- No version is skipped — every release is tagged on GitHub
- `0.9.9` is the final pre-API-freeze release
- Public launch happens at a chosen `0.x.0` milestone (team decision)
- `1.0.0` is reserved for long-term stable API guarantee only

---

## Phase 0 — Foundation (0.0.1 → 0.0.9) ✅ COMPLETE

> Goal: A working HTTP server with basic routing, context, and rendering.
> By `0.0.9`: `go get` → write a handler → get a JSON response. Nothing more.
> **Status: All 9 versions completed and tested.**

---

### `0.0.1` — Project Scaffold ✅
- Initialize Go module: `github.com/AarambhDevHub/rudra`
- Create 10-package workspace layout per ARCHITECTURE §3
- `go.mod`, `go.sum`, `.gitignore`, `LICENSE-MIT`, `LICENSE-APACHE`
- Empty stub files for all packages (compilable, no logic)
- `rudra.go` public API surface with `New()` returning `*Engine`
- `Engine` struct skeleton (no routing, no server yet)
- `README.md` placeholder with build badge
- CI: GitHub Actions `go build ./...` + `go vet ./...`

**Definition of Done:** `go build ./...` passes. Zero logic — just structure.

---

### `0.0.2` — HTTP/1.1 Server Bootstrap ✅

**Deliverables:**
- `Engine.Run(addr string) error` — starts `net/http.Server`
- Default `Options` struct with sane timeouts (ReadTimeout 5s, WriteTimeout 10s, IdleTimeout 120s, ReadHeaderTimeout 2s)
- Functional options pattern: `WithReadTimeout`, `WithWriteTimeout`, `WithIdleTimeout`, `WithMaxHeaderBytes`
- `Engine` implements `http.Handler` via `ServeHTTP(w, r)`
- Placeholder handler: all requests return `200 OK` with `"rudra is alive"`
- Example: `examples/hello/main.go`

**Definition of Done:** `curl localhost:8080` returns `"rudra is alive"`.

---

### `0.0.3` — Context System + Pool ✅

**Deliverables:**
- `context.Context` struct (full implementation per ARCHITECTURE §10)
- `[maxParams]Param` fixed array — zero heap allocation for path params
- `sync.Pool` in Engine for context recycling
- `Context.Reset(w, r)` — O(1) field reset
- `Context.Release()` — clears sensitive refs before pool return
- `Context.Method()`, `Context.Path()`, `Context.Header()`, `Context.SetHeader()`
- `Context.RealIP()` — X-Forwarded-For aware
- `Context.Set()` / `Context.Get()` / `Context.MustGet()` — lazy store map
- `Context.Query()` / `Context.QueryDefault()`
- `context/pool.go` — exported pool management

**Definition of Done:** Benchmark `BenchmarkContextAcquireRelease` shows 0 allocs/op.

---

### `0.0.4` — Static Route Router ✅ (Map-based)

**Deliverables:**
- `router.Router` struct with `map[string]map[string]HandlerFunc` for static routes
- `Router.Add(method, path string, h HandlerFunc)` for static paths only
- `Router.Find(method, path string, c *context.Context) HandlerFunc`
- `Engine.GET`, `Engine.POST`, `Engine.PUT`, `Engine.PATCH`, `Engine.DELETE`, `Engine.OPTIONS`, `Engine.HEAD` registration methods
- Route conflict detection: panic on duplicate registration
- `Engine.ServeHTTP` dispatches to matched handler
- 404 → `errors.NotFound` returned to global error handler
- 405 → `errors.MethodNotAllowed` with `Allow` header

**Definition of Done:** Static routes resolve correctly. BenchmarkRouterStatic: 0 allocs/op.

---

### `0.0.5` — Radix Tree Router (Dynamic Routes) ✅

**Deliverables:**
- `router/tree.go` — full radix tree per ARCHITECTURE §9.2
- `router/node.go` — node struct with `indices []byte` fast-lookup
- `:param` capture — stored into `Context.params` fixed array
- `*wildcard` capture — captures rest of path
- Mixed routes: `/users/:id/posts/:postId` works correctly
- Priority: static > param > wildcard at each node level
- `Router.Find` checks static map first (O(1)), falls back to radix tree
- Full conflict detection for ambiguous param routes
- `Context.Param(key string) string`
- `Context.SetParam(key, value string)` — called by router, not user

**Definition of Done:** BenchmarkRouterParams (3 params): 0 allocs/op.

---

### `0.0.6` — Response Rendering ✅

**Deliverables:**
- `render` package with `Render` interface
- `render.JSON(w, code, v)` — uses `encoding/json`, writes directly to `ResponseWriter`
- `render.Text(w, code, s)` — `text/plain`
- `render.HTML(w, code, html)` — `text/html`
- `render.Blob(w, code, contentType, data)` — binary
- `render.NoContent(w)` — 204
- Context methods: `c.JSON()`, `c.String()`, `c.HTML()`, `c.Blob()`, `c.NoContent()`
- `c.Redirect(code, url)` — wraps `http.Redirect`
- Content-Type set automatically for all renderers

**Definition of Done:** JSON response benchmark: 1 alloc/op (the JSON encoding itself).

---

### `0.0.7` — Error Handling System ✅

**Deliverables:**
- `errors.RudraError` struct — `Code int`, `Message string`, `Detail any`, `Cause error`
- `errors.RudraError.Error() string` implements `error`
- HTTP error constructors: `BadRequest`, `Unauthorized`, `Forbidden`, `NotFound`, `Conflict`, `UnprocessableEntity`, `TooManyRequests`, `InternalServerError`
- `errors.DefaultErrorHandler(c, err)` — JSON error response, never leaks internals
- `Engine.SetErrorHandler(fn)` — custom override
- `Context.AbortWithError(code, err)` — stops chain, returns error
- `Context.Abort()` / `Context.IsAborted() bool`
- `errors.Is` / `errors.As` compatibility on `RudraError`

**Definition of Done:** All error paths return valid JSON. Panic in handler → 500 JSON (not crash).

---

### `0.0.8` — Middleware Chain ✅

**Deliverables:**
- `HandlerFunc` type: `func(*context.Context) error`
- `Engine.Use(middleware ...HandlerFunc)` — global middleware registration
- `applyMiddleware` — builds composed handler chain (onion model)
- `Context.SetNext(fn)` / `Context.Next() error`
- Middleware executes: before-handler code → `c.Next()` → after-handler code
- Short-circuit: `c.Abort()` stops chain propagation
- Route-level middleware: `app.GET("/path", handler, mw1, mw2)`
- `router/group.go` — `Group` struct with prefix + shared middleware

**Definition of Done:** 3-layer middleware chain in correct order. Abort stops at correct layer.

---

### `0.0.9` — Route Groups + Named Routes ✅

**Deliverables:**
- `Engine.Group(prefix string, mw ...HandlerFunc) *Group`
- `Group.Group(prefix string, mw ...HandlerFunc) *Group` — nested groups
- `Group.GET`, `Group.POST`, `Group.PUT`, `Group.PATCH`, `Group.DELETE`
- `Group.Use(mw ...HandlerFunc)` — add middleware to group after creation
- Named routes: `app.GET("/users/:id", handler).Name("user.profile")`
- `router.URL("user.profile", "42")` → `"/users/42"` (URL generation)
- `namedRoutes map[string]string` registry in Router
- Example: `examples/rest-api/main.go` — CRUD API with groups

**Definition of Done:** Nested groups work. URL generation works for param routes.

---

## Phase 1 — Full HTTP/1.1 + Core Middleware (0.1.0 → 0.1.9) 🔄 IN PROGRESS

> Goal: Production-grade HTTP/1.1 with all essential middleware.
> By `0.1.9`: deployable for real workloads.
> **Status: 0.1.0 complete. 0.1.1–0.1.9 planned.**

---

### `0.1.0` — HTTP/1.1 Hardening + Graceful Shutdown ✅

**Deliverables:**
- Custom TCP listener with `SO_REUSEPORT` + `TCP_NODELAY` + `TCP_FASTOPEN` (Linux) ✅
- `Engine.RunListener(net.Listener)` — custom listener support ✅
- `core/signals.go` — `Engine.ListenForShutdown()` (SIGINT + SIGTERM) ✅
- `Engine.Shutdown(ctx context.Context) error` — graceful drain (idempotent, race-free) ✅
- `ShutdownTimeout` option (default 30s) ✅
- `Engine.RunTLS(addr, cert, key)` — HTTPS with hardened AEAD cipher suites + session resumption ✅
- TLS cipher suite hardening per ARCHITECTURE §20.2 ✅
- `examples/hello/main.go` updated with graceful shutdown pattern ✅
- `WithTCPNoDelay()`, `WithSOReusePort()`, `WithTCPFastOpen()` functional options ✅
- Platform-specific build tags: `server_linux.go` / `server_other.go` ✅
- Zero-allocation hot path preserved — all micro-benchmarks: 0 allocs/op ✅

**Definition of Done:** `kill -SIGTERM` drains active connections. Zero dropped requests on clean shutdown. ✅

**wrk benchmarks (4 threads, 100 conn, 10s, i5-1135G7):**
- Static: **213,817 req/sec** @ 671µs avg latency
- Param: **178,407 req/sec** @ 783µs avg latency

---

### `0.1.1` — Logger Middleware

**Deliverables:**
- `middleware.Logger(config LoggerConfig) HandlerFunc`
- Logs: method, path, status, latency, IP, request_id, user_agent, bytes_written
- Formats: `"json"` (default), `"text"`, `"common"` (Apache Combined Log)
- `LoggerConfig.SkipPaths []string` — omit health/metrics routes
- `LoggerConfig.Output io.Writer` — defaults to `os.Stdout`
- `LoggerConfig.TimeFormat string`
- Latency measured around `c.Next()` — accurate to nanosecond
- Uses `log/slog` (Go 1.21+) for structured output

**Definition of Done:** JSON log line produced for every request. Skipped paths produce no output.

---

### `0.1.2` — Recovery Middleware

**Deliverables:**
- `middleware.Recovery(config RecoveryConfig) HandlerFunc`
- `defer/recover` wraps the inner chain
- Captures panic value + full stack trace (`runtime/debug.Stack()`)
- Logs stack trace to configured writer (never to client)
- Returns `500 Internal Server Error` JSON to client (no stack in response)
- `RecoveryConfig.LogStackTrace bool` (default true)
- `RecoveryConfig.OnPanic func(c, err, stack)` — custom hook

**Definition of Done:** `panic("test")` in handler → 500 JSON to client, stack in logs, server continues.

---

### `0.1.3` — RequestID Middleware

**Deliverables:**
- `middleware.RequestID(config RequestIDConfig) HandlerFunc`
- Generates UUID v4 per request (using `crypto/rand`)
- Reads `X-Request-ID` from incoming request first (forwarded ID)
- Sets `X-Request-ID` on response header
- Stores ID on context: `c.Set("request_id", id)`
- `Context.RequestID() string` convenience method
- `RequestIDConfig.Generator func() string` — custom generator
- `RequestIDConfig.Header string` — default `"X-Request-ID"`

**Definition of Done:** Every response has unique `X-Request-ID`. Forwarded IDs are preserved.

---

### `0.1.4` — Timeout Middleware

**Deliverables:**
- `middleware.Timeout(config TimeoutConfig) HandlerFunc`
- Per-request `context.WithTimeout` wrapping the handler
- Returns `503 Service Unavailable` JSON on timeout
- `TimeoutConfig.Timeout time.Duration` (default 30s)
- `TimeoutConfig.OnTimeout func(c) error` — custom handler
- Uses `context.WithDeadline` propagated through `r.WithContext`
- Compatible with downstream DB/HTTP client timeouts

**Definition of Done:** Handler that sleeps 10s with 5s timeout → 503 after 5s.

---

### `0.1.5` — CORS Middleware

**Deliverables:**
- `middleware.CORS(config CORSConfig) HandlerFunc`
- Handles simple requests + preflight `OPTIONS` requests
- `CORSConfig.AllowOrigins []string` — exact match + wildcard `"*"`
- `CORSConfig.AllowMethods []string`
- `CORSConfig.AllowHeaders []string`
- `CORSConfig.ExposeHeaders []string`
- `CORSConfig.AllowCredentials bool`
- `CORSConfig.MaxAge int` — preflight cache in seconds
- `middleware.DefaultCORSConfig()` — permissive defaults for dev
- Dynamic origin validation via `CORSConfig.AllowOriginFunc`

**Definition of Done:** Preflight returns correct headers. Credentials + wildcard origin blocked correctly.

---

### `0.1.6` — Body Limit + Secure Headers Middleware

**Deliverables:**
- `middleware.BodyLimit(limit int64) HandlerFunc`
  - Wraps `r.Body` with `io.LimitReader`
  - Returns `413 Payload Too Large` when exceeded
- `middleware.Secure(config SecureConfig) HandlerFunc`
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY` (or `SAMEORIGIN`)
  - `X-XSS-Protection: 1; mode=block`
  - `Strict-Transport-Security` with configurable `max-age`
  - `Content-Security-Policy` header
  - `Referrer-Policy` header
  - `Permissions-Policy` header
  - `SecureConfig.HSTSPreload bool`
  - `SecureConfig.HSTSIncludeSubdomains bool`

**Definition of Done:** 100MB POST to body-limited route → 413. Security headers present on all responses.

---

### `0.1.7` — Rate Limiter Middleware

**Deliverables:**
- `middleware.RateLimit(config RateLimitConfig) HandlerFunc`
- Token bucket algorithm (in-memory, per key)
- `RateLimitConfig.Rate float64` — tokens per window
- `RateLimitConfig.Burst int` — max burst
- `RateLimitConfig.Window time.Duration`
- `RateLimitConfig.KeyFunc func(*context.Context) string` — default: RealIP
- `RateLimitConfig.OnLimit func(*context.Context) error`
- Sets `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` headers
- Periodic cleanup of expired keys (background goroutine)
- `Retry-After` header on 429 responses

**Definition of Done:** 100 req/min limit enforced. 101st request → 429 with correct headers.

---

### `0.1.8` — Compression Middleware

**Deliverables:**
- `middleware.Compress(config CompressConfig) HandlerFunc`
- Supports `gzip` (stdlib), `brotli` (`github.com/andybalholm/brotli`), `zstd` (`github.com/klauspost/compress/zstd`)
- Algorithm selected via `Accept-Encoding` header, preference ordered
- `CompressConfig.Level int` — compression level
- `CompressConfig.MinLength int` — skip compression for small responses (default 1024 bytes)
- `CompressConfig.ContentTypes []string` — only compress matching types
- Compressed writer pooled via `sync.Pool` per algorithm
- Sets `Content-Encoding`, `Vary: Accept-Encoding` headers

**Definition of Done:** JSON response >1KB compressed with correct Content-Encoding. Small responses pass through uncompressed.

---

### `0.1.9` — CSRF + ETag + Static File Server

**Deliverables:**
- `middleware.CSRF(config CSRFConfig) HandlerFunc`
  - Double-submit cookie pattern
  - Token stored in signed cookie, verified from header
  - `CSRFConfig.TokenLength int`, `CookieName`, `HeaderName`, `Secure`, `SameSite`
  - Safe methods (GET, HEAD, OPTIONS) skipped
- `middleware.ETag(config ETagConfig) HandlerFunc`
  - Computes ETag from response body hash
  - Returns 304 on `If-None-Match` match
- `Engine.Static(prefix, root string)` — serve files from directory
- `Engine.StaticFile(path, file string)` — serve single file
- `Engine.StaticFS(prefix string, fs http.FileSystem)` — serve from `http.FS`

**Definition of Done:** CSRF token mismatch → 403. Static file serves with correct Content-Type + ETag.

---

## Phase 2 — Binding & Validation (0.2.0 → 0.2.9)

> Goal: Complete request data binding and struct validation system.

---

### `0.2.0` — JSON Binding

**Deliverables:**
- `binding.BindJSON(c, v any) error`
- Reads body with `io.LimitReader` (respects `MaxBodyBytes`)
- Decodes with `encoding/json`
- Returns `errors.BadRequest` on decode failure
- `Context.BindJSON(v any) error` convenience method
- Body cached on context after first read (re-readable)
- Empty body → `BadRequest` with clear message

**Definition of Done:** POST with malformed JSON → 400. POST with valid JSON → struct populated correctly.

---

### `0.2.1` — Form + Multipart Binding

**Deliverables:**
- `binding.BindForm(c, v any) error` — `application/x-www-form-urlencoded`
- `binding.BindMultipart(c, v any) error` — `multipart/form-data`
- Struct tags: `form:"field_name"`
- Supports nested structs, slices, and pointers
- `Context.FormValue(key string) string`
- `Context.FormFile(key string) (*multipart.FileHeader, error)`
- `Context.MultipartForm() (*multipart.Form, error)`
- `Context.SaveUploadedFile(file, dst string) error` — save to disk

**Definition of Done:** File upload saves correctly. Form fields bind to struct.

---

### `0.2.2` — Query + Path + Header Binding

**Deliverables:**
- `binding.BindQuery(c, v any) error` — struct tags: `query:"name"`
- `binding.BindPath(c, v any) error` — struct tags: `path:"name"`
- `binding.BindHeader(c, v any) error` — struct tags: `header:"name"`
- Type coercion: `string → int, float64, bool, time.Time` for all binders
- Slice support: `?tags=a&tags=b` → `Tags []string`
- Comma-split: `?tags=a,b,c` → `Tags []string`
- `Context.BindQuery(v any) error`, `Context.BindPath(v any) error`

**Definition of Done:** `/users?page=2&limit=10` binds to `Pagination{Page: 2, Limit: 10}`.

---

### `0.2.3` — XML Binding + Rendering

**Deliverables:**
- `binding.BindXML(c, v any) error` — `application/xml` or `text/xml`
- `render.XML(w, code, v)` — XML response renderer
- `Context.BindXML(v any) error`
- `Context.XML(code int, v any) error`
- Proper Content-Type: `application/xml; charset=utf-8`
- XML prolog included in response

**Definition of Done:** XML request binds. XML response has correct Content-Type.

---

### `0.2.4` — MessagePack Binding + Rendering

**Deliverables:**
- `binding.BindMsgpack(c, v any) error` — `application/msgpack`
- `render.Msgpack(w, code, v)` — MessagePack response
- `Context.BindMsgpack(v any) error`
- `Context.Msgpack(code int, v any) error`
- Uses `github.com/vmihaiela/msgpack/v5`
- Content-Type: `application/msgpack`

**Definition of Done:** MessagePack round-trip (bind + render) works correctly.

---

### `0.2.5` — Validator Core (Basic Rules)

**Deliverables:**
- `validator` package with `Validate(v any) error` function
- Struct tag: `rudra:"rule1,rule2,rule3"`
- Rules implemented: `required`, `min=N`, `max=N`, `email`, `url`
- Error type: `ValidationErrors` — slice of `FieldError{Field, Rule, Value, Message}`
- Reflection-based struct traversal (startup cost, not per-request)
- Cached struct metadata for performance (reflect only once per type)
- `Context.Validate(v any) error`

**Definition of Done:** Invalid email fails `email` rule. Missing required field caught.

---

### `0.2.6` — Validator Extended Rules

**Deliverables:**
- Additional rules: `uuid`, `len=N`, `oneof=a b c`, `alphanum`, `numeric`, `regexp=pattern`
- Nested struct validation (recursive)
- Slice element validation: `rudra:"dive,required,min=1"`
- Pointer field support (nil pointer → required fails)
- Cross-field validation: `rudra:"eqfield=Password"` / `rudra:"nefield=OldPassword"`
- `min` / `max` on numeric fields (value range) vs string/slice (length)

**Definition of Done:** All rules documented with test coverage >95%.

---

### `0.2.7` — Custom Validator Rules

**Deliverables:**
- `validator.Register(name string, fn func(value string) bool)` — simple rule
- `validator.RegisterWithParam(name string, fn func(value, param string) bool)` — parametric rule
- `validator.RegisterStructLevel(fn StructLevelFunc)` — cross-field struct-level validation
- Example: `validator.Register("indianphone", phoneRegex)`
- Custom error message registration: `validator.RegisterMessage(rule, msg)`
- Thread-safe registration (must be called at startup)

**Definition of Done:** Custom rule `indianphone` works in struct tags.

---

### `0.2.8` — ShouldBind + MustBind Auto-Detection

**Deliverables:**
- `Context.ShouldBind(v any) error` — auto-selects binder from `Content-Type`
- `Context.MustBind(v any) error` — binds + validates, aborts with 400/422 on failure
- `Context.ContentType() string` — normalized Content-Type without params
- Fallback order when Content-Type absent: JSON
- `Context.ShouldBindWith(v any, b binding.Binder) error` — explicit binder
- Integration test: all Content-Types route to correct binder

**Definition of Done:** POST with `application/json` auto-binds to JSON. POST with `multipart/form-data` auto-binds to multipart.

---

### `0.2.9` — Sonic JSON Integration

**Deliverables:**
- Replace `encoding/json` with `github.com/bytedance/sonic` in `render/json.go` and `binding/json.go`
- Build tag `//go:build !nosonic` — fallback to stdlib with `nosonic` tag
- `sonic.ConfigFastest` for responses, `sonic.ConfigDefault` for binding
- Encoder pooled (or use sonic's built-in pool)
- Zero intermediate buffer: encode directly to `ResponseWriter`
- Benchmark: `BenchmarkJSONRender` shows improvement over stdlib

**Definition of Done:** JSON render benchmark shows ≥40% speedup vs stdlib on large payloads.

---

## Phase 3 — HTTP/2 (0.3.0 → 0.3.5)

> Goal: Full HTTP/2 support — TLS, h2c, and server push.

---

### `0.3.0` — HTTP/2 over TLS

**Deliverables:**
- `Engine.RunTLS` updated to configure `http2.Server` when `WithHTTP2()` set
- `http2.ConfigureServer(e.server, e.h2server)` integration
- `Options.HTTP2Enabled bool`, `HTTP2MaxConcurrentStreams uint32`, `HTTP2MaxReadFrameSize uint32`, `HTTP2IdleTimeout`
- `WithHTTP2()` functional option
- ALPN negotiation: `h2` + `http/1.1` in TLS `NextProtos`
- Test: verify connection is HTTP/2 via `r.Proto`
- Example: `examples/http2/main.go`

**Definition of Done:** `curl --http2 https://localhost:8443` succeeds with `h2` protocol.

---

### `0.3.1` — h2c Plaintext HTTP/2

**Deliverables:**
- `Engine.RunH2C(addr string) error`
- `h2c.NewHandler(e, h2s)` wraps Engine for plaintext HTTP/2
- `WithH2C()` functional option
- Upgrade path: HTTP/1.1 request with `Upgrade: h2c` header handled
- Test: `curl --http2-prior-knowledge http://localhost:8080` succeeds
- Example: `examples/http2/h2c/main.go`
- Use case documentation: internal microservices, no TLS overhead

**Definition of Done:** h2c connection established. Requests multiplexed over single TCP conn.

---

### `0.3.2` — HTTP/2 Server Push

**Deliverables:**
- `Context.Push(target string, opts *http.PushOptions) error`
- Graceful degradation: no-op on HTTP/1.1 (no error returned)
- `http.Pusher` interface type assertion on ResponseWriter
- Push-compatible default headers (Content-Type)
- Example: `examples/http2/push/main.go` — HTML + pushed CSS/JS
- Benchmark: measure time-to-first-byte improvement with push

**Definition of Done:** Browser receives pushed assets before requesting them. HTTP/1.1 handler works unchanged.

---

### `0.3.3` — HTTP/2 Configuration Tuning

**Deliverables:**
- `Options.HTTP2MaxHandlers int`
- `Options.HTTP2MaxUploadBufferPerStream int32`
- `Options.HTTP2MaxUploadBufferPerConnection int32`
- Flow control window size configuration
- Stream prioritization: accept Priority frames
- `GOAWAY` frame sent on graceful shutdown
- `RST_STREAM` on individual stream errors
- Connection-level vs stream-level error separation

**Definition of Done:** 250 concurrent streams handled. Connection idle timeout respected.

---

### `0.3.4` — HTTP/2 + WebSocket Upgrade Co-existence

**Deliverables:**
- Ensure WebSocket upgrade (HTTP/1.1 Upgrade) still works on HTTP/2-enabled server
- HTTP/2 extended CONNECT method (RFC 8441) groundwork
- Mixed protocol test: one client on h2, another upgrading to WS simultaneously
- Confirm no interference between h2 multiplexed streams and WS connection

**Definition of Done:** Simultaneous HTTP/2 + WebSocket connections work without interference.

---

### `0.3.5` — HTTP/2 Benchmark Pass

**Deliverables:**
- `benchmarks/framework_bench_test.go` extended for HTTP/2 paths
- `wrk2` or `h2load` benchmark script for HTTP/2 throughput
- Target: match or beat Gin's HTTP/1.1 throughput on HTTP/2
- Profile HTTP/2 hot paths: HPACK, frame parsing overhead
- Document HTTP/2 performance characteristics in `docs/http2-performance.md`

**Definition of Done:** HTTP/2 benchmark results documented. No regression vs HTTP/1.1.

---

## Phase 4 — WebSocket (0.4.0 → 0.4.6)

> Goal: Production WebSocket with hub, rooms, compression, and pooled writes.

---

### `0.4.0` — WebSocket Upgrader (RFC 6455)

**Deliverables:**
- `ws/upgrader.go` — full RFC 6455 handshake per ARCHITECTURE §7.2
- `Upgrader.Upgrade(w, r) (*Conn, error)`
- SHA-1 accept key computation: `key + GUID` → base64
- `Hijacker` interface for raw TCP takeover
- `101 Switching Protocols` response written manually
- `Upgrader.CheckOrigin func(r) bool` — origin validation
- `Upgrader.HandshakeTimeout time.Duration`
- Reject non-GET requests → 405

**Definition of Done:** WebSocket handshake completes. Browser `new WebSocket()` connects successfully.

---

### `0.4.1` — WebSocket Connection (Read/Write/Close)

**Deliverables:**
- `ws/conn.go` — full `Conn` struct per ARCHITECTURE §7.3
- `Conn.ReadMessage() (MessageType, []byte, error)` — frame parser
- `Conn.WriteMessage(type, data) error` — frame writer
- `Conn.WriteJSON(v any) error` — marshal + send
- `Conn.ReadJSON(v any) error` — receive + unmarshal
- `Conn.Close() error` — sends CloseNormalClosure, closes TCP
- `Conn.SetReadDeadline(t time.Time) error`
- `Conn.SetWriteDeadline(t time.Time) error`
- Separate `readMu` + `writeMu`: concurrent read + write safe
- `closeOnce sync.Once` — idempotent close

**Definition of Done:** Concurrent `ReadMessage` and `WriteMessage` goroutines work without data race (`go test -race`).

---

### `0.4.2` — WebSocket Hub (Global Broadcast)

**Deliverables:**
- `ws/hub.go` — Hub struct per ARCHITECTURE §7.4
- `Hub.Run()` — channel-based event loop (goroutine-safe)
- `Hub.Register(id string, conn *Conn)`
- `Hub.Unregister(id string)`
- `Hub.Broadcast(type MessageType, data []byte)`
- `Hub.BroadcastJSON(v any)`
- `Hub.ClientCount() int`
- `Hub.HasClient(id string) bool`
- Non-blocking send to client channel (drop if full — prevents one slow client blocking all)
- Example: `examples/websocket/chat/main.go`

**Definition of Done:** 100 connected clients all receive broadcast within 10ms.

---

### `0.4.3` — WebSocket Rooms

**Deliverables:**
- `Hub.JoinRoom(clientID, room string)`
- `Hub.LeaveRoom(clientID, room string)`
- `Hub.BroadcastToRoom(room string, type MessageType, data []byte)`
- `Hub.RoomClients(room string) []string`
- `Hub.ClientRooms(clientID string) []string`
- Client auto-removed from all rooms on disconnect
- Room auto-deleted when last member leaves
- Example: `examples/websocket/rooms/main.go`

**Definition of Done:** Room broadcast reaches only room members. Client in 2 rooms receives room-specific messages correctly.

---

### `0.4.4` — WebSocket Compression (permessage-deflate)

**Deliverables:**
- `Upgrader.EnableCompression bool`
- `Sec-WebSocket-Extensions: permessage-deflate` handshake
- `server_no_context_takeover` + `client_no_context_takeover` flags
- Per-message flate compression/decompression
- Compression skipped for small messages (`<MinCompressSize int`)
- `flate.NewWriter` pooled via `sync.Pool`
- Benchmark: measure CPU vs bandwidth tradeoff

**Definition of Done:** Compressed text messages measurably smaller on wire. No data corruption.

---

### `0.4.5` — WebSocket Ping/Pong Keepalive

**Deliverables:**
- `Conn.SetPingHandler(fn func(data string) error)`
- `Conn.SetPongHandler(fn func(data string) error)`
- `Conn.SetCloseHandler(fn func(code int, text string) error)`
- Automatic pong response to server pings
- `Upgrader.PingInterval time.Duration` — server-side ping ticker
- Dead connection detection: no pong within `PingTimeout` → close
- Heartbeat example in `examples/websocket/keepalive/main.go`

**Definition of Done:** Client silent for 60s → server detects dead connection and cleans up.

---

### `0.4.6` — WebSocket Write Buffer Pool

**Deliverables:**
- `ws/pool.go` — `sync.Pool` of `[]byte` write buffers
- `Upgrader.WriteBufferPool *sync.Pool` — user-injectable pool
- Framing uses pooled buffer, released after write
- Benchmark: `BenchmarkWebSocketWrite` — 0 allocs/op
- `Upgrader.ReadBufferSize int`, `WriteBufferSize int` options
- WebSocket integration test suite: 1000 concurrent connections

**Definition of Done:** WebSocket write benchmark: 0 allocs/op.

---

## Phase 5 — Server-Sent Events (0.5.0 → 0.5.5)

> Goal: Full SSE with broker, heartbeat, Last-Event-ID reconnect, backpressure.

---

### `0.5.0` — SSE Broker + Basic Events

**Deliverables:**
- `sse/event.go` — `Event` struct + wire format serializer per ARCHITECTURE §8.2
- `sse/broker.go` — `Broker` with `publish` channel + client map
- `Broker.ServeHTTP(w, r)` — holds connection open, streams events
- Required headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `X-Accel-Buffering: no`
- `Broker.Publish(e *Event)` — sends to all clients
- `Broker.ClientCount() int`
- `NewBroker(bufSize int) *Broker`
- Example: `examples/sse/counter/main.go` — incrementing counter pushed to browser

**Definition of Done:** Browser `EventSource` receives events without polling.

---

### `0.5.1` — SSE Heartbeat

**Deliverables:**
- `time.NewTicker(30s)` in client goroutine sends `": heartbeat\n\n"` comment
- Heartbeat interval configurable: `BrokerConfig.HeartbeatInterval time.Duration`
- Heartbeat prevents proxy timeout disconnects (nginx default 60s)
- Heartbeat does not increment event ID
- Heartbeat invisible to `EventSource` `.onmessage` handler (comment line)

**Definition of Done:** Connection stays alive for 10 minutes through nginx proxy with 60s timeout.

---

### `0.5.2` — SSE Last-Event-ID Reconnect

**Deliverables:**
- Read `Last-Event-ID` request header on reconnect
- `Broker.SetReplayBuffer(size int)` — circular buffer of last N events
- On reconnect with `Last-Event-ID`, replay missed events before live stream
- Events with non-empty `ID` field increment ID counter
- `Event.ID` set automatically if not provided: sequential integer
- Example: demonstrates reconnect with missed event replay

**Definition of Done:** Client disconnects + reconnects → receives all missed events in order.

---

### `0.5.3` — SSE Backpressure

**Deliverables:**
- Per-client buffered channel (depth = `BrokerConfig.ClientBufSize`, default 64)
- Non-blocking send: `select { case ch <- e: default: }` drops for slow clients
- `Broker.DroppedEvents() int64` — counter for monitoring
- `BrokerConfig.OnDrop func(clientID string, event *Event)` — drop hook
- Slow client disconnect: after N consecutive drops, force-close connection
- `BrokerConfig.MaxDropsBeforeClose int` (default 100)

**Definition of Done:** One extremely slow client does not delay all other clients.

---

### `0.5.4` — SSE Named Events + Retry Field

**Deliverables:**
- `Event.Event string` — named event type (`event: update\n`)
- `Event.Retry int` — client reconnect interval in milliseconds
- Multi-line data: `\n` in `Event.Data` produces multiple `data:` lines
- `Event.Comment string` — comment line for server diagnostics
- `Context.SSE(broker *sse.Broker)` — convenience method to wire SSE handler
- Client-side example (JS): filtering by event type with `addEventListener`

**Definition of Done:** `addEventListener("update", fn)` receives only `event: update` events.

---

### `0.5.5` — SSE + HTTP/2 Multiplexing

**Deliverables:**
- Verify SSE works correctly over HTTP/2 (no chunked encoding needed — h2 frames)
- HTTP/2 SSE: remove `Transfer-Encoding: chunked` header (invalid in h2)
- Flusher still used for `http.Flusher` compatibility; h2 uses DATA frames
- Benchmark: SSE throughput over HTTP/2 vs HTTP/1.1
- Multiple SSE streams multiplexed over single HTTP/2 connection
- Example: `examples/sse/http2-multi/main.go`

**Definition of Done:** 10 SSE streams multiplexed over single HTTP/2 connection.

---

## Phase 6 — Authentication Middleware (0.6.0 → 0.6.4)

> Goal: Production-ready auth middleware suite.

---

### `0.6.0` — JWT Middleware

**Deliverables:**
- `middleware.JWT(config JWTConfig) HandlerFunc` per ARCHITECTURE §16.5
- `JWTConfig.Secret []byte`, `SignMethod`, `TokenLookup`, `AuthScheme`
- `JWTConfig.Claims` — custom claims struct (implements `jwt.Claims`)
- `JWTConfig.SkipRoutes []string`
- `JWTConfig.ContextKey string` (default `"user"`)
- Token parsed + stored on context: `c.Get("user")`
- Explicit algorithm whitelist — reject `alg: none` attacks
- Returns 401 on missing/invalid/expired token
- Uses `github.com/golang-jwt/jwt/v5`

**Definition of Done:** Expired token → 401. Valid token → handler called with claims on context.

---

### `0.6.1` — Basic Auth Middleware

**Deliverables:**
- `middleware.BasicAuth(users map[string]string) HandlerFunc`
- `middleware.BasicAuthWithValidator(fn func(user, pass string, c *Context) bool) HandlerFunc`
- Constant-time password comparison (`subtle.ConstantTimeCompare`)
- `WWW-Authenticate: Basic realm="..."` on 401
- `JWTConfig.Realm string`

**Definition of Done:** Wrong password → 401 with `WWW-Authenticate`. Timing safe against brute force.

---

### `0.6.2` — API Key Middleware

**Deliverables:**
- `middleware.APIKey(config APIKeyConfig) HandlerFunc`
- Key lookup locations: header (default `X-API-Key`), query param, cookie
- `APIKeyConfig.Validator func(key string, c *Context) bool`
- `APIKeyConfig.ContextKey string` — store validated key on context
- Returns 401 on missing key, 403 on invalid key
- Example: static key map + database lookup validator

**Definition of Done:** Invalid API key → 401. Valid key stored on context for downstream handlers.

---

### `0.6.3` — Auth Middleware SkipRoutes + Extractor

**Deliverables:**
- Refactor `SkipRoutes` into shared `Skipper func(*Context) bool` interface
- All auth middleware supports `Skipper`
- Common skippers: `SkipPublicRoutes`, `SkipHealthCheck`, `SkipOptionsMethod`
- `middleware.NewSkipper(paths ...string) Skipper` — convenience builder
- `JWTConfig.ErrorHandler func(c *Context, err error) error` — custom 401 response

**Definition of Done:** `/health` skipped by JWT middleware. All other routes protected.

---

### `0.6.4` — Auth Custom Claims + Token Refresh Helper

**Deliverables:**
- Generic claims: `middleware.JWT[T any](config JWTConfig[T])` (Go 1.21 generics)
- `Context.GetClaims() (*jwt.Claims, bool)` — typed claim retrieval
- `auth.GenerateToken(claims, secret, expiry) (string, error)` — token generation helper
- `auth.RefreshToken(tokenStr, secret, newExpiry) (string, error)` — refresh helper
- Token blacklist interface: `JWTConfig.Blacklist TokenBlacklist` — check revoked tokens
- Example: `examples/auth/jwt-refresh/main.go`

**Definition of Done:** Custom claims struct accessible from handler. Refresh token extends expiry.

---

## Phase 7 — Configuration + Observability (0.7.0 → 0.7.5)

> Goal: Production observability — config, tracing, metrics, health.

---

### `0.7.0` — Configuration System

**Deliverables:**
- `config.Config` top-level struct per ARCHITECTURE §15
- `config.Load(path string) (*Config, error)` — YAML + env overlay
- `overlayEnv` — `RUDRA_SERVER_PORT` → `server.port` mapping
- `${ENV_VAR}` interpolation in YAML values
- Typed sub-configs: `ServerConfig`, `JWTConfig`, `CORSConfig`, `LogConfig`
- `config.Watch(path, fn)` — reload config on file change (fsnotify)
- Validation of required config fields at startup

**Definition of Done:** `RUDRA_SERVER_PORT=9090 ./app` overrides YAML port. Invalid config panics at startup with clear message.

---

### `0.7.1` — Structured Logger (slog)

**Deliverables:**
- Replace `log.Printf` with `log/slog` throughout codebase
- `Engine.Logger() *slog.Logger` — access engine logger
- `WithLogger(logger *slog.Logger)` functional option
- Logger propagated to context: `c.Logger()` returns request-scoped logger with `request_id`, `path`
- `middleware.Logger` uses slog JSON handler by default
- `LoggerConfig.Formatter` — custom log formatter function

**Definition of Done:** All log output is structured JSON by default. `c.Logger().Info("msg", "key", val)` includes request_id.

---

### `0.7.2` — OpenTelemetry Tracing Middleware

**Deliverables:**
- `middleware.Trace(config TraceConfig) HandlerFunc`
- Creates span per request: `span.SetAttributes(method, path, status, user_agent)`
- Extracts parent context from `traceparent` / `tracestate` headers (W3C TraceContext)
- `Context.Span() trace.Span` — access current span from handler
- `Context.TraceID() string` — convenience
- `TraceConfig.Tracer trace.Tracer` — injectable tracer
- `TraceConfig.Propagator propagation.TextMapPropagator`
- Uses `go.opentelemetry.io/otel`

**Definition of Done:** Spans visible in Jaeger when using OTLP exporter.

---

### `0.7.3` — Prometheus Metrics Middleware

**Deliverables:**
- `middleware.Metrics(config MetricsConfig) HandlerFunc`
- Metrics exposed:
  - `rudra_http_requests_total{method, path, status}` — counter
  - `rudra_http_request_duration_seconds{method, path}` — histogram
  - `rudra_http_request_size_bytes` — histogram
  - `rudra_http_response_size_bytes` — histogram
  - `rudra_http_active_requests` — gauge
- `Engine.MetricsHandler()` — returns `promhttp.Handler()` for `/metrics` route
- `MetricsConfig.Namespace string` (default `"rudra"`)
- `MetricsConfig.Buckets []float64` — histogram buckets

**Definition of Done:** `/metrics` returns valid Prometheus exposition format. Grafana dashboard works.

---

### `0.7.4` — Health Check Endpoint

**Deliverables:**
- `Engine.UseHealthCheck(path string, checks ...HealthChecker)` — registers `/health`
- `HealthChecker` interface: `Check(ctx context.Context) error`
- Built-in checkers: `PingChecker`, `MemoryChecker`, `GoroutineChecker`
- Response schema: `{"status": "ok", "checks": {"db": "ok", "redis": "degraded"}}`
- `200 OK` if all pass, `503 Service Unavailable` if any fail
- `/health/live` (liveness) + `/health/ready` (readiness) Kubernetes endpoints
- Timeout per check: `HealthConfig.Timeout time.Duration`

**Definition of Done:** Failing DB check → 503 from `/health/ready`. Liveness still 200.

---

### `0.7.5` — pprof + Debug Endpoints

**Deliverables:**
- `Engine.UsePprof(prefix string)` — registers `net/http/pprof` handlers
- Routes: `/debug/pprof/`, `/debug/pprof/cmdline`, `/debug/pprof/profile`, `/debug/pprof/symbol`, `/debug/pprof/trace`
- Protected by `middleware.BasicAuth` by default
- `Engine.UseExpvar(path string)` — expvar endpoint
- `DebugConfig.EnableInProduction bool` — must opt-in for prod
- Build tag `//go:build !nodebug` to strip from production builds

**Definition of Done:** `go tool pprof http://localhost:8080/debug/pprof/heap` works.

---

## Phase 8 — Testing + DX + CLI (0.8.0 → 0.8.9)

> Goal: Best-in-class developer experience — testing utils, CLI, docs.

---

### `0.8.0` — TestUtil Request Builder

**Deliverables:**
- `testutil.NewTestRequest(app) *TestRequest` — fluent builder
- `TestRequest.GET`, `.POST`, `.PUT`, `.PATCH`, `.DELETE` methods
- `TestRequest.Header(key, value)`, `.Cookie(name, value)`, `.Bearer(token)`
- `TestRequest.Do() *TestResponse` — executes against `httptest.Recorder`
- `TestResponse.Status(t, code)`, `.Body() []byte`, `.JSON(t, v)`
- `TestResponse.HasHeader(t, key, val)`, `.HasCookie(t, name)`
- `TestResponse.BodyContains(t, substr)`
- `testutil.NewTestServer(app) *TestServer` — wraps `httptest.NewServer`

**Definition of Done:** Full CRUD API tested using testutil with zero boilerplate.

---

### `0.8.1` — TestUtil Assert Helpers + Mock Middleware

**Deliverables:**
- `testutil.AssertJSON(t, body, expected)` — deep JSON equality
- `testutil.AssertStatus(t, resp, code)`
- `testutil.MockJWT(claims any) HandlerFunc` — injects claims without real JWT
- `testutil.MockUser(user any) HandlerFunc` — injects user on context
- `testutil.RecordMiddleware() (*RecordedRequests, HandlerFunc)` — captures all requests
- Table-driven test helper: `testutil.RunCases(t, app, []TestCase{...})`

**Definition of Done:** Mock JWT middleware allows testing protected routes without real tokens.

---

### `0.8.2` — Streaming + Chunked Transfer Renderer

**Deliverables:**
- `render.Stream(w, code, contentType, fn)` per ARCHITECTURE §13.2
- `Context.Stream(code int, contentType string, fn func(io.Writer) error) error`
- `Context.NDJSON(code int, items <-chan any) error` — streaming newline-delimited JSON
- `flushWriter` wrapper: writes + flushes on every `Write` call
- File streaming: `Context.File(path string) error` — serves file with Content-Disposition
- `Context.Attachment(path, filename string) error` — force download
- `Context.Inline(path, filename string) error` — display in browser

**Definition of Done:** 1GB file streamed without loading into memory. NDJSON stream delivers items incrementally.

---

### `0.8.3` — Render Template Engine

**Deliverables:**
- `Engine.LoadHTMLGlob(pattern string)` — loads `html/template` files
- `Engine.LoadHTMLFiles(files ...string)`
- `Context.Render(code int, name string, data any) error`
- Template functions: `url`, `csrf_token`, `json`, `date`, `truncate`
- Layout/partial support: `{{ template "layout" . }}`
- Hot reload in development: `WithTemplateReload(true)` option
- `render.HTML(w, code, tmpl, data)` renderer

**Definition of Done:** `c.Render(200, "index.html", data)` renders template with data correctly.

---

### `0.8.4` — Full Benchmark Suite

**Deliverables:**
- `benchmarks/framework_bench_test.go` — all frameworks on equal footing
- Benchmark endpoints: plain text, JSON (small), JSON (large), params, middleware chain
- `benchmarks/scripts/wrk.sh` — wrk benchmark for all frameworks
- `benchmarks/scripts/ab.sh` — Apache Bench comparison
- `benchmarks/scripts/h2load.sh` — HTTP/2 benchmark
- `benchmarks/results/` — tracked results directory
- `make bench` target in Makefile
- CI: run benchmarks on PR, comment results

**Definition of Done:** Benchmark results show Rudra competitive with or beating Gin/Echo on all endpoints.

---

### `0.8.5` — CLI Scaffold Tool

**Deliverables:**
- `cmd/rudra/` — CLI binary
- `rudra new <project-name>` — generates new Rudra project
- `rudra new --template api` — REST API template
- `rudra new --template ws` — WebSocket app template
- `rudra generate route <name>` — generates handler + route registration
- `rudra generate middleware <name>` — generates middleware scaffold
- `go install github.com/AarambhDevHub/rudra/cmd/rudra@latest`

**Definition of Done:** `rudra new myapp && cd myapp && go run .` starts a working Rudra server.

---

### `0.8.6` — Redirect Middleware + HTTPS Enforcer

**Deliverables:**
- `middleware.HTTPSRedirect() HandlerFunc` — redirects all HTTP → HTTPS
- `middleware.WWWRedirect() HandlerFunc` — redirect `www.` → apex
- `middleware.NonWWWRedirect() HandlerFunc` — redirect apex → `www.`
- `middleware.TrailingSlash(config TrailingSlashConfig)` — add/remove trailing slash
- `Engine.UseHTTPS()` — convenience to add HTTPS redirect + HSTS

**Definition of Done:** HTTP request to port 80 redirected to HTTPS with 301.

---

### `0.8.7` — Context Extras

**Deliverables:**
- `Context.IsWebSocket() bool` — detects WS upgrade request
- `Context.IsEventStream() bool` — detects SSE client
- `Context.IsHTTP2() bool` — checks `r.Proto`
- `Context.IsAJAX() bool` — `X-Requested-With: XMLHttpRequest`
- `Context.Accept() string` — best match from `Accept` header
- `Context.Deadline() (time.Time, bool)` — from request context
- `Context.Done() <-chan struct{}` — from request context
- `Context.GoContext() context.Context` — raw `context.Context`
- `Context.WithValue(key, val any) *Context` — derives child context

**Definition of Done:** All context accessors covered by unit tests.

---

### `0.8.8` — Middleware Extras

**Deliverables:**
- `middleware.Cache(config CacheConfig) HandlerFunc` — in-memory LRU response cache
  - `CacheConfig.TTL time.Duration`
  - `CacheConfig.MaxSize int` (LRU eviction)
  - `CacheConfig.KeyFunc func(*Context) string`
  - Only caches GET responses with 200 status
- `middleware.Rewrite(rules map[string]string) HandlerFunc` — URL rewriting
- `middleware.Proxy(target string) HandlerFunc` — reverse proxy to upstream

**Definition of Done:** Cached response served from memory on second request. Cache miss triggers handler.

---

### `0.8.9` — API Documentation (Swagger/OpenAPI)

**Deliverables:**
- `rudra.Swagger(config SwaggerConfig)` — serves Swagger UI at `/docs`
- `@rudra:route` code comment annotations
- Auto-generated `openapi.json` from route + struct annotations
- `SwaggerConfig.Title`, `Version`, `Description`, `BasePath`
- Swagger UI served from embedded `embed.FS`
- ReDoc alternative renderer option

**Definition of Done:** `/docs` serves interactive Swagger UI with all registered routes.

---

## Phase 9 — Hardening + Performance + Launch Prep (0.9.0 → 0.9.9)

> Goal: Zero-alloc audit, security hardening, benchmark proof, API stability.

---

### `0.9.0` — Zero-Allocation Audit

**Deliverables:**
- Systematic `go test -bench=. -benchmem` on every hot path
- Allocation budget table achieved per ARCHITECTURE §18.8
- Fix any remaining allocations in: routing, context reset, middleware chain, JSON render
- `unsafe` string↔bytes conversions on validated hot paths
- Static route map O(1) fast path verified
- Header canonicalization cache implemented

**Definition of Done:** Allocation budget table from ARCHITECTURE §18.8 fully achieved.

---

### `0.9.1` — Security Hardening

**Deliverables:**
- TLS config hardened per ARCHITECTURE §20.2 (ciphers, curves, min version)
- TLS session resumption via session tickets
- Full security threat matrix reviewed per ARCHITECTURE §20.1
- URL path sanitization: clean `../` traversal attempts
- Strict `Content-Length` enforcement (request smuggling prevention)
- `middleware.Secure` extended with `Permissions-Policy` header
- Fuzz test: router path inputs, request body inputs
- `govulncheck` added to CI

**Definition of Done:** `govulncheck` clean. OWASP ZAP scan passes on demo app.

---

### `0.9.2` — Performance Tuning Sprint

**Deliverables:**
- CPU profiling of wrk benchmark run → identify top 3 hotspots → fix
- Memory profiling: reduce live heap on idle server
- `sync.Pool` tuning: pre-warm pools at startup
- Router compile: sort static routes map keys at startup for cache locality
- Response writer: avoid extra `WriteHeader` calls
- Benchmark regression tests added to CI (fail if >5% regression)

**Definition of Done:** wrk benchmark improved by ≥10% over `0.8.x` baseline.

---

### `0.9.3` — Integration Test Suite

**Deliverables:**
- Full integration tests for: HTTP/1.1, HTTP/2, WebSocket, SSE, all middleware
- Docker Compose: nginx reverse proxy → Rudra → integration tests
- Race condition tests: `go test -race ./...` clean
- Fuzz tests: `go test -fuzz` for router, binding, validation
- Load test: `vegeta` 10k req/sec sustained for 60s, 0 errors
- Memory leak test: 1 hour run, heap stable

**Definition of Done:** `go test -race ./...` clean. 0 memory leaks under sustained load.

---

### `0.9.4` — API Stability Review + Deprecation

**Deliverables:**
- Full public API review: rename inconsistencies, remove redundant methods
- Deprecate any methods to be removed before `1.0.0`
- `//go:deprecated` comments on deprecated symbols
- Migration guide for any breaking changes from `0.8.x`
- `CHANGELOG.md` updated with all `0.x.x` changes
- Semantic versioning audit: no accidental breaking changes in patch releases

**Definition of Done:** Public API documented and stable. No surprise breaking changes.

---

### `0.9.5` — Final Benchmark vs Gin/Echo/Fiber

**Deliverables:**
- Clean-room benchmark: identical handler, identical hardware, fresh machine
- Benchmark categories: plain text, JSON 1KB, JSON 100KB, 5 params, 10 middleware layers
- HTTP/1.1 + HTTP/2 separate results
- Results published in `benchmarks/results/v0.9.5/`
- `README.md` benchmark table updated
- Blog post drafted: "How Rudra beats Gin by X%" (for Aarambh Dev Hub)

**Definition of Done:** Rudra beats or matches Gin and Echo. Results reproducible.

---

### `0.9.6` — Documentation Polish

**Deliverables:**
- Full `pkg.go.dev` documentation: every exported symbol has godoc
- `docs/` directory: getting-started, routing, middleware, websocket, sse, http2 guides
- All `examples/` compile and run correctly
- `README.md`: badges, quick start, feature list, benchmark table, comparison matrix
- `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md` finalized

**Definition of Done:** `pkg.go.dev/github.com/AarambhDevHub/rudra` renders complete documentation.

---

### `0.9.7` — gRPC-Web Bridge (Bonus)

**Deliverables:**
- `grpcweb` package: translate gRPC-Web framing → standard gRPC
- `Engine.UseGRPCWeb(grpcServer *grpc.Server)` — bridge registration
- Supports `Content-Type: application/grpc-web+proto`
- Works over HTTP/2 and HTTP/1.1
- CORS for gRPC-Web clients
- Example: `examples/grpc-web/main.go`

**Definition of Done:** Browser gRPC-Web client communicates with Go gRPC server via Rudra bridge.

---

### `0.9.8` — Release Candidate 1

**Deliverables:**
- Tag: `v0.9.8-rc.1`
- Full build + test + benchmark pipeline clean
- Docker image: `ghcr.io/aarambhdevhub/rudra-example:0.9.8-rc.1`
- GitHub Release with full changelog
- `go get github.com/AarambhDevHub/rudra@v0.9.8-rc.1` works
- Community feedback period: 2 weeks

**Definition of Done:** RC1 publicly tagged. Issue tracker open for feedback.

---

### `0.9.9` — Release Candidate 2 (Final Pre-1.0)

**Deliverables:**
- RC1 feedback incorporated
- All reported bugs fixed
- Final `go vet`, `staticcheck`, `golangci-lint` clean
- `LICENSE-MIT` + `LICENSE-APACHE` headers on all source files
- `CHANGELOG.md` complete for all versions `0.0.1` → `0.9.9`
- YouTube video: "Building Rudra — A Go Web Framework from Scratch" (Aarambh Dev Hub)
- Tag: `v0.9.9`

**Definition of Done:** `v0.9.9` tagged. Framework ready for production use. `1.0.0` reserved for stable API guarantee.

---

## Milestone Summary

| Milestone | Version Range | Theme                          |
|-----------|---------------|--------------------------------|
| Foundation| 0.0.1–0.0.9   | Engine, router, context, render| ✅
| HTTP/1.1  | 0.1.0–0.1.9   | Full HTTP/1.1 + core middleware| 🔄
| Binding   | 0.2.0–0.2.9   | Request binding + validation   |
| HTTP/2    | 0.3.0–0.3.5   | TLS, h2c, server push          |
| WebSocket | 0.4.0–0.4.6   | WS, hub, rooms, compression    |
| SSE       | 0.5.0–0.5.5   | SSE, heartbeat, reconnect      |
| Auth      | 0.6.0–0.6.4   | JWT, Basic, API Key            |
| Observability | 0.7.0–0.7.5 | Config, logging, tracing, metrics |
| DX + CLI  | 0.8.0–0.8.9   | Testing, CLI, templates, docs  |
| Launch Prep| 0.9.0–0.9.9  | Hardening, perf, RC1, RC2      |

---

*Rudra (रुद्र) — Fierce. Fast. Fearless.*
*© Aarambh Dev Hub — MIT + Apache 2.0*
