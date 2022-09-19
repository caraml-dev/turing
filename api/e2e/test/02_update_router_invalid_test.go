//go:build e2e

package e2e

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

/*
Pre:
testCreateRouter run successfully.
Steps:
Update existing router with invalid config - High resource requests
a. Wait for failed deployment
b. GET router version 2 > shows status "failed"
c. GET router > config still shows version 1
d. Test cluster state that all deployments exist
e. Make a request to the router, validate the response.
*/
func TestUpdateRouterInvalidConfig(t *testing.T) {
	// Read existing router that MUST have already exists from previous create router e2e test
	// Router name is assumed to follow this format: e2e-experiment-{{.TestID}}
	routerName := "e2e-experiment-" + globalTestContext.TestID
	t.Logf("Retrieving router with name '%s' created from previous test step", routerName)
	existingRouter, err := getRouterByName(
		globalTestContext.httpClient, globalTestContext.APIBasePath, globalTestContext.ProjectID, routerName)
	require.NoError(t, err)

	// Read router config test data
	data := makeRouterPayload(filepath.Join("testdata", "update_router_high_cpu.json.tmpl"), globalTestContext)

	// Update router
	url := fmt.Sprintf(
		"%s/projects/%d/routers/%d",
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		existingRouter.ID,
	)
	t.Log("Updating router: PUT " + url)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := globalTestContext.httpClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, readBody(t, resp))

	// Until deployment timeout, wait for router version status to change
	t.Log("Waiting for router version to deploy")
	err = waitDeployVersion(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(existingRouter.ID),
		2,
	)
	assert.NoError(t, err)

	// Get router version 2
	t.Log("Testing GET router version")
	routerVersion, err := getRouterVersion(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(existingRouter.ID),
		2,
	)
	require.NoError(t, err)
	assert.Equal(t, models.RouterVersionStatusFailed, routerVersion.Status)
	assert.NotEmpty(t, routerVersion.Error)

	t.Log("Ensure existing router do not update the version to failed version i.e the version still unchanged at 1")
	router, err := getRouter(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(existingRouter.ID),
	)
	require.NoError(t, err)
	require.NotNil(t, router.CurrRouterVersion)
	assert.Equal(t, 1, int(router.CurrRouterVersion.Version))

	downstream, err := getRouterDownstream(globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-router", router.Name))
	assert.NoError(t, err)
	assert.Equal(t, downstream, fmt.Sprintf("%s-turing-router-%d.%s.%s",
		router.Name, 1, globalTestContext.ProjectName, globalTestContext.KServiceDomain))

	// Check that previous enricher, router, ensembler deployments exist
	t.Log("Checking cluster state")
	assert.True(t, isDeploymentExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-router-%d-0-deployment", router.Name, 1)))
	assert.True(t, isDeploymentExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-enricher-%d-0-deployment", router.Name, 1)))
	assert.True(t, isDeploymentExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-ensembler-%d-0-deployment", router.Name, 1)))

	// Make request to router
	t.Log("Testing router endpoint")
	router, err = getRouter(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(router.ID),
	)
	require.NoError(t, err)
	assert.Equal(t,
		fmt.Sprintf(
			"http://%s-turing-router.%s.%s/v1/predict",
			router.Name,
			globalTestContext.ProjectName,
			globalTestContext.KServiceDomain,
		),
		router.Endpoint,
	)

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
}
