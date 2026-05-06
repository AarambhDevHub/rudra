//go:build !msgpack

package render

import "errors"

func encodeMsgpackRender(_ any) ([]byte, error) {
	return nil, errors.New("msgpack: add -tags msgpack and go get github.com/shamaton/msgpack/v2")
}
