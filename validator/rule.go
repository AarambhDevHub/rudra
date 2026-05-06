package validator

import (
	"reflect"
	"sync"
)

// RuleFunc is a validation function for a single rule.
//
// Parameters:
//   - field:     the display name of the field (from json tag or struct name)
//   - value:     the reflect.Value of the field being validated
//   - param:     the rule parameter (e.g., "5" for `rudra:"min=5"`)
//   - structVal: the root struct reflect.Value (for cross-field rules like eqfield)
//
// Return value:
//   - ""          → validation passed
//   - "<ruleName>" → validation failed; the engine uses this to build a FieldError
type RuleFunc func(field string, value reflect.Value, param string, structVal reflect.Value) string

// allRules is the live registry of all active validation rules.
// Seeded from builtinRules in init(); extended via Register/Unregister.
var allRules map[string]RuleFunc

// rulesMu protects concurrent reads/writes to allRules.
var rulesMu sync.RWMutex

func init() {
	allRules = make(map[string]RuleFunc, len(builtinRules))
	for k, v := range builtinRules {
		allRules[k] = v
	}
}
