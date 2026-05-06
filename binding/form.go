package binding

import (
	"net/http"
	"reflect"
)

// formBinding handles application/x-www-form-urlencoded requests.
type formBinding struct{}

func (formBinding) Name() string { return "form" }

// Bind parses the URL-encoded form body and populates v.
// Struct fields are matched using the `form` tag (falls back to lower-cased field name).
func (formBinding) Bind(r *http.Request, v any) error {
	if err := r.ParseForm(); err != nil {
		return wrapErr("form", err)
	}
	return wrapErr("form", mapValues(reflect.ValueOf(v), r.Form, "form"))
}

// multipartBinding handles multipart/form-data requests.
type multipartBinding struct {
	MaxMemory int64 // max bytes buffered in memory (remainder spills to disk)
}

func (multipartBinding) Name() string { return "multipart" }

// Bind parses the multipart form and populates v with form field values.
// File uploads are not bound into the struct; use c.FormFile() for files.
// Struct fields are matched using the `form` tag.
func (mb multipartBinding) Bind(r *http.Request, v any) error {
	maxMem := mb.MaxMemory
	if maxMem <= 0 {
		maxMem = MaxBodyBytes
	}

	if err := r.ParseMultipartForm(maxMem); err != nil {
		// Fall back to URL-encoded form if not multipart.
		if err == http.ErrNotMultipart {
			if ferr := r.ParseForm(); ferr != nil {
				return wrapErr("multipart", ferr)
			}
			return wrapErr("multipart", mapValues(reflect.ValueOf(v), r.Form, "form"))
		}
		return wrapErr("multipart", err)
	}

	// Merge multipart form values (excludes file headers).
	values := make(map[string][]string, len(r.MultipartForm.Value))
	for k, vs := range r.MultipartForm.Value {
		values[k] = vs
	}
	return wrapErr("multipart", mapValues(reflect.ValueOf(v), values, "form"))
}
