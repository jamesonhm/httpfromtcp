package response

import (
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

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network conn
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

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

	statusLine := getStatusLine(statusCode)
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

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("Incorrect Writer State for Body: %d", w.writerState)
	}

	n, err := fmt.Fprintf(w.w, "%x\r\n%s\r\n", len(p), p)
	return n, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("Incorrect Writer State for Body: %d", w.writerState)
	}

	n, err := fmt.Fprintf(w.w, "0\r\n\r\n")
	return n, err
}

func getStatusLine(statusCode StatusCode) []byte {
	var reason string
	switch statusCode {
	case StatusCodeSuccess:
		reason = "OK"
	case StatusCodeBadRequest:
		reason = "Bad Request"
	case StatusCodeInternalServerError:
		reason = "Internal Server Error"
	}
	return []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reason))
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}
