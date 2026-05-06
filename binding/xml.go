package binding

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
)

type xmlBinding struct{}

func (xmlBinding) Name() string { return "xml" }

// Bind decodes the request body as XML into v.
func (xmlBinding) Bind(r *http.Request, v any) error {
	if r == nil || r.Body == nil {
		return wrapErr("xml", errors.New("empty body"))
	}

	if err := xml.NewDecoder(io.LimitReader(r.Body, MaxBodyBytes)).Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return wrapErr("xml", errors.New("empty body"))
		}
		return wrapErr("xml", err)
	}
	return nil
}
