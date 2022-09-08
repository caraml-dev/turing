//go:build e2e

package e2e

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

					actualResponse := string(payload)
					expectedResponse := `{
  "request": {},
  "response": {
    "experiment": {},
    "route_responses": [
      {
        "data": {
          "version": "treatment-a"
        },
        "is_default": false,
        "route": "treatment-a"
      }
    ]
  }
}`
					assert.JSONEq(t, expectedResponse, actualResponse)
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
					actualResponse := string(payload)

					expectedResponse := `{
  "request": {
    "service_type": {
       "id": "service-type-b"
    }
  },
  "response": {
    "experiment": {},
    "route_responses": [
      {
        "data": {
          "version": "treatment-b"
        },
        "is_default": false,
        "route": "treatment-b"
      }
    ]
  }
}`
					assert.JSONEq(t, expectedResponse, actualResponse)
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
					actualResponse := string(payload)

					expectedResponse := `{
  "request": {
    "service_type": {
      "id": "service-type-c"
    }
  },
  "response": {
    "experiment": {},
    "route_responses": [
      {
        "data": {
          "version": "control"
        },
        "is_default": false,
        "route": "control"
      }
    ]
  }
}`
					assert.JSONEq(t, expectedResponse, actualResponse)
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
