package binding

import (
	"io"
	"net/http"
)

type msgpackBinding struct{}

func (msgpackBinding) Name() string { return "msgpack" }

// Bind decodes a MessagePack request body into v.
// Requires build tag: -tags msgpack
// Dependency: go get github.com/shamaton/msgpack/v2
func (msgpackBinding) Bind(r *http.Request, v any) error {
	if r == nil || r.Body == nil {
		return wrapErr("msgpack", errMsgpackEmptyBody)
	}

	data, err := io.ReadAll(io.LimitReader(r.Body, MaxBodyBytes))
	if err != nil {
		return wrapErr("msgpack", err)
	}

	return wrapErr("msgpack", decodeMsgpack(data, v))
}

// EncodeMsgpack marshals v to MessagePack bytes.
// Requires build tag: -tags msgpack
func EncodeMsgpack(v any) ([]byte, error) {
	return encodeMsgpack(v)
}
