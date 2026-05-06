//go:build sonic

package binding

import (
	"errors"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
)

type jsonBinding struct{}

func (jsonBinding) Name() string { return "json" }

// Bind decodes the request body as JSON into v using Sonic SIMD acceleration.
// Sonic is 2–4× faster than encoding/json on modern AMD64/ARM64 hardware.
// Enable with: go build -tags sonic ./...
func (jsonBinding) Bind(r *http.Request, v any) error {
	if r == nil || r.Body == nil {
		return wrapErr("json", errors.New("empty body"))
	}

	// Sonic requires the full payload in memory; read with a size cap.
	data, err := io.ReadAll(io.LimitReader(r.Body, MaxBodyBytes))
	if err != nil {
		return wrapErr("json", err)
	}
	if len(data) == 0 {
		return wrapErr("json", errors.New("empty body"))
	}

	if err := sonic.ConfigFastest.Unmarshal(data, v); err != nil {
		return wrapErr("json", err)
	}
	return nil
}
