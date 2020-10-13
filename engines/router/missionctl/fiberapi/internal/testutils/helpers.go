package testutils

import (
	"net/http"
	"testing"

	"github.com/gojek/fiber"
	fiberhttp "github.com/gojek/fiber/http"
)

// NewFiberCallerWithHTTPDispatcher is a helper function to create an instance of Fiber caller in
// the test cases so the test cases are easier to read.
func NewFiberCallerWithHTTPDispatcher(t *testing.T, callerID string) fiber.Component {
	httpDispatcher, err := fiberhttp.NewDispatcher(http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	caller, err := fiber.NewCaller(callerID, httpDispatcher)
	if err != nil {
		t.Fatal(err)
	}
	return caller
}
