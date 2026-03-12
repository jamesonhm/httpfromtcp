package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	ParseState  parseState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parseState int

const (
	requestStateInitialized    parseState = 0
	requestStateParsingHeaders parseState = 2
	requestStateParsingBody    parseState = 4
	requestStateDone           parseState = 9
)

const (
	crlf       = "\r\n"
	bufferSize = 8
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		Headers:    headers.NewHeaders(),
		Body:       make([]byte, 0),
		ParseState: requestStateInitialized,
	}
	//var bytesParsed int
	for req.ParseState != requestStateDone {
		if readToIndex >= len(buffer) {
			newBuff := make([]byte, len(buffer)*2)
			copy(newBuff, buffer)
			buffer = newBuff
		}
		n, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.ParseState != requestStateDone {
					return nil, fmt.Errorf("incomplete request")
				}
				break
			}
			return nil, err
		}
		readToIndex += n
		parsed, err := req.parse(buffer[:readToIndex])
		if err != nil {
			return nil, err
		}
		copy(buffer, buffer[parsed:])
		readToIndex -= parsed
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.ParseState != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.ParseState {
	case requestStateInitialized:
		reqLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *reqLine
		r.ParseState = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.ParseState = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		// check for content length header
		lenStr, ok := r.Headers.Get("Content-Length")
		if !ok {
			r.ParseState = requestStateDone
			return 0, nil
		}
		contLength, err := strconv.Atoi(lenStr)
		if err != nil {
			return 0, err
		}
		r.Body = append(r.Body, data...)
		if len(r.Body) > contLength {
			return 0, fmt.Errorf("error: body longer than content-length header")
		}
		if len(r.Body) == contLength {
			r.ParseState = requestStateDone
			return len(data), nil
		}
		return len(data), nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in done state")
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, idx + 2, nil
}

func requestLineFromString(requestLine string) (*RequestLine, error) {
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("not enough parts in request line: %s.", requestLine)
	}

	for _, c := range parts[0] {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method in request line: %s.", parts[0])
		}
	}

	if parts[2] != "HTTP/1.1" {
		return nil, fmt.Errorf("invalid http version: %s", parts[2])
	}
	versionParts := strings.Split(parts[2], "/")

	reqLine := RequestLine{
		HttpVersion:   versionParts[1],
		RequestTarget: parts[1],
		Method:        parts[0],
	}

	return &reqLine, nil
}
