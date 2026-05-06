package validator

import (
	"reflect"
	"strings"
	"sync"
)

const rudraTag = "rudra"

// ---- Struct Metadata Cache ----------------------------------------------
//
// Validation metadata (parsed rules) is computed once per struct type and
// cached in a sync.Map. This makes repeated validation of the same struct
// type allocation-free from the second call onward.

// ruleSpec is a single parsed rule from a rudra struct tag.
type ruleSpec struct {
	name  string // e.g., "required", "min", "email"
	param string // e.g., "5" for min=5; "" for required
}

// validationField holds pre-computed info for one field of a struct.
type validationField struct {
	index       []int        // nested field index path (supports embedded structs)
	displayName string       // name shown in error messages (json tag > field name)
	rules       []ruleSpec   // ordered list of rules to apply
	isPtr       bool         // field is a pointer
	isNested    bool         // field is a nested struct (needs recursive validation)
	nestedType  reflect.Type // concrete type after pointer dereference
	isDive      bool         // slice/map elements should be validated (dive rule)
}

// validationMeta holds pre-computed metadata for a struct type.
type validationMeta struct {
	fields []validationField
}

// metaCache is keyed by reflect.Type (after pointer deref).
var metaCache sync.Map

// getMeta returns (and lazily builds) the validation metadata for t.
func getMeta(t reflect.Type) *validationMeta {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if v, ok := metaCache.Load(t); ok {
		return v.(*validationMeta)
	}
	meta := buildMeta(t, nil)
	actual, _ := metaCache.LoadOrStore(t, meta)
	return actual.(*validationMeta)
}

// buildMeta walks a struct type and collects all validation-annotated fields.
func buildMeta(t reflect.Type, indexPath []int) *validationMeta {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	meta := &validationMeta{}

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}

		idx := make([]int, len(indexPath)+1)
		copy(idx, indexPath)
		idx[len(indexPath)] = i

		// Transparently descend into anonymous (embedded) structs.
		if sf.Anonymous {
			ft := sf.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct && ft != timeType {
				embedded := buildMeta(ft, idx)
				meta.fields = append(meta.fields, embedded.fields...)
				continue
			}
		}

		tag := sf.Tag.Get(rudraTag)
		if tag == "-" {
			continue
		}

		// Derive display name: prefer json tag, fall back to field name.
		displayName := sf.Name
		if jt := sf.Tag.Get("json"); jt != "" {
			n, _, _ := strings.Cut(jt, ",")
			if n != "" && n != "-" {
				displayName = n
			}
		}

		// Resolve pointer and determine nesting.
		ft := sf.Type
		isPtr := ft.Kind() == reflect.Ptr
		if isPtr {
			ft = ft.Elem()
		}
		isNested := ft.Kind() == reflect.Struct && ft != timeType

		// If there is no rudra tag but the field is a nested struct, still
		// register it with no rules so the engine recurses into its fields.
		if tag == "" {
			if isNested {
				meta.fields = append(meta.fields, validationField{
					index:      idx,
					isPtr:      isPtr,
					isNested:   true,
					nestedType: ft,
				})
			}
			continue
		}

		// Parse the tag into rule specs.
		parts := strings.Split(tag, ",")
		rules := make([]ruleSpec, 0, len(parts))
		isDive := false
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if part == "dive" {
				isDive = true
				continue
			}
			name, param, _ := strings.Cut(part, "=")
			rules = append(rules, ruleSpec{name: strings.TrimSpace(name), param: strings.TrimSpace(param)})
		}

		meta.fields = append(meta.fields, validationField{
			index:       idx,
			displayName: displayName,
			rules:       rules,
			isPtr:       isPtr,
			isNested:    isNested,
			nestedType:  ft,
			isDive:      isDive,
		})
	}
	return meta
}

// ---- Validation Engine --------------------------------------------------

// validateStruct runs all registered rules against v and returns any errors.
// v should be a struct value or a pointer to one.
func validateStruct(v reflect.Value) ValidationErrors {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	meta := getMeta(v.Type())
	var errs ValidationErrors

	for _, f := range meta.fields {
		fv := nestedField(v, f.index)
		if !fv.IsValid() {
			continue
		}

		// Handle nil pointer fields.
		if f.isPtr {
			if fv.IsNil() {
				// Only "required" makes sense on a nil pointer.
				for _, rule := range f.rules {
					if rule.name == "required" {
						errs = append(errs, newFieldError(f.displayName, "required", "", nil))
						break
					}
				}
				continue
			}
			fv = fv.Elem()
		}

		// Run each rule against the field value.
		rulesMu.RLock()
		for _, rule := range f.rules {
			fn, ok := allRules[rule.name]
			if !ok {
				continue
			}
			if failed := fn(f.displayName, fv, rule.param, v); failed != "" {
				var val any
				if fv.IsValid() && fv.CanInterface() {
					val = fv.Interface()
				}
				errs = append(errs, newFieldError(f.displayName, failed, rule.param, val))
			}
		}
		rulesMu.RUnlock()

		// Recurse into nested struct fields.
		if f.isNested && fv.Kind() == reflect.Struct {
			errs = append(errs, validateStruct(fv)...)
		}

		// "dive" — validate each element of a slice/map.
		if f.isDive {
			switch fv.Kind() {
			case reflect.Slice, reflect.Array:
				for i := 0; i < fv.Len(); i++ {
					elem := fv.Index(i)
					if elem.Kind() == reflect.Struct {
						errs = append(errs, validateStruct(elem)...)
					}
				}
			case reflect.Map:
				for _, key := range fv.MapKeys() {
					elem := fv.MapIndex(key)
					if elem.Kind() == reflect.Struct {
						errs = append(errs, validateStruct(elem)...)
					}
				}
			}
		}
	}

	return errs
}

// nestedField navigates a reflect.Value using an index path,
// automatically initialising nil pointers along the way.
func nestedField(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return reflect.Value{} // can't navigate further
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
}
