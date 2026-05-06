//go:build msgpack

package binding

import (
	"errors"

	"github.com/shamaton/msgpack/v2"
)

var errMsgpackEmptyBody = errors.New("empty body")

func decodeMsgpack(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}

func encodeMsgpack(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}
