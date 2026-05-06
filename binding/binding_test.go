package binding_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/AarambhDevHub/rudra/binding"
)

// ---- Test helpers --------------------------------------------------------

func newRequest(method, path, ct string, body []byte) *http.Request {
	r := httptest.NewRequest(method, path, bytes.NewReader(body))
	if ct != "" {
		r.Header.Set("Content-Type", ct)
	}
	return r
}

// ---- For() auto-selector -------------------------------------------------

func TestForJSON(t *testing.T) {
	r := newRequest("POST", "/", "application/json", nil)
	if binding.For(r).Name() != "json" {
		t.Error("expected json binder for application/json")
	}
}

func TestForXML(t *testing.T) {
	r := newRequest("POST", "/", "application/xml", nil)
	if binding.For(r).Name() != "xml" {
		t.Error("expected xml binder for application/xml")
	}
}

func TestForForm(t *testing.T) {
	r := newRequest("POST", "/", "application/x-www-form-urlencoded", nil)
	if binding.For(r).Name() != "form" {
		t.Error("expected form binder")
	}
}

func TestForFallbackJSON(t *testing.T) {
	r := newRequest("POST", "/", "", nil)
	if binding.For(r).Name() != "json" {
		t.Error("expected json fallback for empty content-type")
	}
}

func TestForWithCharset(t *testing.T) {
	r := newRequest("POST", "/", "application/json; charset=utf-8", nil)
	if binding.For(r).Name() != "json" {
		t.Error("expected json binder when charset is appended")
	}
}

// ---- JSON binder ---------------------------------------------------------

type jsonUser struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Admin bool   `json:"admin"`
}

func TestJSONBind(t *testing.T) {
	payload, _ := json.Marshal(jsonUser{Name: "Arjun", Age: 25, Admin: true})
	r := newRequest("POST", "/", "application/json", payload)

	var u jsonUser
	if err := binding.JSON.Bind(r, &u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Name != "Arjun" || u.Age != 25 || !u.Admin {
		t.Errorf("unexpected values: %+v", u)
	}
}

func TestJSONBindEmptyBody(t *testing.T) {
	r := newRequest("POST", "/", "application/json", nil)
	var u jsonUser
	if err := binding.JSON.Bind(r, &u); err == nil {
		t.Error("expected error for empty body")
	}
}

func TestJSONBindNilBody(t *testing.T) {
	r := &http.Request{}
	var u jsonUser
	if err := binding.JSON.Bind(r, &u); err == nil {
		t.Error("expected error for nil body")
	}
}

func TestJSONBindMalformed(t *testing.T) {
	r := newRequest("POST", "/", "application/json", []byte(`{bad json`))
	var u jsonUser
	if err := binding.JSON.Bind(r, &u); err == nil {
		t.Error("expected error for malformed JSON")
	}
}

// ---- XML binder ----------------------------------------------------------

type xmlItem struct {
	XMLName xml.Name `xml:"item"`
	Name    string   `xml:"name"`
	Value   int      `xml:"value"`
}

func TestXMLBind(t *testing.T) {
	payload := []byte(`<item><name>sword</name><value>42</value></item>`)
	r := newRequest("POST", "/", "application/xml", payload)

	var item xmlItem
	if err := binding.XML.Bind(r, &item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "sword" || item.Value != 42 {
		t.Errorf("unexpected values: %+v", item)
	}
}

// ---- Form binder ---------------------------------------------------------

type formData struct {
	Username string   `form:"username"`
	Age      int      `form:"age"`
	Tags     []string `form:"tags"`
}

func TestFormBind(t *testing.T) {
	vals := url.Values{}
	vals.Set("username", "priya")
	vals.Set("age", "24")
	vals.Add("tags", "go")
	vals.Add("tags", "rust")

	r := newRequest("POST", "/", "application/x-www-form-urlencoded",
		[]byte(vals.Encode()))

	var fd formData
	if err := binding.Form.Bind(r, &fd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fd.Username != "priya" {
		t.Errorf("expected username=priya, got %s", fd.Username)
	}
	if fd.Age != 24 {
		t.Errorf("expected age=24, got %d", fd.Age)
	}
	if len(fd.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(fd.Tags))
	}
}

func TestFormBindCommaSplit(t *testing.T) {
	vals := url.Values{}
	vals.Set("tags", "go,rust,wasm")

	r := newRequest("POST", "/", "application/x-www-form-urlencoded",
		[]byte(vals.Encode()))

	var fd formData
	if err := binding.Form.Bind(r, &fd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fd.Tags) != 3 {
		t.Errorf("expected 3 tags from comma-split, got %d: %v", len(fd.Tags), fd.Tags)
	}
}

// ---- Multipart binder ----------------------------------------------------

func TestMultipartBind(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("username", "arjun")
	_ = mw.WriteField("age", "28")
	mw.Close()

	r := newRequest("POST", "/", mw.FormDataContentType(), buf.Bytes())

	type mpData struct {
		Username string `form:"username"`
		Age      int    `form:"age"`
	}
	var d mpData
	if err := binding.Multipart.Bind(r, &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Username != "arjun" || d.Age != 28 {
		t.Errorf("unexpected values: %+v", d)
	}
}

// ---- Query binder --------------------------------------------------------

type queryParams struct {
	Page  int    `query:"page"`
	Limit int    `query:"limit"`
	Q     string `query:"q"`
}

func TestQueryBind(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?page=3&limit=20&q=rudra", nil)

	var p queryParams
	if err := binding.Query.Bind(r, &p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Page != 3 || p.Limit != 20 || p.Q != "rudra" {
		t.Errorf("unexpected values: %+v", p)
	}
}

func TestQueryBindMissingField(t *testing.T) {
	r := httptest.NewRequest("GET", "/?page=1", nil)

	var p queryParams
	if err := binding.Query.Bind(r, &p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Page != 1 || p.Limit != 0 || p.Q != "" {
		t.Errorf("unexpected values: %+v", p)
	}
}

// ---- Path binder ---------------------------------------------------------

type pathParams struct {
	OrgID int64  `path:"org"`
	Repo  string `path:"repo"`
}

func TestPathBind(t *testing.T) {
	params := []binding.Param{
		{Key: "org", Value: "42"},
		{Key: "repo", Value: "rudra"},
	}

	var p pathParams
	if err := binding.BindPath(params, &p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.OrgID != 42 || p.Repo != "rudra" {
		t.Errorf("unexpected values: %+v", p)
	}
}

func TestPathBindIntCoercion(t *testing.T) {
	params := []binding.Param{{Key: "org", Value: "not-a-number"}}
	var p pathParams
	if err := binding.BindPath(params, &p); err == nil {
		t.Error("expected error for non-numeric path param bound to int64")
	}
}

// ---- Header binder -------------------------------------------------------

type reqHeaders struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-Id"`
}

func TestHeaderBind(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer token123")
	r.Header.Set("X-Request-Id", "uuid-abc")

	var h reqHeaders
	if err := binding.Header.Bind(r, &h); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.Authorization != "Bearer token123" {
		t.Errorf("expected bearer token, got %s", h.Authorization)
	}
	if h.RequestID != "uuid-abc" {
		t.Errorf("expected request ID, got %s", h.RequestID)
	}
}

func TestHeaderBindCaseInsensitive(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("authorization", "Bearer xyz") // lowercase key

	type lowerHeader struct {
		Auth string `header:"authorization"` // lowercase tag
	}
	var h lowerHeader
	if err := binding.Header.Bind(r, &h); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(h.Auth, "Bearer") {
		t.Errorf("expected bearer token, got %s", h.Auth)
	}
}

// ---- Decoder type coercion -----------------------------------------------

func TestDecoderBoolCoercion(t *testing.T) {
	r := httptest.NewRequest("GET", "/?active=true&flag=1&disabled=false", nil)

	type flags struct {
		Active   bool `query:"active"`
		Flag     bool `query:"flag"`
		Disabled bool `query:"disabled"`
	}
	var f flags
	if err := binding.Query.Bind(r, &f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.Active || !f.Flag || f.Disabled {
		t.Errorf("unexpected bool values: %+v", f)
	}
}

func TestDecoderSliceQuery(t *testing.T) {
	r := httptest.NewRequest("GET", "/?ids=1&ids=2&ids=3", nil)

	type listReq struct {
		IDs []int `query:"ids"`
	}
	var lr listReq
	if err := binding.Query.Bind(r, &lr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lr.IDs) != 3 || lr.IDs[0] != 1 || lr.IDs[2] != 3 {
		t.Errorf("unexpected ids: %v", lr.IDs)
	}
}

func TestDecoderPointerField(t *testing.T) {
	r := httptest.NewRequest("GET", "/?limit=50", nil)

	type paged struct {
		Limit *int `query:"limit"`
	}
	var p paged
	if err := binding.Query.Bind(r, &p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Limit == nil || *p.Limit != 50 {
		t.Errorf("expected *int=50, got %v", p.Limit)
	}
}

func TestDecoderEmbeddedStruct(t *testing.T) {
	type base struct {
		Page int `query:"page"`
	}
	type embeddedReq struct {
		base
		Q string `query:"q"`
	}

	r := httptest.NewRequest("GET", "/?page=2&q=hello", nil)
	var req embeddedReq
	if err := binding.Query.Bind(r, &req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Page != 2 || req.Q != "hello" {
		t.Errorf("unexpected embedded values: %+v", req)
	}
}
