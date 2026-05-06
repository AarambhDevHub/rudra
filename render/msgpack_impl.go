//go:build msgpack

package render

import "github.com/shamaton/msgpack/v2"

func encodeMsgpackRender(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}
