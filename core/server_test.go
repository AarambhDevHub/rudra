package core_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"runtime"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	rudraContext "github.com/AarambhDevHub/rudra/context"
	"github.com/AarambhDevHub/rudra/core"
)

func startServer(t *testing.T, app *core.Engine) (addr string, shutdown func()) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr = ln.Addr().String()

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := app.RunListener(ln); err != nil && err != http.ErrServerClosed {
			t.Errorf("server error: %v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	shutdown = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = app.Shutdown(ctx)
		<-done
	}
	return addr, shutdown
}

func TestCustomTCPListener(t *testing.T) {
	app := core.New()

	app.GET("/", func(c *rudraContext.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	addr, shutdown := startServer(t, app)
	defer shutdown()

	resp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("expected ok, got %s", body)
	}
}

func TestTCPNoDelayOption(t *testing.T) {
	app := core.New(core.WithTCPNoDelay(true))

	app.GET("/", func(c *rudraContext.Context) error {
		return c.String(http.StatusOK, "nodelay")
	})

	addr, shutdown := startServer(t, app)
	defer shutdown()

	resp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestSOReusePortOption(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("SO_REUSEPORT is Linux-only")
	}

	app := core.New(core.WithSOReusePort(true))

	app.GET("/", func(c *rudraContext.Context) error {
		return c.String(http.StatusOK, "reuseport")
	})

	addr, shutdown := startServer(t, app)
	defer shutdown()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestTCPFastOpenOption(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("TCP_FASTOPEN is Linux-only")
	}

	app := core.New(core.WithTCPFastOpen(true))

	app.GET("/", func(c *rudraContext.Context) error {
		return c.String(http.StatusOK, "fastopen")
	})

	addr, shutdown := startServer(t, app)
	defer shutdown()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestGracefulShutdown(t *testing.T) {
	var inFlight atomic.Int32

	app := core.New(core.WithShutdownTimeout(5 * time.Second))

	app.GET("/slow", func(c *rudraContext.Context) error {
		inFlight.Add(1)
		time.Sleep(200 * time.Millisecond)
		inFlight.Add(-1)
		return c.String(http.StatusOK, "done")
	})

	addr, shutdown := startServer(t, app)

	go func() {
		_, _ = http.Get("http://" + addr + "/slow")
	}()

	time.Sleep(50 * time.Millisecond)

	shutdown()

	if inFlight.Load() != 0 {
		t.Errorf("expected 0 in-flight requests, got %d", inFlight.Load())
	}
}

func TestGracefulShutdownWithTimeout(t *testing.T) {
	app := core.New(core.WithShutdownTimeout(100 * time.Millisecond))

	app.GET("/stall", func(c *rudraContext.Context) error {
		time.Sleep(5 * time.Second)
		return c.String(http.StatusOK, "never")
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := app.RunListener(ln); err != nil && err != http.ErrServerClosed {
			t.Errorf("server error: %v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = app.Shutdown(ctx)
	<-done
}

func TestShutdownIdempotent(t *testing.T) {
	app := core.New(core.WithShutdownTimeout(5 * time.Second))

	app.GET("/", func(c *rudraContext.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	_, shutdown := startServer(t, app)
	shutdown()

	ctx := context.Background()

	err := app.Shutdown(ctx)
	if err != nil {
		t.Errorf("second shutdown should not error, got: %v", err)
	}
}

func TestShutdownTimeoutOption(t *testing.T) {
	app := core.New(core.WithShutdownTimeout(10 * time.Second))
	if app == nil {
		t.Error("expected engine, got nil")
	}
}

func TestLinuxSocketOptions(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	tcpLn, ok := ln.(*net.TCPListener)
	if !ok {
		t.Fatal("expected TCP listener")
	}

	rawFd, err := tcpLn.SyscallConn()
	if err != nil {
		t.Fatal(err)
	}

	var nodelay int
	rawFd.Control(func(fd uintptr) {
		nodelay, _ = syscall.GetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_NODELAY)
	})

	if nodelay == 0 {
		t.Log("TCP_NODELAY not set by default (expected for standard listener)")
	}
}

func TestAllSocketOptionsTogether(t *testing.T) {
	app := core.New(
		core.WithTCPNoDelay(true),
		core.WithSOReusePort(true),
		core.WithTCPFastOpen(true),
		core.WithShutdownTimeout(5*time.Second),
	)

	app.GET("/", func(c *rudraContext.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	addr, shutdown := startServer(t, app)
	defer shutdown()

	resp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
