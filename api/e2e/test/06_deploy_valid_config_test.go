//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gojek/turing/api/turing/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
Pre:
testCreateRouter run successfully, router not currently deployed.
Steps:
Invoke Router Deploy (will deploy the current config)
a. Wait until deployment timeout
b. Test GET router > config section shows version 1
c. Test cluster state to check that the components are deployed
*/
func TestDeployValidConfig(t *testing.T) {
	// Read existing router that MUST have already exists from previous create router e2e test
	// Router name is assumed to follow this format: e2e-experiment-{{.TestID}}
	routerName := "e2e-experiment-" + globalTestContext.TestID
	t.Log(fmt.Sprintf("Retrieving router with name '%s' created from previous test step", routerName))
	existingRouter, err := getRouterByName(
		globalTestContext.httpClient, globalTestContext.APIBasePath, globalTestContext.ProjectID, routerName)
	require.NoError(t, err)

	// Deploy router
	url := fmt.Sprintf("%s/projects/%d/routers/%d/deploy", globalTestContext.APIBasePath,
		globalTestContext.ProjectID, existingRouter.ID)
	t.Log("Deploying router: POST " + url)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	response, err := globalTestContext.httpClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, response.StatusCode, readBody(t, response))
	responseBody, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)
	t.Log("Deploy Response:", string(responseBody))

	// Wait for the version status to to change to success/failed deployment
	t.Log("Waiting for router to deploy")
	err = waitDeployVersion(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(existingRouter.ID),
		1, // router version 1 is a valid version
	)
	require.NoError(t, err)

	// Test router configuration
	t.Log("Testing GET router")
	router, err := getRouter(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(existingRouter.ID),
	)
	require.NoError(t, err)
	require.NotNil(t, router.CurrRouterVersion)
	assert.Equal(t, 1, int(router.CurrRouterVersion.Version))
	assert.Equal(t,
		fmt.Sprintf("http://%s-turing-router.%s.%s/v1/predict",
			router.Name,
			globalTestContext.ProjectName,
			globalTestContext.KServiceDomain,
		),
		router.Endpoint,
	)
	assert.Equal(t, models.RouterStatusDeployed, router.Status)

	// Check router, ensembler deployments and configmap have been removed
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
	assert.True(t, isConfigMapExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-fiber-config-%d", router.Name, 1)))
}
