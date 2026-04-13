package render

import "net/http"

// Text writes a plain text response.
func Text(w http.ResponseWriter, code int, s string) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	_, err := w.Write([]byte(s))
	return err
}