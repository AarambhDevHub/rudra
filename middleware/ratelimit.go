package middleware

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	rudraContext "github.com/AarambhDevHub/rudra/context"
	"github.com/AarambhDevHub/rudra/errors"
)

// RateLimitConfig defines the config for the RateLimit middleware.
type RateLimitConfig struct {
	// Rate is the number of tokens added per second.
	// Default: 10.
	Rate float64

	// Burst is the maximum number of tokens (bucket capacity).
	// Default: 20.
	Burst int

	// KeyFunc extracts the rate limit key from the request.
	// Default: c.RealIP().
	KeyFunc func(c *rudraContext.Context) string

	// OnLimit is an optional handler called when the rate limit is exceeded.
	// If nil, returns 429 Too Many Requests.
	OnLimit func(c *rudraContext.Context) error

	// ExpiresIn is the duration after which an idle bucket is cleaned up.
	// Default: 5 minutes.
	ExpiresIn time.Duration
}

// DefaultRateLimitConfig returns a RateLimitConfig with sane defaults.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Rate:      10,
		Burst:     20,
		ExpiresIn: 5 * time.Minute,
		KeyFunc:   func(c *rudraContext.Context) string { return c.RealIP() },
	}
}

// tokenBucket implements a token bucket rate limiter.
type tokenBucket struct {
	tokens   float64
	lastTime time.Time
	mu       sync.Mutex
}

// allow checks if a request is allowed and consumes a token.
func (b *tokenBucket) allow(rate float64, burst int, now time.Time) (bool, float64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Add tokens based on elapsed time.
	elapsed := now.Sub(b.lastTime).Seconds()
	b.tokens += elapsed * rate
	if b.tokens > float64(burst) {
		b.tokens = float64(burst)
	}
	b.lastTime = now

	if b.tokens < 1 {
		return false, b.tokens
	}

	b.tokens--
	return true, b.tokens
}

// rateLimitStore manages per-key token buckets.
type rateLimitStore struct {
	buckets sync.Map // map[string]*tokenBucket
}

func (s *rateLimitStore) getBucket(key string, burst int) *tokenBucket {
	if v, ok := s.buckets.Load(key); ok {
		return v.(*tokenBucket)
	}
	b := &tokenBucket{
		tokens:   float64(burst),
		lastTime: time.Now(),
	}
	actual, _ := s.buckets.LoadOrStore(key, b)
	return actual.(*tokenBucket)
}

func (s *rateLimitStore) cleanup(expiresIn time.Duration) {
	cutoff := time.Now().Add(-expiresIn)
	s.buckets.Range(func(key, value any) bool {
		b := value.(*tokenBucket)
		b.mu.Lock()
		lastTime := b.lastTime
		b.mu.Unlock()
		if lastTime.Before(cutoff) {
			s.buckets.Delete(key)
		}
		return true
	})
}

// RateLimit returns a token bucket rate limiting middleware.
//
// Each unique key (default: client IP) gets its own bucket with the configured
// rate and burst capacity. When the bucket is empty, requests receive 429
// Too Many Requests with Retry-After and X-RateLimit-* headers.
//
// A background goroutine periodically cleans up expired buckets to prevent
// memory leaks from short-lived clients.
//
// Usage:
//
//	app.Use(middleware.RateLimit())
//	app.Use(middleware.RateLimit(middleware.RateLimitConfig{
//	    Rate:  100,
//	    Burst: 50,
//	    KeyFunc: func(c *context.Context) string { return c.RealIP() },
//	}))
func RateLimit(config ...RateLimitConfig) func(*rudraContext.Context) error {
	cfg := DefaultRateLimitConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Rate <= 0 {
		cfg.Rate = 10
	}
	if cfg.Burst <= 0 {
		cfg.Burst = 20
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(c *rudraContext.Context) string { return c.RealIP() }
	}
	if cfg.ExpiresIn <= 0 {
		cfg.ExpiresIn = 5 * time.Minute
	}

	store := &rateLimitStore{}

	// Background cleanup goroutine.
	go func() {
		ticker := time.NewTicker(cfg.ExpiresIn)
		defer ticker.Stop()
		for range ticker.C {
			store.cleanup(cfg.ExpiresIn)
		}
	}()

	return func(c *rudraContext.Context) error {
		key := cfg.KeyFunc(c)
		bucket := store.getBucket(key, cfg.Burst)
		now := time.Now()

		allowed, remaining := bucket.allow(cfg.Rate, cfg.Burst, now)

		// Set rate limit headers on every response.
		h := c.Writer().Header()
		h.Set("X-RateLimit-Limit", strconv.Itoa(cfg.Burst))
		h.Set("X-RateLimit-Remaining", strconv.Itoa(int(math.Max(0, remaining))))

		// Calculate reset time (when bucket will be full again).
		tokensNeeded := float64(cfg.Burst) - remaining
		resetSeconds := tokensNeeded / cfg.Rate
		resetTime := now.Add(time.Duration(resetSeconds * float64(time.Second)))
		h.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			// Calculate retry-after: time until next token.
			retryAfter := (1 - remaining) / cfg.Rate
			if retryAfter < 1 {
				retryAfter = 1
			}
			h.Set("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter))))

			if cfg.OnLimit != nil {
				return cfg.OnLimit(c)
			}
			return errors.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
		}

		return c.Next()
	}
}
