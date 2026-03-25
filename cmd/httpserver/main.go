package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	var statusCode int
	var body string
	tgt := req.RequestLine.RequestTarget
	switch {
	case strings.HasPrefix(tgt, "/httpbin"):
		tgt = strings.TrimPrefix(tgt, "/httpbin")
		proxyHandler(w, tgt)
		return
	case tgt == "/yourproblem":
		statusCode = 400
		body = `
			<html>
			  <head>
				<title>400 Bad Request</title>
			  </head>
			  <body>
				<h1>Bad Request</h1>
				<p>Your request honestly kinda sucked.</p>
			  </body>
			</html>
		`
	case tgt == "/myproblem":
		statusCode = 500
		body = `
			<html>
			  <head>
				<title>500 Internal Server Error</title>
			  </head>
			  <body>
				<h1>Internal Server Error</h1>
				<p>Okay, you know what? This one is on me.</p>
			  </body>
			</html>
		`
	default:
		statusCode = 200
		body = `
			<html>
			  <head>
				<title>200 OK</title>
			  </head>
			  <body>
				<h1>Success!</h1>
				<p>Your request was an absolute banger.</p>
			  </body>
			</html>
		`
	}
	_ = w.WriteStatusLine(response.StatusCode(statusCode))

	b := []byte(body)
	h := response.GetDefaultHeaders(len(b))
	h.Update("Content-Type", "text/html")
	_ = w.WriteHeaders(h)

	_, _ = w.WriteBody(b)
}

func proxyHandler(w *response.Writer, tgt string) {
	resp, _ := http.Get(fmt.Sprintf("http://httpbin.org%s", tgt))

	w.WriteStatusLine(response.StatusCode(resp.StatusCode))

	h := response.GetDefaultHeaders(0)
	h.Delete("Content-Length")
	h.Set("Transfer-Encoding", "chunked")
	w.WriteHeaders(h)

	buf := make([]byte, 32)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.WriteChunkedBody(buf[:n])
		}
		if err == io.EOF {
			break
		}
	}
	w.WriteChunkedBodyDone()
}
