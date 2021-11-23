// +build e2e

package e2e

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
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
func TestCreateRouter(t *testing.T) {
	// Create router
	t.Log("Creating router")
	data := makeRouterPayload(
		filepath.Join("testdata", "create_router_nop_logger_nop_exp.json.tmpl"),
		globalTestContext)

	withDeployedRouter(t, data,
		func(router *models.Router) {
			t.Log("Testing router endpoint: POST " + router.Endpoint)
			withRouterResponse(t,
				http.MethodPost,
				router.Endpoint,
				nil,
				"{}",
				func(response *http.Response, responsePayload []byte) {
					assert.Equal(t, http.StatusOK, response.StatusCode,
						"Unexpected response (code %d): %s",
						response.StatusCode, string(responsePayload))
					actualResponse := gjson.GetBytes(responsePayload, "json.response").String()
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
					assert.JSONEq(t, expectedResponse, actualResponse)
				})

			batchEndpoint := strings.Replace(router.Endpoint, "/predict", "/batch_predict", -1)
			t.Log("Testing router batch endpoint: POST " + batchEndpoint)
			withRouterResponse(t,
				http.MethodPost,
				batchEndpoint,
				nil,
				"[{},{}]",
				func(response *http.Response, responsePayload []byte) {
					t.Log(string(responsePayload))
					assert.Equal(t, http.StatusOK, response.StatusCode,
						"Unexpected response (code %d): %s",
						response.StatusCode, string(responsePayload))
					actualResponse := gjson.GetBytes(responsePayload, "#.data.json.response").String()
					expectedResponse := `[{
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
					},
					{
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
					}
					]`
					assert.JSONEq(t, expectedResponse, actualResponse)
				})

			t.Log("Test endpoints for router logs")
			baseURL, projectID, routerID := globalTestContext.APIBasePath, globalTestContext.ProjectID, router.ID
			url := fmt.Sprintf("%s/projects/%d/routers/%d/logs", baseURL, projectID, routerID)
			componentTypes := []string{"router", "ensembler", "enricher"}
			var podLogs []service.PodLog

			for _, c := range componentTypes {
				queryString := ""
				if c != "" {
					queryString = "?component_type=" + c
				}
				t.Log("GET", url+queryString)
				resp, err := http.Get(url + queryString)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				podLogs = getPodLogs(t, resp)
				assert.Greater(t, len(podLogs), 0)
			}
		},
		nil,
	)
}
