package main

import (
	"crypto/sha256"
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
	tgt := req.RequestLine.RequestTarget
	switch {
	case strings.HasPrefix(tgt, "/httpbin"):
		proxyHandler(w, req)
		return
	case tgt == "/yourproblem":
		handler400(w, req)
	case tgt == "/myproblem":
		handler500(w, req)
	case tgt == "/video":
		handlerVideo(w, req)
	default:
		handler200(w, req)
	}

}

func handlerVideo(w *response.Writer, req *request.Request) {
	v, err := os.ReadFile("./assets/vim.mp4")
	if err != nil {
		fmt.Println("error loading video:", err)
		handler500(w, req)
		return
	}
	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(len(v))
	h.Update("Content-Type", "video/mp4")
	w.WriteHeaders(h)
	w.WriteBody(v)
}

func handler500(w *response.Writer, _ *request.Request) {
	body := []byte(`
		<html>
		  <head>
			<title>500 Internal Server Error</title>
		  </head>
		  <body>
			<h1>Internal Server Error</h1>
			<p>Okay, you know what? This one is on me.</p>
		  </body>
		</html>
	`)
	w.WriteStatusLine(response.StatusCodeInternalServerError)
	h := response.GetDefaultHeaders(len(body))
	h.Update("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler400(w *response.Writer, _ *request.Request) {
	body := []byte(`
		<html>
		  <head>
			<title>400 Bad Request</title>
		  </head>
		  <body>
			<h1>Bad Request</h1>
			<p>Your request honestly kinda sucked.</p>
		  </body>
		</html>
		`)
	w.WriteStatusLine(response.StatusCodeBadRequest)
	h := response.GetDefaultHeaders(len(body))
	h.Update("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	body := []byte(`
		<html>
		  <head>
			<title>200 OK</title>
		  </head>
		  <body>
			<h1>Success!</h1>
			<p>Your request was an absolute banger.</p>
		  </body>
		</html>
		`)
	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(len(body))
	h.Update("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func proxyHandler(w *response.Writer, req *request.Request) {
	tgt := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	url := fmt.Sprintf("http://httpbin.org%s", tgt)
	fmt.Println("Proxying to", url)
	resp, err := http.Get(url)
	if err != nil {
		handler500(w, req)
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusCode(resp.StatusCode))

	h := response.GetDefaultHeaders(0)
	h.Delete("Content-Length")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders(h)

	const maxChunkSize = 1024
	buf := make([]byte, maxChunkSize)
	nTotal := 0
	fullBody := make([]byte, 0)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, err := w.WriteChunkedBody(buf[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
			nTotal += n
			fullBody = append(fullBody, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
	}
	w.WriteChunkedBodyDone()
	checksum := sha256.Sum256(fullBody)
	h.Set("X-Content-SHA256", fmt.Sprintf("%x", checksum))
	h.Set("X-Content-Length", fmt.Sprintf("%d", nTotal))
	w.WriteTrailers(h)
}
