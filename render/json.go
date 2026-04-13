package render

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response. Encodes directly to ResponseWriter — zero intermediate buffer.
func JSON(w http.ResponseWriter, code int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

// JSONBytes writes a pre-encoded JSON response from raw bytes.
func JSONBytes(w http.ResponseWriter, code int, data []byte) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_, err := w.Write(data)
	return err
}

// JSONP writes a JSONP response.
func JSONP(w http.ResponseWriter, code int, callback string, v any) error {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.WriteHeader(code)
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(append(append([]byte(callback+"("), data...), []byte(");")...))
	return err
}