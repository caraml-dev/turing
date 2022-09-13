package testutils

import (
	"bytes"
	"io"
	"net/http"

	mchttp "github.com/caraml-dev/turing/engines/router/missionctl/http"
)

// MakeTestMisisonControlResponse makes a success response with a dummy json body for testing
func MakeTestMisisonControlResponse() mchttp.Response {
	httpResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"data": "test"}`))),
		Header:     http.Header{},
	}
	mcResp, _ := mchttp.NewCachedResponseFromHTTP(httpResponse)
	return mcResp
}
