package tritonhttp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	// Addr specifies the TCP address for the server to listen on,
	// in the form "host:port". It shall be passed to net.Listen()
	// during ListenAndServe().
	Addr string // e.g. ":0"

	// DocRoot specifies the path to the directory to serve static files from.
	DocRoot string
}

// check the server setup
func (s *Server) validateServerSetup() error {
	fi, err := os.Stat(s.DocRoot)
	if os.IsNotExist(err) {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("doc root %q is not a directory", s.DocRoot)
	}
	return nil
}

// ListenAndServe listens on the TCP network address s.Addr and then
// handles requests on incoming connections.
func (s *Server) ListenAndServe() error {
	// Validate the Server Configuration
	if err := s.validateServerSetup(); err != nil {
		return err
	}
	// Listen on a port
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	fmt.Println("Listening on", ln.Addr())

	// Accept connections and server them
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error in accepting connection", err)
			continue
		}
		fmt.Println("accepted connection from ", conn.RemoteAddr())
		go s.HandleConnection(conn)
	}
}

// HandleConnection reads requests from the accepted conn and handles them.
func (s *Server) HandleConnection(conn net.Conn) {
	br := bufio.NewReader(conn)

	for {
		// set a read timeout
		if err := conn.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
			fmt.Printf("Failed to set timeout for the connection %v", conn.RemoteAddr())
			_ = conn.Close()
			return
		}

		// Read the next request
		req, bytesRecieved, err := ReadRequest(br)

		// Handle errors
		// 1. client has closed the conn ==> io.EOF error
		if errors.Is(err, io.EOF) {
			fmt.Printf("Connection closed by the client %v", conn.RemoteAddr())
			if bytesRecieved {
				log.Printf("Handling Bad Request for error : %v", err)
				res := &Response{}
				res.HandleBadRequest()
				_ = res.Write(conn)
				_ = conn.Close()
				return
			}
			_ = conn.Close()
			return
		}
		// 2. timeout from the server --> net.Error
		if err, ok := err.(net.Error); ok && err.Timeout() {
			if bytesRecieved {
				log.Printf("Handling Bad Request for error : %v", err)
				res := &Response{}
				res.HandleBadRequest()
				_ = res.Write(conn)
				_ = conn.Close()
				return
			}
			log.Printf("Connection to %v timed out.", conn.RemoteAddr())
			_ = conn.Close()
			return
		}
		// 3. malformed/invalid requests -->
		if err != nil || req.URL[0:1] != "/" {
			log.Printf("Handling Bad Request for error : %v", err)
			res := &Response{}
			res.HandleBadRequest()
			_ = res.Write(conn)
			_ = conn.Close()
			return
		}
		// Handle the happy path
		fmt.Printf("Handle good request : %v", req)
		res := s.HandleGoodRequest(req)

		if err := res.Write(conn); err != nil {
			fmt.Println(err)
		}
	}
}

// HandleGoodRequest handles the valid req and generates the corresponding res.
func (s *Server) HandleGoodRequest(req *Request) (res *Response) {
	// Hint: use the other methods below
	res = &Response{}
	var filePath string
	if req.URL[len(req.URL)-1] == '/' {
		req.URL += "index.html"
	}
	filePath = path.Join(s.DocRoot, req.URL)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			res.HandleNotFound(req)
		} else {
			log.Fatal(err)
		}
	} else {
		if fileInfo.IsDir() {
			res.HandleNotFound(req)
		} else {
			res.HandleOK(req, filePath)
		}
	}
	res.Proto = req.Proto
	return res
}

// HandleOK prepares res to be a 200 OK response
// ready to be written back to client.
func (res *Response) HandleOK(req *Request, path string) {
	res.StatusCode = 200
	res.FilePath = path
	header := make(map[string]string)
	header["Date"] = FormatTime(time.Now())
	info, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	header["Last-Modified"] = FormatTime(info.ModTime())
	ext := strings.Split(path, ".")
	header["Content-Type"] = MIMETypeByExtension("." + ext[len(ext)-1])
	header["Content-Length"] = strconv.FormatInt(info.Size(), 10)
	if req.Close {
		header["Connection"] = "close"
	}
	res.Header = header
}

// HandleBadRequest prepares res to be a 400 Bad Request response
// ready to be written back to client.
func (res *Response) HandleBadRequest() {
	res.Proto = "HTTP/1.1"
	res.StatusCode = 400
	header := make(map[string]string)
	header["Date"] = FormatTime(time.Now())
	header["Connection"] = "close"
	res.Header = header
}

// HandleNotFound prepares res to be a 404 Not Found response
// ready to be written back to client.
func (res *Response) HandleNotFound(req *Request) {
	res.StatusCode = 404
	header := make(map[string]string)
	header["Date"] = FormatTime(time.Now())
	if req.Close {
		header["Connection"] = "close"
	}
	res.Header = header
}
