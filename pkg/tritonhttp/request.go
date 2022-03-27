package tritonhttp

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

type Request struct {
	Method string // e.g. "GET"
	URL    string // e.g. "/path/to/a/file"
	Proto  string // e.g. "HTTP/1.1"

	// Header stores misc headers excluding "Host" and "Connection",
	// which are stored in special fields below.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	Host  string // determine from the "Host" header
	Close bool   // determine from the "Connection" header
}

// ReadRequest tries to read the next valid request from br.
//
// If it succeeds, it returns the valid request read. In this case,
// bytesReceived should be true, and err should be nil.
//
// If an error occurs during the reading, it returns the error,
// and a nil request. In this case, bytesReceived indicates whether or not
// some bytes are received before the error occurs. This is useful to determine
// the timeout with partial request received condition.
func ReadRequest(br *bufio.Reader) (req *Request, bytesReceived bool, err error) {
	req = &Request{}
	// read first line
	line, err := ReadLine(br)
	if len(line) == 0 {
		return req, false, io.EOF
	}
	if err != nil {
		return nil, true, err
	}
	requestHeaders := strings.Split(line, " ")
	if len(requestHeaders) != 3 {
		return nil, true, errors.New("request header should contain 3 parts")
	}
	if requestHeaders[0] != "GET" {
		return nil, true, errors.New("should be GET method")
	}
	if !strings.HasPrefix(requestHeaders[1], "/") {
		return nil, true, errors.New("URL should start with /")
	}
	if requestHeaders[2] != "HTTP/1.1" {
		return nil, true, errors.New("proto should be HTTP/1.1")
	}

	req.Method = requestHeaders[0]
	req.URL = requestHeaders[1]
	req.Proto = requestHeaders[2]

	if req.Header == nil {
		req.Header = make(map[string]string)
	}
	//req.Close = false
	// Check required headers
	for {
		line, err := ReadLine(br)
		if err != nil {
			return nil, true, err
		}
		if line == "" {
			break
		}
		headerline := strings.Split(line, ":")
		if len(headerline) != 2 {
			return nil, true, errors.New("malformed header")
		}
		key := headerline[0]
		val := headerline[1]

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		if key == "Host" {
			req.Host = val
		} else if key == "Connection" {
			if val == "close" {
				req.Close = true
			} else {
				req.Close = false
			}
		} else {
			req.Header[CanonicalHeaderKey(key)] = val
		}
	}
	// Handle special headers
	if len(req.Host) == 0 {
		return nil, true, errors.New("there is no host")
	}

	return req, true, nil
}
