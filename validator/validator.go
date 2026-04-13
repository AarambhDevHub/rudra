package validator

// Validator is the interface for struct validation.
type Validator interface {
	Validate(v any) error
}

// DefaultValidator provides built-in validation using struct tags.
type DefaultValidator struct{}

// Validate validates a struct using rudra tags.
func (d *DefaultValidator) Validate(v any) error {
	// TODO: implement in Phase 2 (0.2.5)
	return nil
}