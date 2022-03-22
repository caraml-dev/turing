//go:build e2e
// +build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/gojek/turing/api/turing/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertRouterResponse asserts that two JSON strings are equivalent,
// but ignores the order of `response.route_responses` slice
func assertRouterResponse(t *testing.T, expected, actual string) {
	type response struct {
		Request  map[string]interface{}
		Response *struct {
			Experiment     interface{}
			RouteResponses []interface{}
		}
	}

	var expectedJSONAsInterface, actualJSONAsInterface response

	err := json.Unmarshal([]byte(expected), &expectedJSONAsInterface)
	require.NoError(t, err, fmt.Sprintf(
		"Expected value ('%s') is not valid json.\nJSON parsing error: '%v'", expected, err))

	err = json.Unmarshal([]byte(actual), &actualJSONAsInterface)
	require.NoError(t, err, fmt.Sprintf("Input ('%s') needs to be valid json.\nJSON parsing error: '%v'", actual, err))

	assert.Equal(t, expectedJSONAsInterface.Request, actualJSONAsInterface.Request)
	if expectedJSONAsInterface.Response != nil && actualJSONAsInterface.Response != nil {
		assert.Equal(t, expectedJSONAsInterface.Response.Experiment, actualJSONAsInterface.Response.Experiment)
		assert.ElementsMatch(t,
			expectedJSONAsInterface.Response.RouteResponses, actualJSONAsInterface.Response.RouteResponses)
	} else {
		assert.Equal(t, expectedJSONAsInterface.Response, actualJSONAsInterface.Response)
	}
}

func TestDeployRouterWithTrafficRules(t *testing.T) {
	// Create router
	t.Log("Creating router")
	data := makeRouterPayload(
		filepath.Join("testdata", "create_router_with_traffic_rules.json.tmpl"),
		globalTestContext)

	withDeployedRouter(t, data,
		func(router *models.Router) {
			t.Log("Testing router endpoint with the request, that satisfies first traffic rule: POST " +
				router.Endpoint)
			withRouterResponse(t,
				http.MethodPost,
				router.Endpoint,
				http.Header{
					"Content-Type":  {"application/json"},
					"X-Region":      {"region-a"},
					"X-Mirror-Body": {"true"},
				},
				"{}",
				func(response *http.Response, payload []byte) {
					require.Equal(t, http.StatusOK, response.StatusCode,
						"Unexpected response (code %d): %s", response.StatusCode, string(payload))

					actualResponse := gjson.GetBytes(payload, "response").String()
					expectedResponse := `{
  "experiment": {},
  "route_responses": [
    {
      "data": {
        "version": "treatment-a"
      },
      "is_default": false,
      "route": "treatment-a"
    },
    {
      "data": {
        "version": "control"
      },
      "is_default": true,
      "route": "control"
    }
  ]
}`
					assertRouterResponse(t, expectedResponse, actualResponse)
				},
			)

			t.Log("Testing router endpoint with the request, that satisfies second traffic rule: POST " +
				router.Endpoint)
			withRouterResponse(t,
				http.MethodPost,
				router.Endpoint,
				http.Header{
					"Content-Type":  {"application/json"},
					"X-Mirror-Body": {"true"},
				},
				`{"service_type": {"id": "service-type-b"}}`,
				func(response *http.Response, payload []byte) {
					require.Equal(t, http.StatusOK, response.StatusCode,
						"Unexpected response (code %d): %s", response.StatusCode, string(payload))
					actualRequest := gjson.GetBytes(payload, "request").String()
					actualResponse := gjson.GetBytes(payload, "response").String()

					expectedRequest := `{"service_type":{"id":"service-type-b"}}`
					expectedResponse := `{
  "experiment": {},
  "route_responses": [
    {
      "data": {
        "version": "control"
      },
      "is_default": true,
      "route": "control"
    },
    {
      "data": {
        "version": "treatment-b"
      },
      "is_default": false,
      "route": "treatment-b"
    }
  ]
}`
					assertRouterResponse(t, expectedRequest, actualRequest)
					assertRouterResponse(t, expectedResponse, actualResponse)
				},
			)

			t.Log("Testing router endpoint with the request, that doesn't satisfy any traffic rule: POST " +
				router.Endpoint)
			withRouterResponse(t,
				http.MethodPost,
				router.Endpoint,
				http.Header{
					"Content-Type":  {"application/json"},
					"X-Mirror-Body": {"true"},
				},
				`{"service_type": {"id": "service-type-c"}}`,
				func(response *http.Response, payload []byte) {
					require.Equal(t, http.StatusOK, response.StatusCode,
						"Unexpected response (code %d): %s", response.StatusCode, string(payload))

					actualRequest := gjson.GetBytes(payload, "request").String()
					actualResponse := gjson.GetBytes(payload, "response").String()

					expectedRequest := `{"service_type": {"id": "service-type-c"}}`
					expectedResponse := `{
                      "experiment": {},
                      "route_responses": [
                        {
                          "data": {
                            "version": "control"
                          },
                          "is_default": true,
                          "route": "control"
                        }
                      ]
                    }`
					assertRouterResponse(t, expectedRequest, actualRequest)
					assertRouterResponse(t, expectedResponse, actualResponse)
				},
			)
		},
		func(router *models.Router) {
			deleteExperiments(
				globalTestContext.clusterClients,
				globalTestContext.ProjectName,
				[]struct {
					Name       string
					MaxVersion int
				}{
					{
						Name:       router.Name,
						MaxVersion: 1,
					},
				},
			)
		})
}
