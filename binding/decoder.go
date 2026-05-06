package binding

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// timeType is the reflect.Type for time.Time; used to detect time fields.
var timeType = reflect.TypeOf(time.Time{})

// timeFormats is the ordered list of layouts tried when parsing a time string.
var timeFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02",
}

// ---- Struct Field Metadata Cache ----------------------------------------
//
// Reflection is expensive. We walk each struct type exactly once and cache
// the field metadata. Subsequent requests reuse the cached data.
// The cache is keyed by (reflect.Type, tagKey) so the same struct can be
// decoded under different tag namespaces ("form", "query", "header", "path").

// fieldMeta holds pre-computed info for a single struct field.
type fieldMeta struct {
	index    []int        // nested field index path (supports embedded structs)
	name     string       // decoded tag name used to match source values
	kind     reflect.Kind // concrete kind (after pointer dereference)
	typ      reflect.Type // concrete type (after pointer dereference)
	isSlice  bool         // field is a slice (but not []byte)
	isPtr    bool         // field is a pointer
	elemKind reflect.Kind // element kind if isSlice
}

// structMeta holds pre-computed info for all fields of a struct type.
type structMeta struct {
	fields []fieldMeta
}

// cacheKey uniquely identifies a (type, tagKey) pair for the sync.Map.
type cacheKey struct {
	typ    reflect.Type
	tagKey string
}

// structCache stores *structMeta values keyed by cacheKey.
// sync.Map provides lock-free reads after the first population.
var structCache sync.Map

// getStructMeta returns (and lazily populates) the field metadata for
// a struct type under the given tag namespace.
func getStructMeta(t reflect.Type, tagKey string) *structMeta {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	key := cacheKey{typ: t, tagKey: tagKey}
	if v, ok := structCache.Load(key); ok {
		return v.(*structMeta)
	}
	meta := buildStructMeta(t, tagKey, nil)
	// LoadOrStore prevents duplicate builds under concurrent first-access.
	actual, _ := structCache.LoadOrStore(key, meta)
	return actual.(*structMeta)
}

// buildStructMeta walks a struct type recursively (for embedded fields)
// and collects all bindable fields under tagKey.
func buildStructMeta(t reflect.Type, tagKey string, indexPath []int) *structMeta {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	meta := &structMeta{}
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)

		// Build the index path for this field.
		idx := make([]int, len(indexPath)+1)
		copy(idx, indexPath)
		idx[len(indexPath)] = i

		// Transparently descend into anonymous (embedded) structs.
		// Must check Anonymous BEFORE the IsExported guard because unexported-named
		// types (e.g. `type base struct`) produce Anonymous=true, IsExported=false.
		if sf.Anonymous {
			ft := sf.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct && ft != timeType {
				embedded := buildStructMeta(ft, tagKey, idx)
				meta.fields = append(meta.fields, embedded.fields...)
				continue
			}
		}

		if !sf.IsExported() {
			continue
		}

		// Parse the tag.
		raw := sf.Tag.Get(tagKey)
		name, _, _ := strings.Cut(raw, ",")
		switch name {
		case "-":
			continue // explicitly excluded
		case "":
			// No tag: use lower-cased field name as fallback.
			name = strings.ToLower(sf.Name)
		}

		// Resolve pointer and slice.
		ft := sf.Type
		isPtr := ft.Kind() == reflect.Ptr
		if isPtr {
			ft = ft.Elem()
		}
		isSlice := ft.Kind() == reflect.Slice && ft != reflect.TypeOf([]byte(nil))

		var elemKind reflect.Kind
		if isSlice {
			elemKind = ft.Elem().Kind()
		}

		meta.fields = append(meta.fields, fieldMeta{
			index:    idx,
			name:     name,
			kind:     ft.Kind(),
			typ:      ft,
			isSlice:  isSlice,
			isPtr:    isPtr,
			elemKind: elemKind,
		})
	}
	return meta
}

// ---- Value Population ---------------------------------------------------

// mapValues populates struct v from a map[string][]string (form, query, header).
// tagKey selects which struct tag to read field names from.
func mapValues(v reflect.Value, values map[string][]string, tagKey string) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("binding: expected pointer to struct, got %s", v.Kind())
	}

	meta := getStructMeta(v.Type(), tagKey)
	for _, f := range meta.fields {
		vals, ok := values[f.name]
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
				return fmt.Errorf("binding: field %q: %w", f.name, err)
			}
			fv.Set(ptr)
			continue
		}

		if err := setReflectValue(fv, vals, f); err != nil {
			return fmt.Errorf("binding: field %q: %w", f.name, err)
		}
	}
	return nil
}

// fieldByIndex navigates a nested reflect.Value using an index path.
// It automatically initialises nil pointer fields along the path.
func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
}

// setReflectValue sets a field from a string slice.
// For slices it handles comma-split and multi-value; for scalars it takes [0].
func setReflectValue(v reflect.Value, vals []string, f fieldMeta) error {
	if f.isSlice {
		// Collect all values: each entry may be comma-delimited.
		var all []string
		for _, val := range vals {
			for _, part := range strings.Split(val, ",") {
				if p := strings.TrimSpace(part); p != "" {
					all = append(all, p)
				}
			}
		}
		slice := reflect.MakeSlice(f.typ, len(all), len(all))
		for i, s := range all {
			if err := setScalar(slice.Index(i), f.elemKind, s); err != nil {
				return err
			}
		}
		v.Set(slice)
		return nil
	}
	return setScalar(v, f.kind, vals[0])
}

// setScalar converts a string value to the target kind and sets it.
func setScalar(v reflect.Value, kind reflect.Kind, val string) error {
	switch kind {
	case reflect.String:
		v.SetString(val)

	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid bool %q", val)
		}
		v.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int %q", val)
		}
		if v.OverflowInt(n) {
			return fmt.Errorf("int overflow: %q", val)
		}
		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint %q", val)
		}
		if v.OverflowUint(n) {
			return fmt.Errorf("uint overflow: %q", val)
		}
		v.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid float %q", val)
		}
		v.SetFloat(n)

	case reflect.Struct:
		// Only time.Time is supported as a scalar struct type.
		if v.Type() == timeType {
			t, err := parseTime(val)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(t))
		}
	}
	return nil
}

// parseTime tries each layout in order until one succeeds.
func parseTime(val string) (time.Time, error) {
	for _, layout := range timeFormats {
		if t, err := time.Parse(layout, val); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time %q (expected RFC3339 or YYYY-MM-DD)", val)
}

// ---- Unsafe helpers (hot path, read-only) --------------------------------

// b2s converts a byte slice to a string without allocation.
// The caller must not modify b after the call.
//
//go:nosplit
func b2s(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// s2b converts a string to a byte slice without allocation.
// The caller must not modify the returned slice.
//
//go:nosplit
func s2b(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
