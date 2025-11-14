package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

func RequestFromReader(reader io.Reader) (*Request, error) {
	rawBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	reqLine, err := parseRequestLine(rawBytes)
	if err != nil {
		return nil, err
	}
	return &Request{
		RequestLine: *reqLine,
	}, nil
}

func parseRequestLine(data []byte) (*RequestLine, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, fmt.Errorf("could not find CRLF in request-line")
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, err
	}
	return requestLine, nil
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
