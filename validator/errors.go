package validator

import (
	"fmt"
	"strings"
)

// FieldError describes a single field validation failure.
type FieldError struct {
	Field   string // display name (from json tag or field name)
	Rule    string // rule that failed (e.g., "required", "email")
	Param   string // rule parameter (e.g., "5" for min=5)
	Value   any    // the field's actual value at validation time
	Message string // human-readable description
}

func (e FieldError) Error() string { return e.Message }

// ValidationErrors is a slice of FieldError.
// It implements the error interface so it can be returned directly.
type ValidationErrors []FieldError

// Error returns all messages joined by "; ".
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = e.Message
	}
	return strings.Join(msgs, "; ")
}

// First returns the first FieldError, or nil if there are none.
func (ve ValidationErrors) First() *FieldError {
	if len(ve) == 0 {
		return nil
	}
	return &ve[0]
}

// ForField returns all errors for the given display field name.
func (ve ValidationErrors) ForField(field string) []FieldError {
	var out []FieldError
	for _, e := range ve {
		if e.Field == field {
			out = append(out, e)
		}
	}
	return out
}

// Has returns true if there is at least one error for the given field.
func (ve ValidationErrors) Has(field string) bool {
	for _, e := range ve {
		if e.Field == field {
			return true
		}
	}
	return false
}

// Map converts ValidationErrors to a map[field]message for JSON serialisation.
func (ve ValidationErrors) Map() map[string]string {
	m := make(map[string]string, len(ve))
	for _, e := range ve {
		if _, exists := m[e.Field]; !exists {
			m[e.Field] = e.Message
		}
	}
	return m
}

// newFieldError creates a FieldError with a formatted message.
// If a custom message is registered for the rule, it is used instead.
func newFieldError(field, rule, param string, value any) FieldError {
	customMessagesMu.RLock()
	msg, ok := customMessages[rule]
	customMessagesMu.RUnlock()

	if !ok {
		msg = defaultMessage(field, rule, param, value)
	} else {
		// Substitute {field} and {param} placeholders.
		msg = strings.ReplaceAll(msg, "{field}", field)
		msg = strings.ReplaceAll(msg, "{param}", param)
	}

	return FieldError{
		Field:   field,
		Rule:    rule,
		Param:   param,
		Value:   value,
		Message: msg,
	}
}

// defaultMessage returns a human-readable message for built-in rules.
func defaultMessage(field, rule, param string, _ any) string {
	switch rule {
	case "required":
		return fmt.Sprintf("'%s' is required", field)
	case "min":
		return fmt.Sprintf("'%s' must be at least %s", field, param)
	case "max":
		return fmt.Sprintf("'%s' must be at most %s", field, param)
	case "len":
		return fmt.Sprintf("'%s' must be exactly %s characters", field, param)
	case "email":
		return fmt.Sprintf("'%s' must be a valid email address", field)
	case "url":
		return fmt.Sprintf("'%s' must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("'%s' must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("'%s' must be one of: %s", field, param)
	case "alphanum":
		return fmt.Sprintf("'%s' must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("'%s' must contain only numeric characters", field)
	case "alpha":
		return fmt.Sprintf("'%s' must contain only alphabetic characters", field)
	case "regexp":
		return fmt.Sprintf("'%s' does not match the required pattern", field)
	case "eqfield":
		return fmt.Sprintf("'%s' must equal '%s'", field, param)
	case "nefield":
		return fmt.Sprintf("'%s' must not equal '%s'", field, param)
	case "dive":
		return fmt.Sprintf("'%s' contains invalid elements", field)
	default:
		return fmt.Sprintf("'%s' failed rule '%s'", field, rule)
	}
}
