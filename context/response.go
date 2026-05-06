package context

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/AarambhDevHub/rudra/binding"
	"github.com/AarambhDevHub/rudra/render"
	"github.com/AarambhDevHub/rudra/validator"
)

// ---- Response Renderers -------------------------------------------------

// JSON writes a JSON-encoded response with the given status code.
// Encodes directly to the ResponseWriter — zero intermediate buffer.
func (c *Context) JSON(code int, v any) error {
	return render.JSON(c.writer, code, v)
}

// String writes a plain text response.
func (c *Context) String(code int, s string) error {
	return render.Text(c.writer, code, s)
}

// HTML writes an HTML response.
func (c *Context) HTML(code int, html string) error {
	return render.HTML(c.writer, code, html)
}

// Blob writes a binary response with the specified content type.
func (c *Context) Blob(code int, contentType string, data []byte) error {
	return render.Blob(c.writer, code, contentType, data)
}

// XML writes an XML-encoded response.
func (c *Context) XML(code int, v any) error {
	return render.XML(c.writer, code, v)
}

// Stream writes chunked data incrementally via a writer callback.
func (c *Context) Stream(code int, contentType string, fn func(w io.Writer) error) error {
	return render.Stream(c.writer, code, contentType, fn)
}

// JSONP writes a JSONP response.
func (c *Context) JSONP(code int, callback string, v any) error {
	return render.JSONP(c.writer, code, callback, v)
}

// Msgpack writes a MessagePack-encoded response.
// Requires build tag: -tags msgpack
func (c *Context) Msgpack(code int, v any) error {
	return render.Msgpack(c.writer, code, v)
}

// ---- Request Body Binding (Phase 0 / legacy) ----------------------------

// BindJSON decodes the JSON request body into v.
// The body is cached so it can be re-read by subsequent calls.
func (c *Context) BindJSON(v any) error {
	if err := c.ensureBody(); err != nil {
		return jsonBodyError("failed to read body: " + err.Error())
	}
	if len(c.body) == 0 {
		return jsonBodyError("request body is empty")
	}
	if err := json.Unmarshal(c.body, v); err != nil {
		return jsonBodyError(err.Error())
	}
	return nil
}

// BindXML decodes the XML request body into v.
// The body is cached so it can be re-read by subsequent calls.
func (c *Context) BindXML(v any) error {
	if err := c.ensureBody(); err != nil {
		return xmlBodyError("failed to read body: " + err.Error())
	}
	if len(c.body) == 0 {
		return xmlBodyError("request body is empty")
	}
	if err := xml.Unmarshal(c.body, v); err != nil {
		return xmlBodyError(err.Error())
	}
	return nil
}

// ensureBody reads the request body once and caches it in c.body.
// After the first call, c.request.Body is replaced with a reader over the cache
// so subsequent binders (or re-reads) still work correctly.
func (c *Context) ensureBody() error {
	if c.body != nil {
		// Already cached — restore the body reader so binders can read it again.
		c.request.Body = io.NopCloser(bytes.NewReader(c.body))
		return nil
	}
	if c.request.Body == nil {
		c.body = []byte{}
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(c.request.Body, 32<<20))
	if err != nil {
		return err
	}
	c.body = body
	c.request.Body = io.NopCloser(bytes.NewReader(c.body))
	return nil
}

// ---- Phase 2: Auto-detect binding ----------------------------------------

// ShouldBind selects the appropriate binder based on the request's Content-Type
// and decodes the request body (or form/query data) into v.
//
// Content-Type mapping:
//   - application/json                  → JSON
//   - application/xml / text/xml        → XML
//   - application/x-www-form-urlencoded → Form
//   - multipart/form-data               → Multipart
//   - application/msgpack               → Msgpack
//   - (absent or unknown)               → JSON
//
// ShouldBind does NOT run validation. Use MustBind for bind+validate in one call.
func (c *Context) ShouldBind(v any) error {
	b := binding.For(c.request)
	// For body binders, ensure the body is pre-cached so it can be re-read.
	switch b.Name() {
	case "json", "xml", "msgpack":
		if err := c.ensureBody(); err != nil {
			return err
		}
	}
	return b.Bind(c.request, v)
}

// ShouldBindWith uses the provided binder explicitly, bypassing Content-Type detection.
//
// Example:
//
//	if err := c.ShouldBindWith(&req, binding.Query); err != nil {
//	    return err
//	}
func (c *Context) ShouldBindWith(v any, b binding.Binder) error {
	switch b.Name() {
	case "json", "xml", "msgpack":
		if err := c.ensureBody(); err != nil {
			return err
		}
	}
	return b.Bind(c.request, v)
}

// MustBind binds and validates v in a single call.
//
//   - Binding failure  → 400 Bad Request
//   - Validation failure → 422 Unprocessable Entity
//
// The handler chain is aborted on any failure.
//
// Example:
//
//	var req CreateUserRequest
//	if err := c.MustBind(&req); err != nil {
//	    return err // already aborted with appropriate status
//	}
func (c *Context) MustBind(v any) error {
	if err := c.ShouldBind(v); err != nil {
		return c.AbortWithError(http.StatusBadRequest, err)
	}
	if err := c.Validate(v); err != nil {
		return c.AbortWithError(http.StatusUnprocessableEntity, err)
	}
	return nil
}

// ---- Phase 2: Part-specific binders -------------------------------------

// BindQuery binds URL query parameters into v.
// Struct fields are matched using the `query` tag (falls back to lower-cased field name).
//
// Example:
//
//	// GET /search?page=2&limit=20&q=rudra
//	type SearchParams struct {
//	    Page  int    `query:"page"`
//	    Limit int    `query:"limit"`
//	    Q     string `query:"q"`
//	}
//	var params SearchParams
//	if err := c.BindQuery(&params); err != nil { ... }
func (c *Context) BindQuery(v any) error {
	return binding.Query.Bind(c.request, v)
}

// BindPath binds URL path parameters into v.
// Struct fields are matched using the `path` tag (falls back to lower-cased field name).
//
// Example:
//
//	// Route: /orgs/:org/repos/:repo
//	type RepoParams struct {
//	    Org  string `path:"org"`
//	    Repo string `path:"repo"`
//	}
//	var p RepoParams
//	if err := c.BindPath(&p); err != nil { ... }
func (c *Context) BindPath(v any) error {
	// Convert context.Param → binding.Param without circular import.
	cp := c.Params()
	bp := make([]binding.Param, len(cp))
	for i, p := range cp {
		bp[i] = binding.Param{Key: p.Key, Value: p.Value}
	}
	return binding.BindPath(bp, v)
}

// BindHeader binds request headers into v.
// Struct fields are matched using the `header` tag (case-insensitive).
//
// Example:
//
//	type AuthHeaders struct {
//	    Authorization string `header:"Authorization"`
//	    RequestID     string `header:"X-Request-Id"`
//	}
//	var h AuthHeaders
//	if err := c.BindHeader(&h); err != nil { ... }
func (c *Context) BindHeader(v any) error {
	return binding.Header.Bind(c.request, v)
}

// ---- Phase 2: Validation -------------------------------------------------

// Validate runs rudra struct tag validation against v.
// Returns nil if all rules pass, or a validator.ValidationErrors slice on failure.
//
// Use MustBind to bind + validate + abort in one call.
// Use Validate alone when you need to inspect errors before responding.
//
// Example:
//
//	if err := c.Validate(&req); err != nil {
//	    ve := err.(validator.ValidationErrors)
//	    return c.JSON(422, map[string]any{
//	        "errors": ve.Map(),
//	    })
//	}
func (c *Context) Validate(v any) error {
	return validator.Validate(v)
}

// ---- Multipart helpers --------------------------------------------------

// FormFile returns the multipart form file for the given field name.
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.request.MultipartForm == nil {
		if err := c.request.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
	}
	_, fh, err := c.request.FormFile(name)
	return fh, err
}

// FormValue returns the form value for the given key.
func (c *Context) FormValue(name string) string {
	return c.request.FormValue(name)
}

// MultipartForm returns the parsed multipart form.
func (c *Context) MultipartForm() (*multipart.Form, error) {
	if c.request.MultipartForm == nil {
		if err := c.request.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
	}
	return c.request.MultipartForm, nil
}

// ---- Error types --------------------------------------------------------

type jsonBodyError string

func (e jsonBodyError) Error() string { return "json binding: " + string(e) }

type xmlBodyError string

func (e xmlBodyError) Error() string { return "xml binding: " + string(e) }
