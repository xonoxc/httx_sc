package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\ncocococ:cbababab\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "cbababab", headers.Get("cocococ"))
	assert.Equal(t, "", headers.Get("MissingKey"))
	assert.Equal(t, 43, n)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069 \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 26, n)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 29, n)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nHost: localhost:3000\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069,localhost:3000", headers.Get("Host"))
	assert.Equal(t, 47, n)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nHost:  localhost:3000 \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069,localhost:3000", headers.Get("Host"))
	assert.Equal(t, 49, n)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069 \r\nHost:  localhost:3000 \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069,localhost:3000", headers.Get("Host"))
	assert.Equal(t, 50, n)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	require.Empty(t, headers.headers)
	assert.Error(t, err, ErrMalformedFieldName)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069 \r\n Host:  localhost:3000 \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotEmpty(t, headers.headers)
	assert.NoError(t, err)
	assert.Equal(t, 51, n)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("  Host: localhost:42069    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	require.Empty(t, headers.headers)
	assert.Error(t, err, ErrMalformedFieldName)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
