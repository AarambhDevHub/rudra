package validator_test

import (
	"reflect"
	"testing"

	"github.com/AarambhDevHub/rudra/validator"
)

// ---- Test helpers --------------------------------------------------------

func assertValid(t *testing.T, v any) {
	t.Helper()
	if err := validator.Validate(v); err != nil {
		t.Errorf("expected valid, got errors: %v", err)
	}
}

func assertInvalid(t *testing.T, v any, expectedRules ...string) {
	t.Helper()
	err := validator.Validate(v)
	if err == nil {
		t.Error("expected validation errors, got nil")
		return
	}
	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
	for _, rule := range expectedRules {
		found := false
		for _, fe := range ve {
			if fe.Rule == rule {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected rule %q to fail; errors: %v", rule, ve)
		}
	}
}

// ---- Nil / zero inputs ---------------------------------------------------

func TestValidateNil(t *testing.T) {
	if err := validator.Validate(nil); err != nil {
		t.Error("expected nil for nil input")
	}
}

func TestValidateNilPointer(t *testing.T) {
	type S struct {
		Name string `rudra:"required"`
	}
	var p *S
	if err := validator.Validate(p); err != nil {
		t.Error("expected nil for nil pointer")
	}
}

// ---- required -----------------------------------------------------------

func TestRequiredString(t *testing.T) {
	type S struct {
		Name string `rudra:"required"`
	}
	assertInvalid(t, &S{}, "required")
	assertValid(t, &S{Name: "Rudra"})
}

func TestRequiredInt(t *testing.T) {
	type S struct {
		Age int `rudra:"required"`
	}
	assertInvalid(t, &S{Age: 0}, "required")
	assertValid(t, &S{Age: 1})
}

func TestRequiredBool(t *testing.T) {
	type S struct {
		Active bool `rudra:"required"`
	}
	assertInvalid(t, &S{Active: false}, "required")
	assertValid(t, &S{Active: true})
}

func TestRequiredSlice(t *testing.T) {
	type S struct {
		Tags []string `rudra:"required"`
	}
	assertInvalid(t, &S{}, "required")
	assertValid(t, &S{Tags: []string{"go"}})
}

func TestRequiredPointer(t *testing.T) {
	type S struct {
		Val *int `rudra:"required"`
	}
	assertInvalid(t, &S{}, "required")
	n := 42
	assertValid(t, &S{Val: &n})
}

// ---- min / max ----------------------------------------------------------

func TestMinString(t *testing.T) {
	type S struct {
		Name string `rudra:"min=3"`
	}
	assertInvalid(t, &S{Name: "hi"}, "min")
	assertValid(t, &S{Name: "Rudra"})
}

func TestMaxString(t *testing.T) {
	type S struct {
		Name string `rudra:"max=5"`
	}
	assertInvalid(t, &S{Name: "toolongvalue"}, "max")
	assertValid(t, &S{Name: "ok"})
}

func TestMinInt(t *testing.T) {
	type S struct {
		Age int `rudra:"min=18"`
	}
	assertInvalid(t, &S{Age: 17}, "min")
	assertValid(t, &S{Age: 18})
}

func TestMaxInt(t *testing.T) {
	type S struct {
		Score int `rudra:"max=100"`
	}
	assertInvalid(t, &S{Score: 101}, "max")
	assertValid(t, &S{Score: 100})
}

func TestMinFloat(t *testing.T) {
	type S struct {
		Price float64 `rudra:"min=0.01"`
	}
	assertInvalid(t, &S{Price: 0.001}, "min")
	assertValid(t, &S{Price: 1.0})
}

func TestMinSlice(t *testing.T) {
	type S struct {
		Tags []string `rudra:"min=2"`
	}
	assertInvalid(t, &S{Tags: []string{"one"}}, "min")
	assertValid(t, &S{Tags: []string{"one", "two"}})
}

// ---- len ----------------------------------------------------------------

func TestLenString(t *testing.T) {
	type S struct {
		Pin string `rudra:"len=4"`
	}
	assertInvalid(t, &S{Pin: "123"}, "len")
	assertInvalid(t, &S{Pin: "12345"}, "len")
	assertValid(t, &S{Pin: "1234"})
}

// ---- email --------------------------------------------------------------

func TestEmail(t *testing.T) {
	type S struct {
		Email string `rudra:"email"`
	}
	assertInvalid(t, &S{Email: "notanemail"}, "email")
	assertInvalid(t, &S{Email: "missing@"}, "email")
	assertValid(t, &S{Email: "user@example.com"})
	assertValid(t, &S{Email: "user+tag@sub.domain.io"})
}

// ---- url ----------------------------------------------------------------

func TestURL(t *testing.T) {
	type S struct {
		Website string `rudra:"url"`
	}
	assertInvalid(t, &S{Website: "not-a-url"}, "url")
	assertInvalid(t, &S{Website: "ftp://example.com"}, "url")
	assertValid(t, &S{Website: "https://rudra.dev"})
	assertValid(t, &S{Website: "http://localhost:8080"})
}

// ---- uuid ---------------------------------------------------------------

func TestUUID(t *testing.T) {
	type S struct {
		ID string `rudra:"uuid"`
	}
	assertInvalid(t, &S{ID: "not-a-uuid"}, "uuid")
	assertValid(t, &S{ID: "550e8400-e29b-41d4-a716-446655440000"})
	assertValid(t, &S{ID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"})
}

// ---- oneof --------------------------------------------------------------

func TestOneOf(t *testing.T) {
	type S struct {
		Role string `rudra:"oneof=admin user guest"`
	}
	assertInvalid(t, &S{Role: "superuser"}, "oneof")
	assertValid(t, &S{Role: "admin"})
	assertValid(t, &S{Role: "guest"})
}

// ---- alphanum / alpha / numeric -----------------------------------------

func TestAlphanum(t *testing.T) {
	type S struct {
		Code string `rudra:"alphanum"`
	}
	assertInvalid(t, &S{Code: "abc-123"}, "alphanum")
	assertValid(t, &S{Code: "abc123"})
}

func TestAlpha(t *testing.T) {
	type S struct {
		Name string `rudra:"alpha"`
	}
	assertInvalid(t, &S{Name: "John123"}, "alpha")
	assertValid(t, &S{Name: "John"})
}

func TestNumeric(t *testing.T) {
	type S struct {
		Phone string `rudra:"numeric"`
	}
	assertInvalid(t, &S{Phone: "123abc"}, "numeric")
	assertValid(t, &S{Phone: "9876543210"})
}

// ---- regexp -------------------------------------------------------------

func TestRegexp(t *testing.T) {
	type S struct {
		Phone string `rudra:"regexp=^[6-9]\\d{9}$"`
	}
	assertInvalid(t, &S{Phone: "1234567890"}, "regexp")
	assertValid(t, &S{Phone: "9876543210"})
}

// ---- eqfield / nefield --------------------------------------------------

func TestEqField(t *testing.T) {
	type S struct {
		Password string `rudra:"required"`
		Confirm  string `rudra:"eqfield=Password"`
	}
	assertInvalid(t, &S{Password: "secret", Confirm: "different"}, "eqfield")
	assertValid(t, &S{Password: "secret", Confirm: "secret"})
}

func TestNeField(t *testing.T) {
	type S struct {
		OldPass string `rudra:"required"`
		NewPass string `rudra:"nefield=OldPass"`
	}
	assertInvalid(t, &S{OldPass: "same", NewPass: "same"}, "nefield")
	assertValid(t, &S{OldPass: "old", NewPass: "new"})
}

// ---- Multiple rules on one field ----------------------------------------

func TestMultipleRules(t *testing.T) {
	type S struct {
		Name string `rudra:"required,min=2,max=64,alphanum"`
	}
	assertInvalid(t, &S{}, "required")
	assertInvalid(t, &S{Name: "a"}, "min")
	assertInvalid(t, &S{Name: "valid-but-has-dash"}, "alphanum")
	assertValid(t, &S{Name: "Arjun"})
}

// ---- Nested structs -----------------------------------------------------

func TestNestedStruct(t *testing.T) {
	type Address struct {
		City string `json:"city" rudra:"required"`
	}
	type User struct {
		Name    string  `json:"name"    rudra:"required"`
		Address Address `json:"address"`
	}

	assertInvalid(t, &User{Name: "Arjun", Address: Address{City: ""}}, "required")
	assertValid(t, &User{Name: "Arjun", Address: Address{City: "Surat"}})
}

// ---- Embedded structs ---------------------------------------------------

func TestEmbeddedStruct(t *testing.T) {
	type Base struct {
		CreatedBy string `json:"created_by" rudra:"required"`
	}
	type Event struct {
		Base
		Name string `json:"name" rudra:"required"`
	}
	assertInvalid(t, &Event{Name: "Launch"}, "required") // Base.CreatedBy missing
	assertValid(t, &Event{Base: Base{CreatedBy: "admin"}, Name: "Launch"})
}

// ---- ValidationErrors helpers -------------------------------------------

func TestValidationErrorsMap(t *testing.T) {
	type S struct {
		Name  string `json:"name"  rudra:"required"`
		Email string `json:"email" rudra:"required,email"`
	}
	err := validator.Validate(&S{})
	if err == nil {
		t.Fatal("expected errors")
	}
	ve := err.(validator.ValidationErrors)
	m := ve.Map()
	if _, ok := m["name"]; !ok {
		t.Error("expected 'name' key in error map")
	}
}

func TestValidationErrorsFirst(t *testing.T) {
	type S struct {
		Name string `rudra:"required"`
	}
	err := validator.Validate(&S{})
	ve := err.(validator.ValidationErrors)
	if ve.First() == nil {
		t.Error("expected First() to return a FieldError")
	}
}

func TestValidationErrorsForField(t *testing.T) {
	type S struct {
		Name string `json:"name" rudra:"required,min=5"`
	}
	err := validator.Validate(&S{Name: "ab"}) // fails min=5
	ve := err.(validator.ValidationErrors)
	if len(ve.ForField("name")) == 0 {
		t.Error("expected ForField('name') to return errors")
	}
}

// ---- Custom rule ---------------------------------------------------------

func TestCustomRule(t *testing.T) {
	validator.Register("startswith_a", func(field string, v reflect.Value, _ string, _ reflect.Value) string {
		if v.Kind() == reflect.String && len(v.String()) > 0 && v.String()[0] == 'A' {
			return ""
		}
		return "startswith_a"
	})

	type S struct {
		Name string `rudra:"startswith_a"`
	}
	assertInvalid(t, &S{Name: "Bravo"}, "startswith_a")
	assertValid(t, &S{Name: "Arjun"})
}

func TestCustomMessage(t *testing.T) {
	validator.RegisterMessage("required", "Le champ '{field}' est obligatoire")

	type S struct {
		Name string `rudra:"required"`
	}
	err := validator.Validate(&S{})
	if err == nil {
		t.Fatal("expected error")
	}
	ve := err.(validator.ValidationErrors)
	if ve.First() == nil || ve.First().Message == "" {
		t.Error("expected custom message")
	}
	// Restore original
	validator.RegisterMessage("required", "")
}

// ---- Struct metadata cache performance ----------------------------------

func BenchmarkValidateCached(b *testing.B) {
	type S struct {
		Name  string `rudra:"required,min=2,max=64"`
		Email string `rudra:"required,email"`
		Age   int    `rudra:"required,min=18,max=120"`
	}
	s := &S{Name: "Arjun", Email: "arjun@example.com", Age: 25}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(s)
	}
}
