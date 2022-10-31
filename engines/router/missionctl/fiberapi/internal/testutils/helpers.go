package testutils

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber"
	fiberGrpc "github.com/gojek/fiber/grpc"
	fiberHttp "github.com/gojek/fiber/http"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type mockDispatcher struct {
}

func (m mockDispatcher) Do(request fiber.Request) fiber.Response {
	return nil
}

// NewFiberCaller is a helper function to create an instance of Fiber caller in
// the test cases so the test cases are easier to read.
func NewFiberCaller(t testing.TB, callerID string) fiber.Component {
	caller, err := fiber.NewCaller(callerID, &mockDispatcher{})
	if err != nil {
		t.Fatal(err)
	}
	return caller
}

// NewHTTPFiberRequest create new http fiber request based on header and body
func NewHTTPFiberRequest(t testing.TB, header http.Header, body string) fiber.Request {
	r, err := fiberHttp.NewHTTPRequest(&http.Request{
		Header: header,
		Body:   ioutil.NopCloser(strings.NewReader(body)),
	})
	if err != nil {
		t.Fatal(err)
	}

	return r
}

// NewUPIFiberRequest create new grpc fiber request based on header and upi request
func NewUPIFiberRequest(t testing.TB, header map[string]string, request *upiv1.PredictValuesRequest) fiber.Request {
	m, err := proto.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	return &fiberGrpc.Request{
		Metadata: metadata.New(header),
		Message:  m,
		Proto:    request,
	}
}
