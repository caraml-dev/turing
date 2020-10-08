package testutils

import (
	"bytes"
	"net/http"
	"testing"
)

// MakeTestRequest creates a dummy request for testing
func MakeTestRequest(
	t *testing.T,
	httpRequestModifier func(*http.Request),
) *http.Request {
	payload := `{"customer_id": "test_customer"}`
	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer([]byte(payload)))
	req.Header.Set("req_id", "test_req_id")
	FailOnError(t, err)
	httpRequestModifier(req)
	return req
}

// NopHTTPRequestModifier makes no modification to the input request
func NopHTTPRequestModifier(req *http.Request) {}
