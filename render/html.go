package render

import "net/http"

// HTML writes an HTML response.
func HTML(w http.ResponseWriter, code int, html string) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_, err := w.Write([]byte(html))
	return err
}