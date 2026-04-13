package render

import "net/http"

// Blob writes a binary response with the given content type.
func Blob(w http.ResponseWriter, code int, contentType string, data []byte) error {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	_, err := w.Write(data)
	return err
}