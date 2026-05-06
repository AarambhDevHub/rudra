package binding

import "reflect"

// Param is a single key-value pair from a URL path (e.g., /users/:id → {id, "42"}).
// It mirrors context.Param to avoid a circular import between binding ↔ context.
type Param struct {
	Key   string
	Value string
}

// BindPath binds URL path parameters into v.
// params is the slice of {Key, Value} pairs captured by the router.
// Struct fields are matched using the `path` tag (falls back to lower-cased field name).
//
// Example:
//
//	// Route: /orgs/:org/repos/:repo
//	type RepoParams struct {
//	    Org  string `path:"org"`
//	    Repo string `path:"repo"`
//	}
func BindPath(params []Param, v any) error {
	values := make(map[string][]string, len(params))
	for _, p := range params {
		values[p.Key] = []string{p.Value}
	}
	return wrapErr("path", mapValues(reflect.ValueOf(v), values, "path"))
}
