package render

import "net/http"

// Render writes a response with the given content type and body.
func Render(w http.ResponseWriter, code int, contentType string, body []byte) error {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	_, err := w.Write(body)
	return err
}