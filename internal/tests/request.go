// Package request provides a simple HTTP request parser that can parse HTTP/1.1 requests.
package request

import (
	"bytes"
	"errors"
	"io"
	"strconv"

	"tcp.scratch.i/internal/headers"
)

type ParserState string

const (
	StateInitialized ParserState = "init"
	StateDone        ParserState = "done"
	StateBody        ParserState = "body"
	StateHeader      ParserState = "headers"
	StateError       ParserState = "error"
)

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        string
	state       ParserState
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valueStr, exits := headers.Get(name)
	if !exits {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func newRequest() *Request {
	return &Request{
		state:   StateInitialized,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
}

func (r *Request) parse(b []byte) (int, error) {
	read := 0

dance:
	for {
		currentData := b[read:]
		if len(currentData) == 0 {
			break dance
		}

		switch r.state {
		case StateError:
			return 0, ErrRequestInErrorState

		case StateInitialized:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if rl == nil && n == 0 {
				break dance
			}

			r.RequestLine = *rl
			read += n

			r.state = StateHeader

		case StateHeader:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break dance
			}

			read += n

			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			contentLen := getInt(r.Headers, "Content-Length", 0)

			if contentLen == 0 {
				panic("chunked not implemented")
			}

			remainingLen := min(contentLen-len(r.Body), len(currentData))
			r.Body += string(currentData[:remainingLen])
			read += remainingLen

			if len(r.Body) == contentLen {
				r.state = StateDone
			}

		case StateDone:
			break dance
		}
	}

	return read, nil
}

func (r *Request) hasBody() bool {
	contentLen := getInt(r.Headers, "Content-Length", 0)
	return contentLen > 0
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

func (r *RequestLine) ValidHTTP() bool {
	return r.HTTPVersion == "1.1"
}

var (
	ErrBadStartLine           = errors.New("malformend request-line")
	ErrIncompleteStartLine    = errors.New("incomplete startline")
	ErrUnsupportedHTTPVersion = errors.New("unsupported http version ! only http/1.1 is supported as of now")
	ErrRequestInErrorState    = errors.New("Request in error state")
)

var SEPERATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPERATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	restOfMsg := idx + len(SEPERATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrBadStartLine
	}

	parseHTTPVersion := bytes.Split(parts[2], []byte("/"))
	if len(parseHTTPVersion) != 2 || string(parseHTTPVersion[0]) != "HTTP" || string(parseHTTPVersion[1]) != "1.1" {
		return nil, restOfMsg, ErrBadStartLine
	}

	requestLine := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HTTPVersion:   string(parseHTTPVersion[1]),
	}

	if !requestLine.ValidHTTP() {
		return nil, restOfMsg, ErrUnsupportedHTTPVersion
	}

	return requestLine, restOfMsg, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// NOTE: this buffer can overrun easysily so we need to take care of this
	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, pErr := request.parse(buf[:bufLen])
		if pErr != nil {
			return nil, pErr
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN

		if err == io.EOF {
			if request.state != StateDone {
				return nil, io.ErrUnexpectedEOF
			}
			break
		}

	}

	return request, nil
}
