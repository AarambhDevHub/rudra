// Package binding provides request data binding for the Rudra web framework.
//
// All binders are stateless singleton values safe for concurrent use.
// The For() function selects the appropriate binder from the request's
// Content-Type header.
//
// Build tags:
//   - sonic   → enables Sonic SIMD-accelerated JSON (bytedance/sonic)
//   - msgpack → enables MessagePack binding/rendering (shamaton/msgpack)
package binding

import (
	"mime"
	"net/http"
)

// MIME type constants used for Content-Type negotiation.
const (
	MIMEJSON          = "application/json"
	MIMEXML           = "application/xml"
	MIMEXMLText       = "text/xml"
	MIMEForm          = "application/x-www-form-urlencoded"
	MIMEMultipartForm = "multipart/form-data"
	MIMEMsgpack       = "application/msgpack"
	MIMEMsgpackAlt    = "application/x-msgpack"
)

// MaxBodyBytes is the default maximum request body size (32 MB).
// Binders use io.LimitReader to enforce this limit.
const MaxBodyBytes int64 = 32 << 20

// Binder is the interface all request data binders implement.
// Implementations must be safe for concurrent use without external locking.
type Binder interface {
	// Name returns the binding format name (e.g., "json", "form").
	Name() string
	// Bind reads from r and populates v. v must be a non-nil pointer.
	Bind(r *http.Request, v any) error
}

// Singleton binders — stateless, allocated once at package init.
var (
	JSON      Binder = jsonBinding{}
	XML       Binder = xmlBinding{}
	Form      Binder = formBinding{}
	Multipart Binder = multipartBinding{MaxMemory: MaxBodyBytes}
	Query     Binder = queryBinding{}
	Header    Binder = headerBinding{}
	Msgpack   Binder = msgpackBinding{}
)

// For selects the appropriate body binder based on the request's Content-Type.
// Falls back to JSON for absent or unrecognised content types.
//
// Mapping:
//
//	application/json                   → JSON
//	application/xml / text/xml         → XML
//	application/x-www-form-urlencoded  → Form
//	multipart/form-data                → Multipart
//	application/msgpack                → Msgpack
//	<anything else>                    → JSON
func For(r *http.Request) Binder {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return JSON
	}

	// Strip optional parameters (e.g., "; charset=utf-8").
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return JSON
	}

	switch mt {
	case MIMEJSON:
		return JSON
	case MIMEXML, MIMEXMLText:
		return XML
	case MIMEForm:
		return Form
	case MIMEMultipartForm:
		return Multipart
	case MIMEMsgpack, MIMEMsgpackAlt:
		return Msgpack
	default:
		return JSON
	}
}
