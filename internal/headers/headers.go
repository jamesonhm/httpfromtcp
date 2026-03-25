package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
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
	if v, exists := h[key]; exists {
		value = strings.Join([]string{v, value}, ", ")
	}
	h[key] = value
}

func (h Headers) Update(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Get(key string) (value string, ok bool) {
	key = strings.ToLower(key)
	v, ok := h[key]
	return v, ok
}

func (h Headers) Delete(key string) (value string, ok bool) {
	key = strings.ToLower(key)
	if value, ok := h[key]; ok {
		delete(h, key)
		return value, true
	}
	return "", false
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
