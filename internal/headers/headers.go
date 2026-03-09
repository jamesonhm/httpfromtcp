package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers(make(map[string]string))
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	// not enough data given
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		// the empty line
		// headers are done, consume the CRLF
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	key := string(parts[0])

	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	value := bytes.TrimSpace(parts[1])
	key = strings.TrimSpace(key)
	if !validKey([]byte(key)) {
		return 0, false, fmt.Errorf("invalid header token")
	}

	h.Set(key, string(value))
	return idx + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func validKey(key []byte) bool {
	for _, c := range key {
		if !validChar(c) {
			return false
		}
	}
	return true
}

func validChar(c byte) bool {
	specialChars := []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}
	if c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z' ||
		c >= '0' && c <= '9' {
		return true
	}

	return slices.Contains(specialChars, c)
}
