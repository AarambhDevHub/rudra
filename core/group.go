package core

import (
	rudraContext "github.com/AarambhDevHub/rudra/context"
	"github.com/AarambhDevHub/rudra/router"
)

// Group wraps a router.Group and provides the core HandlerFunc signature.
type Group struct {
	group  *router.Group
	engine *Engine
}

// Group creates a sub-group with an additional prefix and optional middleware.
func (g *Group) Group(prefix string, mw ...HandlerFunc) *Group {
	routerMW := make([]router.HandlerFunc, len(mw))
	for i, m := range mw {
		m := m
		routerMW[i] = func(ctx router.ContextParamSetter) error {
			return m(ctx.(*rudraContext.Context))
		}
	}
	rg := g.group.Group(prefix, routerMW...)
	return &Group{group: rg, engine: g.engine}
}

// Use adds middleware to the group.
func (g *Group) Use(mw ...HandlerFunc) {
	routerMW := make([]router.HandlerFunc, len(mw))
	for i, m := range mw {
		m := m
		routerMW[i] = func(ctx router.ContextParamSetter) error {
			return m(ctx.(*rudraContext.Context))
		}
	}
	g.group.Use(routerMW...)
}

// GET registers a GET handler on this group.
func (g *Group) GET(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.addRoute("GET", path, h, mw...)
}

// POST registers a POST handler on this group.
func (g *Group) POST(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.addRoute("POST", path, h, mw...)
}

// PUT registers a PUT handler on this group.
func (g *Group) PUT(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.addRoute("PUT", path, h, mw...)
}

// PATCH registers a PATCH handler on this group.
func (g *Group) PATCH(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.addRoute("PATCH", path, h, mw...)
}

// DELETE registers a DELETE handler on this group.
func (g *Group) DELETE(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.addRoute("DELETE", path, h, mw...)
}

// OPTIONS registers an OPTIONS handler on this group.
func (g *Group) OPTIONS(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.addRoute("OPTIONS", path, h, mw...)
}

// HEAD registers a HEAD handler on this group.
func (g *Group) HEAD(path string, h HandlerFunc, mw ...HandlerFunc) {
	g.addRoute("HEAD", path, h, mw...)
}

func (g *Group) addRoute(method, path string, h HandlerFunc, mw ...HandlerFunc) {
	allMW := make([]router.HandlerFunc, 0, len(mw))
	for _, m := range mw {
		m := m
		allMW = append(allMW, func(ctx router.ContextParamSetter) error {
			return m(ctx.(*rudraContext.Context))
		})
	}

	routerHandler := func(ctx router.ContextParamSetter) error {
		return h(ctx.(*rudraContext.Context))
	}

	g.group.Router.Add(method, g.group.Prefix+path, routerHandler, allMW...)
}