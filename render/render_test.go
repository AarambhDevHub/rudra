package render_test

import (
	"encoding/json"
	"encoding/xml"
	"net/http/httptest"
	"testing"

	"github.com/AarambhDevHub/rudra/render"
)

func TestRenderJSON(t *testing.T) {
	w := httptest.NewRecorder()
	err := render.JSON(w, 200, map[string]string{"key": "value"})
	if err != nil {
		t.Fatal(err)
	}
	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("expected JSON content type, got %s", w.Header().Get("Content-Type"))
	}
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["key"] != "value" {
		t.Errorf("expected value, got %s", resp["key"])
	}
}

func TestRenderText(t *testing.T) {
	w := httptest.NewRecorder()
	err := render.Text(w, 200, "hello world")
	if err != nil {
		t.Fatal(err)
	}
	if w.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("expected text content type, got %s", w.Header().Get("Content-Type"))
	}
	if w.Body.String() != "hello world" {
		t.Errorf("expected hello world, got %s", w.Body.String())
	}
}

func TestRenderHTML(t *testing.T) {
	w := httptest.NewRecorder()
	err := render.HTML(w, 200, "<h1>Hello</h1>")
	if err != nil {
		t.Fatal(err)
	}
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("expected HTML content type")
	}
}

func TestRenderBlob(t *testing.T) {
	w := httptest.NewRecorder()
	err := render.Blob(w, 200, "image/png", []byte{0x89, 0x50, 0x4E})
	if err != nil {
		t.Fatal(err)
	}
	if w.Header().Get("Content-Type") != "image/png" {
		t.Errorf("expected image/png content type")
	}
}

func TestRenderXML(t *testing.T) {
	type Item struct {
		XMLName xml.Name `xml:"item"`
		Name    string   `xml:"name"`
	}
	w := httptest.NewRecorder()
	err := render.XML(w, 200, Item{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if w.Header().Get("Content-Type") != "application/xml; charset=utf-8" {
		t.Errorf("expected XML content type")
	}
}

func BenchmarkRenderJSON(b *testing.B) {
	data := map[string]string{"key": "value", "framework": "rudra"}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		render.JSON(w, 200, data)
	}
}