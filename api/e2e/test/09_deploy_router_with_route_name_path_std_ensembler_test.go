//go:build e2e

package e2e

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

/*
Steps:
Create a new router with valid config for enricher, ensembler, router.
a. Test GET router immediately > empty config
b. Wait for success response from deployment
c. Test GET router version > status shows "deployed"
d. Test GET router > config section shows version 1, status "deployed"
e. Test cluster that deployments exist
f. Make a request to the router, validate the response.
*/
func TestDeployRouterWithRouteNamePathInStandardEnsembler(t *testing.T) {
	// Create router
	t.Log("Creating router")
	data := makeRouterPayload(
		filepath.Join("testdata", "create_router_std_ensembler_route_name_path.json.tmpl"),
		globalTestContext)

	withDeployedRouter(t, data,
		func(router *models.Router) {
			t.Log("Testing router endpoint: POST " + router.Endpoint)
			withRouterResponse(t,
				http.MethodPost,
				router.Endpoint,
				http.Header{
					"Content-Type":  {"application/json"},
					"X-Mirror-Body": {"true"},
				},
				`{"client": {"id": 4}}`,
				func(response *http.Response, responsePayload []byte) {
					assert.Equal(t, http.StatusOK, response.StatusCode,
						"Unexpected response (code %d): %s",
						response.StatusCode, string(responsePayload))
					actualResponse := gjson.GetBytes(responsePayload, "response").String()
					expectedResponse := `{
					  "experiment": {
						"configuration": {
							"foo":"bar",
						}
					  },
					  "route_responses": [
						{
						  "data": {
							"version": "control"
						  },
						  "is_default": false,
						  "route": "control"
						}
					  ]
					}`
					assert.JSONEq(t, expectedResponse, actualResponse)
				})
		},
		nil,
	)
}
