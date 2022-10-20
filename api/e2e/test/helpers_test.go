//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"text/template"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func readBody(t *testing.T, resp *http.Response) string {
	if resp.Body == nil {
		return ""
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// make payload to create new router from a template file
func makeRouterPayload(payloadTemplateFile string, args TestContext) []byte {
	data, err := os.ReadFile(payloadTemplateFile)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New("new router payload").Parse(string(data))
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, args); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// withRouterResponse sends request to the router
// and then asserts received response by using
// assertion function, passed as the argument
func withRouterResponse(
	t *testing.T,
	method, url string,
	headers http.Header,
	body string,
	assertion func(response *http.Response, responsePayload []byte)) {

	req, err := http.NewRequest(method, url, io.NopCloser(bytes.NewReader([]byte(body))))
	require.NoError(t, err)

	req.Header = headers

	resp, err := globalTestContext.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assertion(resp, responseBytes)
}

// withUPIRouterResponse sends request to UPI router
// and asserts received response by using assertion function
// the assertion function accept the response from the request
func withUPIRouterResponse(t *testing.T,
	client upiv1.UniversalPredictionServiceClient,
	headers metadata.MD,
	request *upiv1.PredictValuesRequest,
	assertion func(response *upiv1.PredictValuesResponse)) {
	ctx := metadata.NewOutgoingContext(context.Background(), headers)
	resp, err := client.PredictValues(ctx, request)
	require.NoError(t, err)

	assertion(resp)
}
