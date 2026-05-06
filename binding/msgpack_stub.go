//go:build !msgpack

package binding

import "errors"

var errMsgpackEmptyBody = errors.New("empty body")

// decodeMsgpack is a no-op stub. Enable with: go build -tags msgpack
func decodeMsgpack(_ []byte, _ any) error {
	return errors.New("msgpack: add -tags msgpack and go get github.com/shamaton/msgpack/v2")
}

func encodeMsgpack(_ any) ([]byte, error) {
	return nil, errors.New("msgpack: add -tags msgpack and go get github.com/shamaton/msgpack/v2")
}
