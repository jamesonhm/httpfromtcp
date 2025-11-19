package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

type parseState int

const (
	initialized parseState = 0
	done        parseState = 9
)

const (
	crlf       = "\r\n"
	bufferSize = 8
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	ParseState  parseState
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.ParseState {
	case initialized:
		reqLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *reqLine
		r.ParseState = done
		return n, nil
	case done:
		return 0, fmt.Errorf("error: trying to read data in done state")
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := Request{
		ParseState: initialized,
	}
	buffer := make([]byte, bufferSize)
	readToIndex := 0
	//var bytesParsed int
	for req.ParseState != done {
		if readToIndex >= len(buffer) {
			newBuff := make([]byte, len(buffer)*2)
			copy(newBuff, buffer)
			buffer = newBuff
		}
		n, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.ParseState = done
				break
			}
			return nil, err
		}
		readToIndex += n
		parsed, err := req.parse(buffer[:readToIndex])
		if err != nil {
			return nil, err
		}
		copy(buffer, buffer[parsed:readToIndex])
		readToIndex -= parsed
	}

	return &req, nil
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
	return requestLine, idx, nil
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
