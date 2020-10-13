package http

import (
	"io/ioutil"
	"net/http"
)

// Response is an interface for a HTTP response-like structure that has cached data
type Response interface {
	Body() []byte
	Header() http.Header
}

// CachedResponse implements the Response interface, providing the ability to cache
// the critical pieces of information from a response object
type CachedResponse struct {
	header http.Header
	body   []byte
}

// Header returns the cached header of a HTTP response
func (r *CachedResponse) Header() http.Header {
	return r.header
}

// Body returns the cached body of a HTTP response
func (r *CachedResponse) Body() []byte {
	return r.body
}

// NewCachedResponseFromHTTP is a creator for a cached response object that accepts
// a HTTP response as input
func NewCachedResponseFromHTTP(r *http.Response) (*CachedResponse, error) {
	// Consume response body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Create cached response
	return &CachedResponse{
		header: r.Header,
		body:   data,
	}, nil
}

// NewCachedResponse is a constructor method that creates a cached response object
// from provided response payload and header
func NewCachedResponse(body []byte, header http.Header) Response {
	return &CachedResponse{
		header: header,
		body:   body,
	}
}
