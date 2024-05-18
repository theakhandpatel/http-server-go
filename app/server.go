package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	// Uncomment this block to pass the first stage
	// "net"
	// "os"
)

// GET                          // HTTP method
// /index.html                  // Request target
// HTTP/1.1                     // HTTP version
// \r\n                         // CRLF that marks the end of the request line

// // Headers
// Host: localhost:4221\r\n     // Header that specifies the server's host and port
// User-Agent: curl/7.64.1\r\n  // Header that describes the client's user agent
// Accept: */*\r\n              // Header that specifies which media types the client can accept
// \r\n                         // CRLF that marks the end of the headers

// Request body (empty)
type Request struct {
	Method  string
	Target  string
	Version string
	Headers string
	Body    string
}

func NewRequest(bytes *[]byte) Request {
	reqValues := strings.SplitN(string(*bytes), "\r\n", 2)
	reqLine := strings.Split(reqValues[0], " ")
	rest := strings.SplitN(reqValues[1], "\r\n\r\n", 2)

	req := Request{
		Method:  reqLine[0],
		Target:  reqLine[1],
		Version: reqLine[2],
		Headers: rest[0],
		Body:    rest[1],
	}
	return req
}

func (r *Request) String() string {
	return fmt.Sprintf(
		"%s %s %s\r\n%s\r\n\r\n%s",
		r.Method, r.Target, r.Version, r.Headers, r.Body,
	)
}

type Response struct {
	Version    string
	StatusCode int
	Message    string
	Headers    string
	Body       string
}

func NewResponse(code int, body string) Response {
	res := Response{
		Version: "HTTP/1.1",
		Headers: "",
	}
	if code == 200 {
		res.StatusCode = code
		res.Message = "OK"
	} else if code == 404 {
		res.StatusCode = 404
		res.Message = "Not Found"
	} else {
		panic(fmt.Errorf("not a valid response code"))
	}
	res.Body = body
	return res
}

func (r *Response) String() string {
	return fmt.Sprintf(
		"%s %d %s\r\n%s\r\n%s",
		r.Version, r.StatusCode, r.Message, r.Headers, r.Body,
	)
}

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	buffer := make([]byte, 1024)

	_, err = conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		os.Exit(1)
	}
	req := NewRequest(&buffer)
	fmt.Println("req", req)

	var res Response
	if req.Target == "/" {
		res = NewResponse(200, "")
	} else {
		res = NewResponse(404, "")
	}

	_, err = conn.Write([]byte(res.String()))
	if err != nil {
		fmt.Println("Error writing response ", err.Error())
	}

}
