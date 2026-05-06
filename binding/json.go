//go:build !sonic

package binding

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type jsonBinding struct{}

func (jsonBinding) Name() string { return "json" }

// Bind decodes the request body as JSON into v.
// Uses encoding/json with a streaming decoder (zero intermediate buffer).
// Build with -tags sonic to enable Sonic SIMD-accelerated decoding instead.
func (jsonBinding) Bind(r *http.Request, v any) error {
	if r == nil || r.Body == nil {
		return wrapErr("json", errors.New("empty body"))
	}

	dec := json.NewDecoder(io.LimitReader(r.Body, MaxBodyBytes))
	if err := dec.Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return wrapErr("json", errors.New("empty body"))
		}
		return wrapErr("json", err)
	}
	return nil
}
