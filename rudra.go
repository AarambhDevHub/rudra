// Package rudra is a high-performance, batteries-included Go web framework.
//
// Rudra (रुद्र) — Fierce. Fast. Fearless.
//
// Example:
//
//	package main
//
//	import (
//	    "github.com/AarambhDevHub/rudra"
//	    "github.com/AarambhDevHub/rudra/core"
//	    "github.com/AarambhDevHub/rudra/context"
//	)
//
//	func main() {
//	    app := rudra.New()
//	    app.GET("/", func(c *context.Context) error {
//	        return c.JSON(200, map[string]string{"framework": "Rudra"})
//	    })
//	    app.Run(":8080")
//	}
package rudra

import (
	"github.com/AarambhDevHub/rudra/core"
	rudraContext "github.com/AarambhDevHub/rudra/context"
)

// HandlerFunc is the core handler signature.
type HandlerFunc = core.HandlerFunc

// Context is an alias for the request context.
type Context = rudraContext.Context

// New creates a new Rudra Engine with default options.
func New(opts ...core.Option) *core.Engine {
	return core.New(opts...)
}