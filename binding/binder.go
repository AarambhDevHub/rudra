package binding

// Binder is the interface for all request data binders.
type Binder interface {
	Bind(v any) error
}