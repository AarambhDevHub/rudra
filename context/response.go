package context

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"

	"github.com/AarambhDevHub/rudra/render"
)

// JSON writes a JSON response with the given status code.
func (c *Context) JSON(code int, v any) error {
	return render.JSON(c.writer, code, v)
}

// String writes a plain text response.
func (c *Context) String(code int, s string) error {
	return render.Text(c.writer, code, s)
}

// HTML writes an HTML response.
func (c *Context) HTML(code int, html string) error {
	return render.HTML(c.writer, code, html)
}

// Blob writes a binary response.
func (c *Context) Blob(code int, contentType string, data []byte) error {
	return render.Blob(c.writer, code, contentType, data)
}

// XML writes an XML response.
func (c *Context) XML(code int, v any) error {
	return render.XML(c.writer, code, v)
}

// Stream writes chunked data incrementally.
func (c *Context) Stream(code int, contentType string, fn func(w io.Writer) error) error {
	return render.Stream(c.writer, code, contentType, fn)
}

// JSONP writes a JSONP response.
func (c *Context) JSONP(code int, callback string, v any) error {
	return render.JSONP(c.writer, code, callback, v)
}

// BindJSON decodes the request body as JSON into v.
func (c *Context) BindJSON(v any) error {
	if c.request.Body == nil {
		return jsonBodyError("request body is empty")
	}
	body, err := io.ReadAll(io.LimitReader(c.request.Body, 32<<20))
	if err != nil {
		return jsonBodyError("failed to read body: " + err.Error())
	}
	if len(body) == 0 {
		return jsonBodyError("request body is empty")
	}
	if err := json.Unmarshal(body, v); err != nil {
		return jsonBodyError(err.Error())
	}
	c.body = body
	return nil
}

// BindXML decodes the request body as XML into v.
func (c *Context) BindXML(v any) error {
	if c.request.Body == nil {
		return xmlBodyError("request body is empty")
	}
	body, err := io.ReadAll(io.LimitReader(c.request.Body, 32<<20))
	if err != nil {
		return xmlBodyError("failed to read body: " + err.Error())
	}
	if len(body) == 0 {
		return xmlBodyError("request body is empty")
	}
	if err := xml.Unmarshal(body, v); err != nil {
		return xmlBodyError(err.Error())
	}
	c.body = body
	return nil
}

type jsonBodyError string

func (e jsonBodyError) Error() string { return "json binding: " + string(e) }

type xmlBodyError string

func (e xmlBodyError) Error() string { return "xml binding: " + string(e) }

// FormFile returns the multipart form file for the given key.
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.request.MultipartForm == nil {
		if err := c.request.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
	}
	_, fh, err := c.request.FormFile(name)
	return fh, err
}

// FormValue returns the form value for the given key.
func (c *Context) FormValue(name string) string {
	return c.request.FormValue(name)
}

// MultipartForm returns the parsed multipart form.
func (c *Context) MultipartForm() (*multipart.Form, error) {
	if c.request.MultipartForm == nil {
		if err := c.request.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
	}
	return c.request.MultipartForm, nil
}