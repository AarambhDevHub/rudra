package validator

import "sync"

// customMessages stores per-rule message overrides.
// Placeholder support: {field} and {param}.
var (
	customMessages   = make(map[string]string)
	customMessagesMu sync.RWMutex
)

// Register registers a custom validation rule globally.
// The rule is available immediately and concurrency-safe.
//
// Usage:
//
//	import (
//	    "regexp"
//	    "reflect"
//	    "github.com/AarambhDevHub/rudra/validator"
//	)
//
//	var phoneRe = regexp.MustCompile(`^[6-9]\d{9}$`)
//
//	func init() {
//	    validator.Register("indianphone", func(field string, v reflect.Value, _ string, _ reflect.Value) string {
//	        if v.Kind() != reflect.String || !phoneRe.MatchString(v.String()) {
//	            return "indianphone"
//	        }
//	        return ""
//	    })
//	    validator.RegisterMessage("indianphone", "'{field}' must be a valid 10-digit Indian phone number")
//	}
func Register(name string, fn RuleFunc) {
	rulesMu.Lock()
	allRules[name] = fn
	rulesMu.Unlock()
}

// RegisterMessage registers a custom error message template for a rule.
// Supported placeholders: {field}, {param}.
//
// Example:
//
//	validator.RegisterMessage("required", "Please fill in the '{field}' field")
func RegisterMessage(rule, message string) {
	customMessagesMu.Lock()
	customMessages[rule] = message
	customMessagesMu.Unlock()
}

// Unregister removes a previously registered custom rule.
// Built-in rules can also be removed, but this is discouraged.
func Unregister(name string) {
	rulesMu.Lock()
	delete(allRules, name)
	rulesMu.Unlock()
}
