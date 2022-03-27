package tritonhttp

import (
	"errors"
	"io"
	"os"
	"sort"
	"strconv"
)

type Response struct {
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.1"

	// Header stores all headers to write to the response.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	// Request is the valid request that leads to this response.
	// It could be nil for responses not resulting from a valid request.
	Request *Request

	// FilePath is the local path to the file to serve.
	// It could be "", which means there is no file to serve.
	FilePath string
}

// Write writes the res to the w.
func (res *Response) Write(w io.Writer) error {
	if err := res.WriteStatusLine(w); err != nil {
		return err
	}
	if err := res.WriteSortedHeaders(w); err != nil {
		return err
	}
	if err := res.WriteBody(w); err != nil {
		return err
	}
	return nil
}

// WriteStatusLine writes the status line of res to w, including the ending "\r\n".
// For example, it could write "HTTP/1.1 200 OK\r\n".
func (res *Response) WriteStatusLine(w io.Writer) error {
	if res.StatusCode == 200 {
		_, err := w.Write([]byte(res.Proto + " " + strconv.Itoa(res.StatusCode) + " OK\r\n"))
		if err != nil {
			return errors.New("200 write failed")
		}
	} else if res.StatusCode == 400 {
		_, err := w.Write([]byte(res.Proto + " " + strconv.Itoa(res.StatusCode) + " Bad Request\r\n"))
		if err != nil {
			return errors.New("400 write failed")
		}
	} else if res.StatusCode == 404 {
		_, err := w.Write([]byte(res.Proto + " " + strconv.Itoa(res.StatusCode) + " Not Found\r\n"))
		if err != nil {
			return errors.New("404 write failed")
		}
	} else {
		return errors.New("not the code")
	}
	return nil
}

// WriteSortedHeaders writes the headers of res to w, including the ending "\r\n".
// For example, it could write "Connection: close\r\nDate: foobar\r\n\r\n".
// For HTTP, there is no need to write headers in any particular order.
// TritonHTTP requires to write in sorted order for the ease of testing.
func (res *Response) WriteSortedHeaders(w io.Writer) error {
	keys := make([]string, 0, len(res.Header))
	for k := range res.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		_, err := w.Write([]byte(k + ": " + res.Header[k] + "\r\n"))
		if err != nil {
			return errors.New("failing wrtie sorted header")
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return errors.New("failing write header end")
	}
	return nil
}

// WriteBody writes res' file content as the response body to w.
// It doesn't write anything if there is no file to serve.
func (res *Response) WriteBody(w io.Writer) error {
	if res.FilePath == "" {
		return nil
	} else {
		file, err := os.ReadFile(res.FilePath)
		if err != nil {
			return errors.New("failing to open file")
		}
		_, errB := w.Write([]byte(file))
		if errB != nil {
			return errors.New("failing to write file")
		}
	}
	return nil
}
