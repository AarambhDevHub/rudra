# Contributing to Rudra (रुद्र)

Thank you for your interest in contributing to Rudra. Every contribution — bug fix, feature, test, documentation, typo — is appreciated and makes Rudra better for everyone.

---

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Ways to Contribute](#ways-to-contribute)
3. [Development Setup](#development-setup)
4. [Project Structure](#project-structure)
5. [Coding Standards](#coding-standards)
6. [Writing Tests](#writing-tests)
7. [Benchmarking](#benchmarking)
8. [Submitting a Pull Request](#submitting-a-pull-request)
9. [Commit Message Format](#commit-message-format)
10. [Issue Guidelines](#issue-guidelines)
11. [Performance Contribution Rules](#performance-contribution-rules)
12. [Dependency Policy](#dependency-policy)

---

## Code of Conduct

This project follows our [Code of Conduct](./CODE_OF_CONDUCT.md). By participating, you agree to uphold it. Please report unacceptable behavior to the maintainers.

---

## Ways to Contribute

**You don't have to write code to contribute.**

- **Report a bug** — open a GitHub issue with a minimal reproduction
- **Request a feature** — open a discussion before opening a PR for large features
- **Fix a bug** — check the `good first issue` and `help wanted` labels
- **Write tests** — we always need more coverage, especially edge cases
- **Improve docs** — godoc comments, README, examples, guides
- **Write examples** — practical examples in `examples/` are incredibly valuable
- **Improve benchmarks** — new benchmark scenarios, fairer comparisons
- **Review PRs** — code review from the community is always welcome
- **Answer questions** — GitHub Discussions and Discord

---

## Development Setup

### Requirements

- Go 1.22 or later
- Git
- `make` (optional but recommended)

### Clone and Build

```bash
git clone https://github.com/AarambhDevHub/rudra.git
cd rudra
go mod tidy
go build ./...
```

### Run Tests

```bash
# All tests
go test ./...

# With race detector (required before all PRs)
go test -race ./...

# Specific package
go test ./router/...
go test ./middleware/...
```

### Run Benchmarks

```bash
# All benchmarks with allocation reporting
go test -bench=. -benchmem ./...

# Specific benchmark
go test -bench=BenchmarkRouterStatic -benchmem ./router/...

# Run 5 times for stable results
go test -bench=. -benchmem -count=5 ./...
```

### Lint

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run ./...
```

### Check for Vulnerabilities

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

---

## Project Structure

```
rudra/
├── core/           Engine, server lifecycle, options, signals
├── router/         Radix tree router, groups, params, named routes
├── context/        Request/response context, sync.Pool
├── middleware/      All built-in middleware (logger, cors, jwt, etc.)
│   └── auth/       Auth-specific middleware
├── binding/        Request binding (json, xml, form, query, path, header)
├── render/         Response renderers (json, html, text, stream, blob)
├── ws/             WebSocket engine (upgrader, conn, hub, pool)
├── sse/            SSE broker and event types
├── validator/      Struct validation with tag-based rules
├── config/         Configuration loader (yaml + env)
├── errors/         HTTP error types and global handler
├── testutil/       Test utilities and assertion helpers
├── examples/       Runnable examples for all features
└── benchmarks/     Benchmark suite and comparison scripts
```

Each package has a clear, single responsibility. If you are unsure where a new file belongs, open a discussion.

---

## Coding Standards

### Style

Rudra follows standard Go style. Run `gofmt` and `goimports` before committing.

```bash
gofmt -w .
goimports -w .
```

### Comments

Every **exported** symbol must have a godoc comment. The comment must explain **intent** — not restate what the code already says.

```go
// BAD — restates the code
// SetParam sets the param key to value.
func (c *Context) SetParam(key, value string) { ... }

// GOOD — explains intent and constraint
// SetParam stores a URL path parameter captured during routing.
// Called exclusively by the router during route matching.
// Uses a pre-allocated fixed-size array — zero heap allocation for ≤16 params.
func (c *Context) SetParam(key, value string) { ... }
```

### Error Handling

Always return errors — never silently swallow them. Use `errors.RudraError` for HTTP errors. Use `fmt.Errorf("rudra/package: %w", err)` for internal errors.

```go
// BAD
conn.Close()

// GOOD
if err := conn.Close(); err != nil {
    return fmt.Errorf("rudra/ws: failed to close connection: %w", err)
}
```

### No Reflection on Hot Paths

Reflection is acceptable at startup (route registration, validator struct metadata caching). It is **never acceptable** in a per-request code path. If you need to parse a struct on every request, use a cache keyed by `reflect.Type`.

### Interfaces for Testability

Any component that accesses external state (file system, network, time, random) must be hidden behind an interface so tests can inject fakes.

```go
// Good — injectable clock for testing timeouts
type Clock interface {
    Now() time.Time
    After(d time.Duration) <-chan time.Time
}
```

### No Global State

Avoid package-level variables that mutate at runtime. All state must live on structs. The only acceptable package-level variables are pre-compiled `regexp` patterns and `sync.Pool` instances.

---

## Writing Tests

Every PR must include tests. There are three categories:

### Unit Tests

File: `{file}_test.go` in the same package.

```go
func TestContextParam(t *testing.T) {
    c := context.New()
    c.SetParam("id", "42")
    assert.Equal(t, "42", c.Param("id"))
}
```

### Integration Tests

File: `{package}_integration_test.go`. Use `testutil.NewTestRequest`:

```go
func TestUserRoute(t *testing.T) {
    app := setupTestApp()

    var resp UserResponse
    testutil.NewTestRequest(app).
        GET("/api/users/42").
        Header("Authorization", "Bearer "+testToken).
        Do().
        Status(t, 200).
        JSON(t, &resp)

    assert.Equal(t, "42", resp.ID)
}
```

### Table-Driven Tests

Prefer table-driven tests for anything with multiple cases:

```go
func TestRouter(t *testing.T) {
    cases := []struct {
        name     string
        method   string
        path     string
        wantCode int
    }{
        {"static match",     "GET",  "/users",     200},
        {"param match",      "GET",  "/users/42",  200},
        {"not found",        "GET",  "/missing",   404},
        {"method not allowed","POST", "/users/42",  405},
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            // ...
        })
    }
}
```

### Race Condition Tests

All concurrent components (Hub, Broker, Router) must be tested with `-race`:

```bash
go test -race ./ws/...
go test -race ./sse/...
```

---

## Benchmarking

Rudra takes performance seriously. If your PR touches a hot path — router, context, middleware chain, JSON rendering — you must include benchmark results.

### Rules

1. Run benchmarks on your local machine both before and after your change.
2. Include both run results in your PR description.
3. A PR that introduces a regression of **>2%** on any existing benchmark will not be merged without a strong justification.
4. New features on the hot path must include a benchmark showing allocations.

### Benchmark Template

```go
func BenchmarkMyFeature(b *testing.B) {
    // setup outside the loop
    ...

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        // what you're benchmarking
    }
}
```

### Required Output Format in PR

```
Before:
BenchmarkRouterStatic-4   50000000   24.1 ns/op   0 B/op   0 allocs/op

After:
BenchmarkRouterStatic-4   52000000   23.3 ns/op   0 B/op   0 allocs/op

Change: +3.3% throughput, 0 alloc unchanged ✅
```

---

## Submitting a Pull Request

1. **Fork** the repository and create a branch from `main`.

   ```bash
   git checkout -b fix/router-param-trailing-slash
   ```

2. **Make your changes.** Follow the coding standards above.

3. **Add tests.** All new code must be tested.

4. **Run the full test suite:**

   ```bash
   go test -race ./...
   golangci-lint run ./...
   govulncheck ./...
   ```

5. **Update documentation** if you changed public API or behavior.

6. **Open the PR** against the `main` branch with:
   - A clear title: `fix(router): handle trailing slash with params`
   - A description of what changed and why
   - Benchmark results if touching hot paths
   - Reference to any related issue: `Closes #42`

7. **Wait for review.** A maintainer will review within a few days. Be ready to iterate.

### PR Checklist

- [ ] `go test -race ./...` passes
- [ ] `golangci-lint run ./...` passes
- [ ] All new exported symbols have godoc comments
- [ ] Benchmark results included if touching hot paths
- [ ] `CHANGELOG.md` updated under `[Unreleased]`
- [ ] Example added or updated if new feature

---

## Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short description>

[optional body]

[optional footer: Closes #issue]
```

**Types:** `feat`, `fix`, `perf`, `refactor`, `test`, `docs`, `chore`, `ci`

**Scopes:** `router`, `context`, `middleware`, `ws`, `sse`, `binding`, `render`, `validator`, `config`, `errors`, `core`, `testutil`, `deps`, `ci`

**Examples:**

```
feat(ws): add permessage-deflate compression support
fix(router): correctly handle wildcard after static segment
perf(context): replace map with fixed array for path params
docs(sse): add Last-Event-ID reconnect example
test(middleware): add race condition tests for rate limiter
chore(deps): upgrade sonic to v1.12.0
```

---

## Issue Guidelines

### Bug Reports

Include all of the following:

- Rudra version: `v0.x.y`
- Go version: `go version`
- OS and architecture
- Minimal reproduction code (as small as possible)
- Expected behavior
- Actual behavior
- Error output / stack trace if available

### Feature Requests

Before opening a feature request:

1. Check if it is already on the [ROADMAP](./docs/ROADMAP.md)
2. Check existing issues and discussions
3. Open a **Discussion** first for large features — get alignment before writing code

Feature requests should include:

- The problem you are solving (not just the solution)
- How you would use this feature
- Whether you are willing to implement it

### Performance Issues

Include benchmark results showing the regression, Go version, hardware, and a minimal reproduction.

---

## Performance Contribution Rules

These rules apply to any PR that touches a hot path (routing, context, middleware, JSON encoding):

1. **No new allocations on the hot path.** If your feature requires an allocation, it must be in a `sync.Pool`.
2. **No reflection per request.** Cache struct metadata at startup.
3. **No unnecessary copying.** Use `unsafe` string↔bytes conversion where safe and documented.
4. **Measure first.** Always profile before optimizing. Include `pprof` output for significant changes.
5. **Document the why.** Non-obvious performance code must have a comment explaining the technique and the tradeoff.

---

## Dependency Policy

Rudra has a strict dependency policy. We minimize external dependencies to keep the framework lean and auditable.

**Allowed external dependencies (already in go.mod):**

- `golang.org/x/net` — HTTP/2, h2c
- `golang.org/x/crypto` — bcrypt, argon2
- `github.com/bytedance/sonic` — SIMD JSON (optional, build tag gated)
- `github.com/golang-jwt/jwt/v5` — JWT parsing
- `github.com/klauspost/compress` — zstd + brotli
- `github.com/andybalholm/brotli` — brotli
- `go.opentelemetry.io/otel` — tracing
- `github.com/prometheus/client_golang` — metrics
- `gopkg.in/yaml.v3` — config

**Adding a new dependency requires:**

1. A strong justification (cannot be implemented reasonably in stdlib)
2. The dependency must be actively maintained
3. No transitive dependencies that conflict with existing ones
4. Maintainer approval before the PR is opened

**When in doubt: implement it yourself.** Rudra's custom WebSocket upgrader, radix tree router, and validator exist because the stdlib (or a small implementation) is better than pulling in a heavy dependency.

---

## Thank You

Rudra exists because of the open-source community. Every star, issue, PR, and comment matters. Building in public is hard — your support makes it worthwhile.

**Aarambh Dev Hub — Start building, keep building.**
