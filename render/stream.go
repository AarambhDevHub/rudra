package render

import (
	"io"
	"net/http"
)

// Stream writes chunked data incrementally.
func Stream(w http.ResponseWriter, code int, contentType string, fn func(w io.Writer) error) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return StreamNotSupported()
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(code)

	fw := &flushWriter{w: w, flusher: flusher}
	return fn(fw)
}

type flushWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func (fw *flushWriter) Write(p []byte) (int, error) {
	n, err := fw.w.Write(p)
	fw.flusher.Flush()
	return n, err
}

// StreamNotSupported returns an error indicating streaming is not supported.
func StreamNotSupported() error {
	return streamError("streaming not supported")
}

type streamError string

func (e streamError) Error() string { return string(e) }