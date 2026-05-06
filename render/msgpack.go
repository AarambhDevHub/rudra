package render

import "net/http"

// Msgpack writes a MessagePack-encoded response.
// Requires build tag: -tags msgpack
// Dependency:  go get github.com/shamaton/msgpack/v2
//
// Usage:
//
//	app.GET("/data", func(c *context.Context) error {
//	    return c.Msgpack(200, myStruct)
//	})
func Msgpack(w http.ResponseWriter, code int, v any) error {
	data, err := encodeMsgpackRender(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/msgpack")
	w.WriteHeader(code)
	_, err = w.Write(data)
	return err
}
