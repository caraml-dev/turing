// +build e2e

package e2e

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

func readBody(t *testing.T, resp *http.Response) string {
	if resp.Body == nil {
		return ""
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func readFile(filepath string) ([]byte, error) {
	// Open file
	fileObj, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer fileObj.Close()
	// Read contents
	byteValue, err := ioutil.ReadAll(fileObj)
	if err != nil {
		return nil, err
	}
	return byteValue, nil
}

// make payload to create new router from a template file
func makeRouterPayload(payloadTemplateFile string, args TestContext) []byte {
	data, err := ioutil.ReadFile(payloadTemplateFile)
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

	req, err := http.NewRequest(method, url, ioutil.NopCloser(bytes.NewReader([]byte(body))))
	require.NoError(t, err)

	req.Header = headers

	resp, err := globalTestContext.httpClient.Do(req)
	require.NoError(t, err)

	responseBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assertion(resp, responseBytes)
}
