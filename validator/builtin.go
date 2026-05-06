package validator

import (
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// builtinRules is populated in init() and merged into allRules.
var builtinRules = map[string]RuleFunc{
	"required": validateRequired,
	"min":      validateMin,
	"max":      validateMax,
	"len":      validateLen,
	"email":    validateEmail,
	"url":      validateURL,
	"uuid":     validateUUID,
	"oneof":    validateOneOf,
	"alphanum": validateAlphanum,
	"alpha":    validateAlpha,
	"numeric":  validateNumeric,
	"regexp":   validateRegexp,
	"eqfield":  validateEqField,
	"nefield":  validateNeField,
}

// timeType is the reflect.Type for time.Time.
var timeType = reflect.TypeOf(time.Time{})

// ---- required -----------------------------------------------------------

// validateRequired fails for zero values: empty string, 0, false, nil,
// nil pointer, nil/empty slice or map.
func validateRequired(field string, v reflect.Value, _ string, _ reflect.Value) string {
	if isZero(v) {
		return "required"
	}
	return ""
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map, reflect.Chan:
		return v.IsNil() || v.Len() == 0
	case reflect.Struct:
		if v.Type() == timeType {
			return v.Interface().(time.Time).IsZero()
		}
		return v.IsZero()
	}
	return false
}

// ---- min ----------------------------------------------------------------

// validateMin enforces a minimum:
//   - string/slice/map/array → minimum length
//   - numeric               → minimum value
func validateMin(field string, v reflect.Value, param string, _ reflect.Value) string {
	n, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return ""
	}
	switch v.Kind() {
	case reflect.String:
		if float64(len([]rune(v.String()))) < n {
			return "min"
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if float64(v.Len()) < n {
			return "min"
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) < n {
			return "min"
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) < n {
			return "min"
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() < n {
			return "min"
		}
	}
	return ""
}

// ---- max ----------------------------------------------------------------

func validateMax(field string, v reflect.Value, param string, _ reflect.Value) string {
	n, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return ""
	}
	switch v.Kind() {
	case reflect.String:
		if float64(len([]rune(v.String()))) > n {
			return "max"
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if float64(v.Len()) > n {
			return "max"
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) > n {
			return "max"
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) > n {
			return "max"
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() > n {
			return "max"
		}
	}
	return ""
}

// ---- len ----------------------------------------------------------------

func validateLen(field string, v reflect.Value, param string, _ reflect.Value) string {
	n, err := strconv.Atoi(param)
	if err != nil {
		return ""
	}
	switch v.Kind() {
	case reflect.String:
		if len([]rune(v.String())) != n {
			return "len"
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if v.Len() != n {
			return "len"
		}
	}
	return ""
}

// ---- email --------------------------------------------------------------

// emailRe is compiled once at package init — never per-validation.
// Covers the practical 99 % of email formats (not full RFC 5322).
var emailRe = regexp.MustCompile(
	`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
)

func validateEmail(field string, v reflect.Value, _ string, _ reflect.Value) string {
	if v.Kind() != reflect.String {
		return ""
	}
	if !emailRe.MatchString(v.String()) {
		return "email"
	}
	return ""
}

// ---- url ----------------------------------------------------------------

func validateURL(field string, v reflect.Value, _ string, _ reflect.Value) string {
	if v.Kind() != reflect.String {
		return ""
	}
	u, err := url.ParseRequestURI(v.String())
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "url"
	}
	return ""
}

// ---- uuid ---------------------------------------------------------------

// uuidRe matches UUID v4 (and all version/variant combinations).
var uuidRe = regexp.MustCompile(
	`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
)

func validateUUID(field string, v reflect.Value, _ string, _ reflect.Value) string {
	if v.Kind() != reflect.String {
		return ""
	}
	if !uuidRe.MatchString(v.String()) {
		return "uuid"
	}
	return ""
}

// ---- oneof --------------------------------------------------------------

func validateOneOf(field string, v reflect.Value, param string, _ reflect.Value) string {
	allowed := strings.Fields(param) // space-separated list
	if len(allowed) == 0 {
		return ""
	}
	var s string
	switch v.Kind() {
	case reflect.String:
		s = v.String()
	default:
		return ""
	}
	for _, a := range allowed {
		if s == a {
			return ""
		}
	}
	return "oneof"
}

// ---- alphanum / alpha / numeric -----------------------------------------

var (
	alphanumRe = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	numericRe  = regexp.MustCompile(`^[0-9]+$`)
)

func validateAlphanum(field string, v reflect.Value, _ string, _ reflect.Value) string {
	if v.Kind() != reflect.String || v.Len() == 0 {
		return ""
	}
	if !alphanumRe.MatchString(v.String()) {
		return "alphanum"
	}
	return ""
}

func validateAlpha(field string, v reflect.Value, _ string, _ reflect.Value) string {
	if v.Kind() != reflect.String || v.Len() == 0 {
		return ""
	}
	for _, r := range v.String() {
		if !unicode.IsLetter(r) {
			return "alpha"
		}
	}
	return ""
}

func validateNumeric(field string, v reflect.Value, _ string, _ reflect.Value) string {
	if v.Kind() != reflect.String || v.Len() == 0 {
		return ""
	}
	if !numericRe.MatchString(v.String()) {
		return "numeric"
	}
	return ""
}

// ---- regexp -------------------------------------------------------------

// regexpCache caches compiled patterns (keyed by pattern string).
// Patterns are compiled once; subsequent validations are O(match) only.
var regexpCache sync.Map // map[string]*regexp.Regexp

func validateRegexp(field string, v reflect.Value, param string, _ reflect.Value) string {
	if v.Kind() != reflect.String {
		return ""
	}
	var re *regexp.Regexp
	if cached, ok := regexpCache.Load(param); ok {
		re = cached.(*regexp.Regexp)
	} else {
		compiled, err := regexp.Compile(param)
		if err != nil {
			return "" // invalid pattern — skip silently
		}
		actual, _ := regexpCache.LoadOrStore(param, compiled)
		re = actual.(*regexp.Regexp)
	}
	if !re.MatchString(v.String()) {
		return "regexp"
	}
	return ""
}

// ---- cross-field: eqfield / nefield ------------------------------------

func validateEqField(field string, v reflect.Value, param string, sv reflect.Value) string {
	if !sv.IsValid() {
		return ""
	}
	other := sv.FieldByName(param)
	if !other.IsValid() {
		return ""
	}
	if !reflect.DeepEqual(v.Interface(), other.Interface()) {
		return "eqfield"
	}
	return ""
}

func validateNeField(field string, v reflect.Value, param string, sv reflect.Value) string {
	if !sv.IsValid() {
		return ""
	}
	other := sv.FieldByName(param)
	if !other.IsValid() {
		return ""
	}
	if reflect.DeepEqual(v.Interface(), other.Interface()) {
		return "nefield"
	}
	return ""
}
