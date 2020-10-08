package http

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCachedResponse(t *testing.T) {
	// Make test HTTP response
	testResp := makeTestHTTPResponse()
	if testResp != nil && testResp.Body != nil {
		testResp.Body.Close()
	}
	// Create cached response from the HTTP response
	resp, err := NewCachedResponseFromHTTP(testResp)

	// Validate
	// Check that the error is nil
	assert.NoError(t, err)
	// Check the header and body
	assert.Equal(t, "header_value", resp.header.Get("header_key"))
}

func TestCachedResponseGetters(t *testing.T) {
	// Make test HTTP response
	testResp := makeTestHTTPResponse()
	if testResp != nil && testResp.Body != nil {
		testResp.Body.Close()
	}
	// Create cached response from the HTTP response
	resp, err := NewCachedResponseFromHTTP(testResp)

	// Validate
	// Check that the error is nil
	assert.NoError(t, err)
	// Check the header and body, with getters
	assert.Equal(t, "header_value", resp.Header().Get("header_key"))
	assert.Equal(t, []byte(`{"key": "value"}`), resp.Body())
}

func makeTestHTTPResponse() *http.Response {
	body := []byte(`{"key": "value"}`)
	header := http.Header{}
	header.Set("header_key", "header_value")

	testResp := http.Response{
		Header: header,
		Body:   ioutil.NopCloser(bytes.NewReader(body)),
	}

	return &testResp
}
