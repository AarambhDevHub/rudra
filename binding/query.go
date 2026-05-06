package binding

import (
	"net/http"
	"reflect"
)

type queryBinding struct{}

func (queryBinding) Name() string { return "query" }

// Bind parses the URL query string and populates v.
// Struct fields are matched using the `query` tag (falls back to lower-cased field name).
//
// Examples:
//
//	// ?page=2&limit=10&tags=go,rust
//	type Pagination struct {
//	    Page  int      `query:"page"`
//	    Limit int      `query:"limit"`
//	    Tags  []string `query:"tags"`   // comma-split or repeated: ?tags=go&tags=rust
//	}
func (queryBinding) Bind(r *http.Request, v any) error {
	return wrapErr("query", mapValues(reflect.ValueOf(v), r.URL.Query(), "query"))
}
