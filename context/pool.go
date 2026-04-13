package context

import "sync"

// Pool manages a sync.Pool of Context objects.
type Pool struct {
	pool sync.Pool
}

// NewPool creates a Context pool.
func NewPool() *Pool {
	return &Pool{
		pool: sync.Pool{
			New: func() any { return New() },
		},
	}
}

// Get acquires a Context from the pool — zero allocation after warmup.
func (p *Pool) Get() *Context {
	return p.pool.Get().(*Context)
}

// Put returns a Context to the pool.
func (p *Pool) Put(c *Context) {
	c.Release()
	p.pool.Put(c)
}