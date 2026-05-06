package validator

import "reflect"

// Validate runs rudra struct-tag validation against v.
// v must be a pointer to a struct.
// Returns nil if all rules pass, or a ValidationErrors slice on failure.
func Validate(v any) error {
	rv := reflect.ValueOf(v)
	errs := validateStruct(rv)
	if len(errs) == 0 {
		return nil
	}
	return errs
}
