//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
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
		filepath.Join("testdata", "create_router_nop_logger_proprietary_exp.json.tmpl"),
		globalTestContext)

	withDeployedRouter(t, data,
		func(router *models.Router) {
			t.Log("Testing router endpoint: POST " + router.Endpoint)
			expectedEndpoint := fmt.Sprintf(
				"http://%s-turing-router.%s.%s/v1/predict",
				router.Name,
				globalTestContext.ProjectName,
				globalTestContext.KServiceDomain,
			)
			assert.Equal(t, expectedEndpoint, router.Endpoint)
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
								"route_name":"treatment-a"
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

			batchEndpoint := strings.Replace(router.Endpoint, "/predict", "/batch_predict", -1)
			t.Log("Testing router batch endpoint: POST " + batchEndpoint)
			withRouterResponse(t,
				http.MethodPost,
				batchEndpoint,
				http.Header{
					"Content-Type":  {"application/json"},
					"X-Mirror-Body": {"true"},
				},
				`[{"client": {"id": 4}}, {"client": {"id": 7}}]`,
				func(response *http.Response, responsePayload []byte) {
					assert.Equal(t, http.StatusOK, response.StatusCode,
						"Unexpected response (code %d): %s",
						response.StatusCode, string(responsePayload))
					t.Logf("Response Payload:\n%s", string(responsePayload))
					expectedResponse := `[
						{
							"code": 200,
							"data": {
								"request": {
									"client": {
										"id": 4
									}
								},
								"response": {
									"experiment": {
										"configuration": {
											"foo": "bar",
											"route_name":"treatment-a"
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
								}
							}
						},
						{
							"code": 200,
							"data": {
								"request": {
									"client": {
										"id": 7
									}
								},
								"response": {
									"experiment": {
										"configuration": {
											"bar": "baz",
											"route_name":"control"
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
								}
							}
						}
					]`
					assert.JSONEq(t, expectedResponse, string(responsePayload))
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
