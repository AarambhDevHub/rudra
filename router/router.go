package router

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// HandlerFunc is the core handler type in Rudra.
type HandlerFunc func(ctx ctxParamSetter) error

// Router holds one radix tree per HTTP method plus a static route fast path.
type Router struct {
	staticRoutes map[string]map[string]HandlerFunc
	trees        map[string]*node
	namedRoutes  map[string]string
	mu           sync.RWMutex
}

// New creates a new Router.
func New() *Router {
	return &Router{
		staticRoutes: make(map[string]map[string]HandlerFunc),
		trees:        make(map[string]*node),
		namedRoutes:  make(map[string]string),
	}
}

// Add registers a route. Panics on conflicts.
func (r *Router) Add(method, path string, handler HandlerFunc, middleware ...HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		panic("rudra: path must start with '/': " + path)
	}
	if handler == nil {
		panic("rudra: handler must not be nil for path: " + path)
	}

	h := handler
	if len(middleware) > 0 {
		h = applyMiddleware(h, middleware...)
	}

	if !strings.Contains(path, ":") && !strings.Contains(path, "*") {
		if r.staticRoutes[method] == nil {
			r.staticRoutes[method] = make(map[string]HandlerFunc)
		}
		if _, exists := r.staticRoutes[method][path]; exists {
			panic(fmt.Sprintf("rudra: route conflict: %s %s already registered", method, path))
		}
		r.staticRoutes[method][path] = h
		return
	}

	root, ok := r.trees[method]
	if !ok {
		root = &node{}
		r.trees[method] = root
	}
	r.addRoute(root, path, h)
}

// Name assigns a name to a route for URL generation.
func (r *Router) Name(method, path, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.namedRoutes[name] = path
}

// URL generates a URL from a named route, substituting parameter values.
func (r *Router) URL(name string, params ...string) string {
	r.mu.RLock()
	pattern, ok := r.namedRoutes[name]
	r.mu.RUnlock()
	if !ok {
		return ""
	}

	result := pattern
	paramIdx := 0
	for i := 0; i < len(result); {
		if result[i] == ':' {
			end := i + 1
			for end < len(result) && result[end] != '/' {
				end++
			}
			if paramIdx < len(params) {
				result = result[:i] + params[paramIdx] + result[end:]
				paramIdx++
			} else {
				i = end
			}
		} else if result[i] == '*' {
			if paramIdx < len(params) {
				result = result[:i] + params[paramIdx]
				paramIdx++
			}
			break
		} else {
			i++
		}
	}
	return result
}

// Find matches the given method and path.
func (r *Router) Find(method, path string, ctx ctxParamSetter) HandlerFunc {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if methods, ok := r.staticRoutes[method]; ok {
		if h, ok := methods[path]; ok {
			return h
		}
	}

	root, ok := r.trees[method]
	if !ok {
		return nil
	}

	if path == "/" || path == "" {
		return root.handler
	}
	return root.search(path, ctx)
}

// HasMethod checks if any route exists for the given method.
func (r *Router) HasMethod(method string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.staticRoutes[method]) > 0 {
		return true
	}
	_, ok := r.trees[method]
	return ok
}

// AllMethods returns all registered HTTP methods.
func (r *Router) AllMethods() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	methods := make(map[string]struct{})
	for m := range r.staticRoutes {
		methods[m] = struct{}{}
	}
	for m := range r.trees {
		methods[m] = struct{}{}
	}
	result := make([]string, 0, len(methods))
	for m := range methods {
		result = append(result, m)
	}
	return result
}

// addRoute inserts a path into the radix tree. Paths are stored WITH leading /.
func (r *Router) addRoute(root *node, fullPath string, handler HandlerFunc) {
	if fullPath == "/" {
		if root.handler != nil {
			panic("rudra: route conflict: / already registered")
		}
		root.handler = handler
		return
	}

	// Keep the leading / — it's part of the path stored in nodes
	current := root
	remaining := fullPath

	for len(remaining) > 0 {
		if remaining[0] == '/' && len(remaining) > 1 && remaining[1] == ':' {
			// Leading /: before a param — the / belongs to the static part
			// but the : starts a param. Include the / in a static segment.
			// Actually, handle the / as part of the path, then the : starts
			// a new param segment on next iteration.
			// We need to check: is the / already consumed by a parent?
			// If we're at root (empty path), we need to add a "/" static node
			// or add the param child directly.

			// Simpler: just treat "/" before a param as a static prefix of the param
			// We'll handle it by letting the param consume the leading / too.
			// Actually no. Let me rethink.
		}

		if remaining[0] == ':' {
			// Param segment
			end := strings.IndexByte(remaining, '/')
			var seg, rest string
			if end == -1 {
				seg = remaining
				rest = ""
			} else {
				seg = remaining[:end]
				rest = remaining[end:]
			}

			paramName := seg[1:]

			var paramChild *node
			for _, child := range current.children {
				if child.isParam && child.paramName == paramName {
					paramChild = child
					break
				}
			}
			if paramChild == nil {
				for _, child := range current.children {
					if child.isWildcard {
						panic(fmt.Sprintf("rudra: route conflict: param ':%s' conflicts with wildcard '*%s'", paramName, child.paramName))
					}
				}
				paramChild = &node{
					isParam:   true,
					paramName: paramName,
				}
				current.addChild(paramChild)
			}

			current = paramChild
			remaining = rest
			continue
		}

		if remaining[0] == '*' {
			paramName := remaining[1:]
			if paramName == "" {
				paramName = "wildcard"
			}
			for _, child := range current.children {
				if child.isWildcard {
					panic(fmt.Sprintf("rudra: route conflict: wildcard '*%s' already exists", child.paramName))
				}
				if child.isParam {
					panic(fmt.Sprintf("rudra: route conflict: wildcard '*%s' conflicts with param ':%s'", paramName, child.paramName))
				}
			}
			wildcardNode := &node{
				isWildcard: true,
				paramName:  paramName,
				handler:    handler,
			}
			current.addChild(wildcardNode)
			return
		}

		// Static segment: extract up to next /: or /*
		end := len(remaining)
		for i := 0; i < len(remaining); i++ {
			if remaining[i] == '/' && i+1 < len(remaining) && (remaining[i+1] == ':' || remaining[i+1] == '*') {
				end = i + 1 // include the trailing /
				break
			}
			if (remaining[i] == ':' || remaining[i] == '*') && (i == 0 || remaining[i-1] != '/') {
				end = i
				break
			}
		}

		seg := remaining[:end]
		rest := remaining[end:]

		// Look for a child with common prefix
		var matched *node
		for _, child := range current.children {
			if child.isParam || child.isWildcard {
				continue
			}
			if len(child.path) > 0 && len(seg) > 0 && child.path[0] == seg[0] {
				commonLen := commonPrefix(seg, child.path)
				if commonLen > 0 {
					matched = child
					if commonLen < len(child.path) {
						splitNode(child, commonLen)
					}
					seg = seg[commonLen:]
					break
				}
			}
		}

		if matched != nil {
			current = matched
			remaining = seg + rest
			continue
		}

		newNode := &node{path: seg}
		current.addChild(newNode)

		if rest == "" {
			newNode.handler = handler
			return
		}
		current = newNode
		remaining = rest
	}

	if current.handler != nil {
		panic(fmt.Sprintf("rudra: route conflict: %s already registered", fullPath))
	}
	current.handler = handler
}

func commonPrefix(a, b string) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	i := 0
	for i < minLen && a[i] == b[i] {
		i++
	}
	return i
}

func splitNode(n *node, pos int) {
	prefix := n.path[:pos]
	suffix := n.path[pos:]

	child := &node{
		path:     suffix,
		children: n.children,
		handler:  n.handler,
	}

	n.path = prefix
	n.children = []*node{child}
	n.handler = nil
}

func applyMiddleware(h HandlerFunc, middleware ...HandlerFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		mw := middleware[i]
		next := h
		h = func(ctx ctxParamSetter) error {
			if c, ok := ctx.(interface{ SetNext(func() error) }); ok {
				c.SetNext(func() error { return next(ctx) })
			}
			return mw(ctx)
		}
	}
	return h
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}