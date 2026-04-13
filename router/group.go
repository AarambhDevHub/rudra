package router

// Group represents a set of routes sharing a common prefix and middleware.
type Group struct {
	Prefix     string
	Middleware []HandlerFunc
	Router     *Router
}

// NewGroup creates a route group.
func NewGroup(prefix string, r *Router, middleware ...HandlerFunc) *Group {
	return &Group{
		Prefix:     prefix,
		Middleware: middleware,
		Router:     r,
	}
}

// Group creates a sub-group with an additional prefix and optional middleware.
func (g *Group) Group(prefix string, middleware ...HandlerFunc) *Group {
	return &Group{
		Prefix:     g.Prefix + prefix,
		Middleware: append(g.Middleware, middleware...),
		Router:     g.Router,
	}
}

// Use adds middleware to the group.
func (g *Group) Use(middleware ...HandlerFunc) {
	g.Middleware = append(g.Middleware, middleware...)
}

// GET registers a GET handler on this group.
func (g *Group) GET(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.Router.Add("GET", g.Prefix+path, h, append(g.Middleware, mw...)...)
}

// POST registers a POST handler on this group.
func (g *Group) POST(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.Router.Add("POST", g.Prefix+path, h, append(g.Middleware, mw...)...)
}

// PUT registers a PUT handler on this group.
func (g *Group) PUT(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.Router.Add("PUT", g.Prefix+path, h, append(g.Middleware, mw...)...)
}

// PATCH registers a PATCH handler on this group.
func (g *Group) PATCH(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.Router.Add("PATCH", g.Prefix+path, h, append(g.Middleware, mw...)...)
}

// DELETE registers a DELETE handler on this group.
func (g *Group) DELETE(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.Router.Add("DELETE", g.Prefix+path, h, append(g.Middleware, mw...)...)
}

// OPTIONS registers an OPTIONS handler on this group.
func (g *Group) OPTIONS(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.Router.Add("OPTIONS", g.Prefix+path, h, append(g.Middleware, mw...)...)
}

// HEAD registers a HEAD handler on this group.
func (g *Group) HEAD(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.Router.Add("HEAD", g.Prefix+path, h, append(g.Middleware, mw...)...)
}