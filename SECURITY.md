# Security Policy

## Supported Versions

Rudra is currently in active early development. Security fixes are applied to the **latest released version only**.

| Version       | Supported          |
|---------------|--------------------|
| Latest `0.x.y`| ✅ Active support  |
| Older `0.x.y` | ❌ Please upgrade  |

Once Rudra reaches `v1.0.0`, we will maintain a defined support window for older minor versions.

---

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

If you discover a security vulnerability in Rudra, please report it responsibly through one of the following channels:

### Option 1 — GitHub Private Security Advisory (Preferred)

Use GitHub's built-in private vulnerability reporting:

1. Go to [https://github.com/AarambhDevHub/rudra/security/advisories](https://github.com/AarambhDevHub/rudra/security/advisories)
2. Click **"New draft security advisory"**
3. Fill in the details as described below

### Option 2 — Email

Send a detailed report to the maintainer. Find the contact in the GitHub profile at [github.com/aarambh-darshan](https://github.com/aarambh-darshan).

Use the subject line: `[RUDRA SECURITY] <brief description>`

---

## What to Include in Your Report

A good vulnerability report helps us respond faster. Please include:

- **Description** — a clear explanation of the vulnerability and what it allows an attacker to do
- **Affected component** — which package or feature (e.g., `middleware/jwt`, `ws/upgrader`, `router`)
- **Affected versions** — which version(s) you confirmed the issue on
- **Reproduction steps** — the minimal code or request sequence to trigger the vulnerability
- **Impact assessment** — what is the worst-case impact if this is exploited in production
- **Suggested fix** — if you have a recommendation, we welcome it (not required)
- **CVE** — if you have already requested a CVE, include the number

---

## Response Timeline

We take security seriously and will respond as quickly as possible.

| Stage                        | Target Timeline        |
|------------------------------|------------------------|
| Acknowledgement of report    | Within **48 hours**    |
| Initial assessment           | Within **5 days**      |
| Fix development start        | Within **7 days**      |
| Fix released (patch version) | Within **30 days**     |
| Public disclosure            | After fix is released  |

For critical vulnerabilities with active exploitation in the wild, we will aim to release a fix within **72 hours**.

---

## Disclosure Policy

We follow a **coordinated disclosure** model:

1. Reporter submits vulnerability privately.
2. Maintainers acknowledge and assess the report.
3. Fix is developed and tested privately.
4. Patch version is released.
5. Security advisory is published on GitHub with full details.
6. Reporter is credited (unless they prefer to remain anonymous).

We ask that reporters allow us at least **14 days** after the fix is released before publishing their own write-up, to give users time to upgrade.

---

## Security Scope

### In Scope

The following are considered in scope for security reports:

- Authentication bypass in `middleware/auth/jwt.go`, `basic.go`, `apikey.go`
- CSRF protection bypass in `middleware/csrf.go`
- Path traversal via router or static file server
- Request smuggling via HTTP/1.1 handling
- WebSocket handshake bypass or header injection
- Information leakage (stack traces, internal errors exposed to clients)
- Panic-to-crash via crafted inputs (DoS via panic)
- Memory exhaustion via unbounded input (DoS)
- JWT algorithm confusion attacks (e.g., `alg: none` accepted)
- TLS configuration weaknesses
- Race conditions leading to data corruption or auth bypass
- Dependencies with known CVEs (`govulncheck` findings)

### Out of Scope

The following are **not** considered security vulnerabilities in Rudra:

- Vulnerabilities in user application code built on top of Rudra
- Issues that require physical access to the server
- Social engineering attacks
- Vulnerabilities in dependencies that do not affect Rudra's exposed surface
- Theoretical attacks with no practical exploitation path
- Rate limiting bypass if the user has not configured the rate limiter
- Security misconfigurations in user-written middleware

---

## Security Best Practices for Rudra Users

These are not vulnerabilities in Rudra, but mistakes we commonly see in applications built with it.

### Always use Recovery middleware

```go
app.Use(middleware.Recovery())
```

Without Recovery, a panic in any handler crashes the entire server.

### Set TLS minimum version

```go
// RunTLS automatically configures TLS 1.2+ with hardened cipher suites.
// Do NOT override TLSConfig to use weaker settings.
app.RunTLS(":443", "cert.pem", "key.pem")
```

### Restrict JWT algorithm

```go
// Always specify the signing method explicitly.
// Never allow `alg: none` or multiple algorithms.
middleware.JWT(middleware.JWTConfig{
    Secret:     []byte(os.Getenv("JWT_SECRET")),
    SignMethod: jwt.SigningMethodHS256, // explicit — never omit this
})
```

### Set body size limits

```go
// Without this, a client can send an unlimited body and exhaust memory.
app.Use(middleware.BodyLimit(32 << 20)) // 32MB limit
```

### Do not expose pprof in production without auth

```go
// Wrong — exposes heap, goroutines, CPU profile to anyone
app.UsePprof("/debug/pprof")

// Right — protect with Basic Auth and IP whitelist
app.UsePprof("/debug/pprof",
    middleware.BasicAuth(map[string]string{"admin": os.Getenv("PPROF_PASS")}),
    middleware.AllowIPs("10.0.0.0/8"),
)
```

### Validate all user input

```go
type CreateUserRequest struct {
    Name  string `json:"name"  rudra:"required,min=2,max=64,alphanum"`
    Email string `json:"email" rudra:"required,email"`
}

func createUser(c *context.Context) error {
    var req CreateUserRequest
    return c.MustBind(&req) // binds + validates, returns 422 on failure
}
```

### Use CSRF middleware for browser-facing APIs

```go
app.Use(middleware.CSRF(middleware.CSRFConfig{
    Secure:   true,
    SameSite: http.SameSiteStrictMode,
}))
```

### Check WebSocket origin

```go
upgrader := ws.NewUpgrader()
upgrader.CheckOrigin = func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    return origin == "https://yourdomain.com"
}
```

---

## Security-Related Configuration Reference

| Option / Middleware        | Security Purpose                              | Default       |
|---------------------------|-----------------------------------------------|---------------|
| `WithReadHeaderTimeout`   | Prevents Slowloris attacks                    | 2s            |
| `WithReadTimeout`         | Limits slow-body attacks                      | 5s            |
| `middleware.BodyLimit`    | Prevents memory exhaustion from large bodies  | Not set       |
| `middleware.Secure`       | HSTS, CSP, X-Frame-Options, nosniff           | Not set       |
| `middleware.CSRF`         | CSRF token validation for state-changing reqs | Not set       |
| `middleware.RateLimit`    | Prevents brute-force and DoS                  | Not set       |
| `middleware.JWT`          | Algorithm whitelist, rejects `alg:none`       | HS256 only    |
| `RunTLS`                  | TLS 1.2+, strong ciphers, X25519 preferred    | Auto-hardened |

---

## Credits

We publicly credit security researchers who responsibly disclose vulnerabilities, unless they prefer to remain anonymous. Credits appear in the GitHub security advisory and in `CHANGELOG.md`.

---

*Rudra (रुद्र) — Fierce. Fast. Fearless.*
*© Aarambh Dev Hub*
