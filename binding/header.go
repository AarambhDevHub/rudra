package binding

import (
	"net/http"
	"net/textproto"
	"reflect"
)

type headerBinding struct{}

func (headerBinding) Name() string { return "header" }

// Bind reads HTTP request headers and populates v.
// Struct fields are matched using the `header` tag against canonical header names.
// Tag values are canonicalised with textproto.CanonicalMIMEHeaderKey, so
// both `header:"x-request-id"` and `header:"X-Request-Id"` work correctly.
//
// Example:
//
//	type AuthHeaders struct {
//	    Authorization string `header:"Authorization"`
//	    RequestID     string `header:"X-Request-Id"`
//	    ContentType   string `header:"Content-Type"`
//	}
func (headerBinding) Bind(r *http.Request, v any) error {
	// Canonicalise tag names so case-insensitive matching works.
	values := make(map[string][]string, len(r.Header))
	for k, vs := range r.Header {
		values[textproto.CanonicalMIMEHeaderKey(k)] = vs
	}

	// Normalise the struct tags too via a custom tag-normaliser wrapper.
	return wrapErr("header", mapValuesCanon(reflect.ValueOf(v), values))
}

// mapValuesCanon is like mapValues("header") but canonicalises field tag names
// before lookup so `header:"authorization"` matches "Authorization".
func mapValuesCanon(v reflect.Value, values map[string][]string) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	meta := getStructMeta(v.Type(), "header")
	for _, f := range meta.fields {
		// Canonicalise the tag-derived name.
		canon := textproto.CanonicalMIMEHeaderKey(f.name)
		vals, ok := values[canon]
		if !ok || len(vals) == 0 {
			continue
		}

		fv := fieldByIndex(v, f.index)
		if !fv.CanSet() {
			continue
		}

		if f.isPtr {
			ptr := reflect.New(f.typ)
			if err := setReflectValue(ptr.Elem(), vals, f); err != nil {
				return err
			}
			fv.Set(ptr)
			continue
		}

		if err := setReflectValue(fv, vals, f); err != nil {
			return err
		}
	}
	return nil
}
