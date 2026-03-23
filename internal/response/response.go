package response

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode int
type writerState int

const (
	StatusCodeSuccess             StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
	// WriterStates
	writerStateStatusLine writerState = 0
	writerStateHeaders    writerState = 1
	writerStateBody       writerState = 2
)

type Writer struct {
	w           io.Writer
	writerState writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:           w,
		writerState: writerStateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("Incorrect Writer State for Status Line: %d", w.writerState)
	}

	var reason string
	switch statusCode {
	case StatusCodeSuccess:
		reason = "OK"
	case StatusCodeBadRequest:
		reason = "Bad Request"
	case StatusCodeInternalServerError:
		reason = "Internal Server Error"
	}
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reason)
	_, err := w.w.Write([]byte(statusLine))
	if err != nil {
		return err
	}
	w.writerState = writerStateHeaders
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("Incorrect Writer State for Headers: %d", w.writerState)
	}

	for k, v := range headers {
		_, err := w.w.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	_, err := w.w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.writerState = writerStateBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("Incorrect Writer State for Body: %d", w.writerState)
	}

	n, err := w.w.Write(p)
	return n, err
}

//func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
//	var reason string
//	switch statusCode {
//	case StatusCodeSuccess:
//		reason = "OK"
//	case StatusCodeBadRequest:
//		reason = "Bad Request"
//	case StatusCodeInternalServerError:
//		reason = "Internal Server Error"
//	}
//	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reason)
//	_, err := w.Write([]byte(statusLine))
//	return err
//}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

//func WriteHeaders(w io.Writer, headers headers.Headers) error {
//	for k, v := range headers {
//		_, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
//		if err != nil {
//			return err
//		}
//	}
//	_, err := w.Write([]byte("\r\n"))
//	return err
//}
