package render

import (
	"encoding/xml"
	"net/http"
)

// XML writes an XML response.
func XML(w http.ResponseWriter, code int, v any) error {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(code)
	output, err := xml.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(append([]byte(xml.Header), output...))
	return err
}