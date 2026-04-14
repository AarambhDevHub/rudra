//go:build !linux

package core

import (
	"context"
	"net"
	"syscall"
)

func newTCPListener(addr string, opts *Options) (net.Listener, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				if opts.TCPNoDelay {
					_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
				}
			})
		},
	}
	return lc.Listen(context.Background(), "tcp", addr)
}
