// Package headers provides functionality to parse HTTP headers from a byte slice.
package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

var SEPERATOR = []byte("\r\n")

var (
	ErrMalformedHeader    = errors.New("malformed header")
	ErrMalformedFieldName = errors.New("malformed field name")
)

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", ErrMalformedHeader
	}

	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	if len(name) == 0 || name[0] == ' ' || name[0] == '\t' || name[len(name)-1] == ' ' {
		return "", "", ErrMalformedFieldName
	}

	if !isTokenicallyValidFieldName(name) {
		return "", "", ErrMalformedFieldName
	}

	return string(name), string(value), nil
}

func NewHeaders() *Headers {
	return &Headers{
		headers: make(map[string]string),
	}
}

func (h *Headers) Map(cb func(k, v string)) {
	fmt.Println(h.headers)

	for k, v := range h.headers {
		cb(k, v)
	}
}

func (h *Headers) Get(name string) (string, bool) {
	headers, ok := h.headers[strings.ToLower(name)]
	return headers, ok
}

func (h *Headers) Set(name, value string) {
	key := strings.ToLower(name)

	exisitingValue, exists := h.Get(key)
	if exists {
		h.headers[key] = strings.Join([]string{exisitingValue, value}, ",")
		return
	}
	h.headers[key] = value
}

func isTokenicallyValidFieldName(fieldName []byte) bool {
	specials := []byte("!#$%&'*+-.^_`|~")

	for _, c := range fieldName {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || bytes.IndexByte(specials, c) != -1 {
			continue
		}
		return false
	}

	return true
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], SEPERATOR)
		if idx == -1 {
			break
		}

		if idx == 0 {
			done = true
			read += len(SEPERATOR)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			if len(h.headers) > 0 {
				read += idx + len(SEPERATOR)
				continue
			}
			return 0, false, err
		}
		h.Set(name, value)

		read += idx + len(SEPERATOR)
	}

	return read, done, nil
}
