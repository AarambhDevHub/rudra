package router

import "strings"

// node represents a single node in the radix tree.
type node struct {
	path       string
	children   []*node
	handler    HandlerFunc
	isParam    bool
	isWildcard bool
	paramName  string
}

func (n *node) addChild(child *node) {
	n.children = append(n.children, child)
}

// search traverses the tree for the given path, populating params into ctx.
// The path includes the leading /.
func (n *node) search(path string, ctx ctxParamSetter) HandlerFunc {
	for {
		if n.isWildcard {
			ctx.SetParam(n.paramName, path)
			return n.handler
		}

		if n.isParam {
			// Capture until next / or end of path
			end := strings.IndexByte(path, '/')
			if end == -1 {
				end = len(path)
			}
			ctx.SetParam(n.paramName, path[:end])
			path = path[end:]
			if len(path) == 0 {
				return n.handler
			}
			// path now starts with "/" — search children for the remaining path
			return n.searchChildren(path, ctx)
		}

		// Static node (or root with empty path)
		if len(n.path) > 0 {
			if !strings.HasPrefix(path, n.path) {
				return nil
			}
			path = path[len(n.path):]
		}

		if len(path) == 0 {
			return n.handler
		}

		return n.searchChildren(path, ctx)
	}
}

// searchChildren tries static children first (by first byte match), then param/wildcard.
func (n *node) searchChildren(path string, ctx ctxParamSetter) HandlerFunc {
	// Static children first
	for _, child := range n.children {
		if child.isParam || child.isWildcard {
			continue
		}
		if len(child.path) > 0 && len(path) > 0 && child.path[0] == path[0] {
			if h := child.search(path, ctx); h != nil {
				return h
			}
		}
	}
	// Then param/wildcard children
	for _, child := range n.children {
		if child.isParam || child.isWildcard {
			if h := child.search(path, ctx); h != nil {
				return h
			}
		}
	}
	return nil
}

// ctxParamSetter abstracts param setting so router doesn't import context.
type ctxParamSetter interface {
	SetParam(key, value string)
}