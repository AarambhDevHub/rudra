package core

import "time"

// Options holds all server-level configuration.
type Options struct {
	// Timeouts — tuned for low-latency APIs
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration

	// Limits
	MaxHeaderBytes int
	MaxBodyBytes   int64

	// HTTP/2 settings
	HTTP2Enabled              bool
	HTTP2MaxHandlers          int
	HTTP2MaxConcurrentStreams uint32
	HTTP2MaxReadFrameSize     uint32
	HTTP2IdleTimeout          time.Duration

	// TLS
	TLSCertFile string
	TLSKeyFile  string

	// Behavior
	StrictRouting  bool
	CaseSensitive  bool
	UnescapeParams bool
	TrustProxyIPs []string
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

// Option is a functional option for the Engine.
type Option func(*Options)

// WithReadTimeout sets the read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(o *Options) { o.ReadTimeout = d }
}

// WithWriteTimeout sets the write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(o *Options) { o.WriteTimeout = d }
}

// WithIdleTimeout sets the idle timeout.
func WithIdleTimeout(d time.Duration) Option {
	return func(o *Options) { o.IdleTimeout = d }
}

// WithReadHeaderTimeout sets the read header timeout.
func WithReadHeaderTimeout(d time.Duration) Option {
	return func(o *Options) { o.ReadHeaderTimeout = d }
}

// WithShutdownTimeout sets the graceful shutdown timeout.
func WithShutdownTimeout(d time.Duration) Option {
	return func(o *Options) { o.ShutdownTimeout = d }
}

// WithMaxHeaderBytes sets the max header bytes.
func WithMaxHeaderBytes(n int) Option {
	return func(o *Options) { o.MaxHeaderBytes = n }
}

// WithMaxBodyBytes sets the max body bytes.
func WithMaxBodyBytes(n int64) Option {
	return func(o *Options) { o.MaxBodyBytes = n }
}

// WithHTTP2 enables HTTP/2 support.
func WithHTTP2() Option {
	return func(o *Options) { o.HTTP2Enabled = true }
}

// WithStrictRouting enables strict routing (/foo ≠ /foo/).
func WithStrictRouting() Option {
	return func(o *Options) { o.StrictRouting = true }
}

// WithCaseSensitive enables case-sensitive routing.
func WithCaseSensitive() Option {
	return func(o *Options) { o.CaseSensitive = true }
}

// WithTLS sets the TLS certificate and key files.
func WithTLS(certFile, keyFile string) Option {
	return func(o *Options) {
		o.TLSCertFile = certFile
		o.TLSKeyFile = keyFile
	}
}