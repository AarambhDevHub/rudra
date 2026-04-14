//go:build linux

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
				if opts.SOReusePort {
					_ = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, 0x0F, 1) // SO_REUSEPORT = 15
				}
				if opts.TCPFastOpen {
					_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, 0x17, 5) // TCP_FASTOPEN = 23
				}
			})
		},
	}
	return lc.Listen(context.Background(), "tcp", addr)
}
