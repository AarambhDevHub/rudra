package binding

import "fmt"

// BindingError wraps an error with the binding format that produced it.
// It implements the standard error interface and supports errors.Unwrap.
type BindingError struct {
	Type string // e.g., "json", "xml", "form", "query"
	Err  error
}

func (e *BindingError) Error() string {
	return fmt.Sprintf("%s binding: %v", e.Type, e.Err)
}

func (e *BindingError) Unwrap() error { return e.Err }

// wrapErr returns nil if err is nil, otherwise wraps it as a BindingError.
func wrapErr(bindType string, err error) error {
	if err == nil {
		return nil
	}
	return &BindingError{Type: bindType, Err: err}
}
