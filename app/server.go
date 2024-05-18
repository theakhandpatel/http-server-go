package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const CRLF string = "\r\n"

var dir string

type Request struct {
	Method  string
	Target  string
	Version string
	Headers map[string]string
	Body    string
}

func NewRequest(bytes *[]byte) Request {
	reqValues := strings.SplitN(string(*bytes), CRLF, 2)
	reqLine := strings.Split(reqValues[0], " ")
	rest := strings.SplitN(reqValues[1], CRLF+CRLF, 2)

	headers := make(map[string]string)
	for _, line := range strings.Split(rest[0], CRLF) {
		parts := strings.SplitN(line, ": ", 2)
		headers[strings.ToLower(parts[0])] = parts[1]
	}

	cl, ok := headers["content-length"]
	var body string
	if !ok {
		fmt.Println("content-length not set")
		body = rest[1]
	} else {
		contentLength, _ := strconv.Atoi(cl)
		body = rest[1][:contentLength]
	}
	req := Request{
		Method:  reqLine[0],
		Target:  reqLine[1],
		Version: reqLine[2],
		Headers: headers,
		Body:    body,
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
	Headers    map[string]string
	Body       string
}

func NewResponse(code int, body string, headers map[string]string) Response {
	res := Response{
		Version: "HTTP/1.1",
		Headers: make(map[string]string),
	}
	if code == 200 {
		res.StatusCode = code
		res.Message = "OK"
	} else if code == 201 {
		res.StatusCode = code
		res.Message = "Created"
	} else if code == 404 {
		res.StatusCode = 404
		res.Message = "Not Found"
	} else {
		res.StatusCode = 500
		res.Message = "Internal Server Error"
	}
	res.Body = body

	if headers == nil {
		res.Headers["Content-Type"] = "text/plain"
		res.Headers["Content-Length"] = strconv.Itoa(len(body))
	} else {
		if _, ok := headers["Content-Type"]; !ok {
			res.Headers["Content-Type"] = "text/plain"
		}
		if _, ok := headers["Content-Length"]; !ok {
			res.Headers["Content-Length"] = strconv.Itoa(len(body))
		}

		for key, value := range headers {
			res.Headers[key] = value
		}
	}

	return res
}

func (r *Response) String() string {
	var headersBuilder strings.Builder
	for key, value := range r.Headers {
		headersBuilder.WriteString(fmt.Sprintf("%s: %s%s", key, value, CRLF))
	}

	return fmt.Sprintf(
		"%s %d %s%s%s%s%s",
		r.Version, r.StatusCode, r.Message, CRLF, headersBuilder.String(), CRLF, r.Body,
	)
}

func readRequest(conn net.Conn) (Request, error) {
	buffer := make([]byte, 1024)

	_, err := conn.Read(buffer)
	if err != nil {
		return Request{}, fmt.Errorf("error reading request: %v", err)
	}
	req := NewRequest(&buffer)

	return req, nil
}

func getResponse(req Request) Response {
	var res Response

	switch {
	case req.Target == "/":
		res = NewResponse(200, "", nil)

	case strings.HasPrefix(req.Target, "/echo/"):
		value := strings.TrimPrefix(req.Target, "/echo/")
		res = NewResponse(200, value, nil)

	case req.Target == "/user-agent":
		useragent := req.Headers["user-agent"]
		res = NewResponse(200, useragent, nil)

	case strings.HasPrefix(req.Target, "/files/"):
		fileName := strings.TrimPrefix(req.Target, "/files/")
		filePath := filepath.Join(dir, filepath.Base(fileName))

		if req.Method == "GET" {
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				res = NewResponse(404, "", nil)
				break
			}

			res = NewResponse(200, string(fileContent), map[string]string{
				"Content-Type": "application/octet-stream",
			})
		} else {
			// req.Body = "pineapple raspberry pear mango apple blueberry strawberry banana"
			fmt.Println("'", req.Body, "'", len(req.Body))
			err := os.WriteFile(filePath, []byte(req.Body), 0777)
			if err != nil {
				res = NewResponse(500, "", nil)
				break
			}

			res = NewResponse(201, "", nil)
		}

	default:
		res = NewResponse(404, "", nil)
	}

	return res
}

func handleConnection(conn net.Conn) {
	req, err := readRequest(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
	res := getResponse(req)

	_, err = conn.Write([]byte(res.String()))
	if err != nil {
		fmt.Println(err)
	}
}

func main() {

	flag.StringVar(&dir, "directory", "./", "Static files directory path(absolute)")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn) // Start a new goroutine for each connection
	}

}
