//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
Pre:
testUpdateRouterInvalidConfig run successfully.
Steps:
Invoke Router Undeploy
a. Success response received immediately
b. GET router > empty config
c. Test cluster that all deployments removed
*/
func TestUndeployRouter(t *testing.T) {
	// Read existing router that MUST have already exists from previous create router e2e test
	// Router name is assumed to follow this format: e2e-experiment-{{.TestID}}
	routerName := "e2e-experiment-" + globalTestContext.TestID
	t.Logf("Retrieving router with name '%s' created from previous test step", routerName)
	existingRouter, err := getRouterByName(
		globalTestContext.httpClient, globalTestContext.APIBasePath, globalTestContext.ProjectID, routerName)
	require.NoError(t, err)

	// Undeploy router
	t.Log("Undeploying router")
	url := fmt.Sprintf("%s/projects/%d/routers/%d/undeploy", globalTestContext.APIBasePath,
		globalTestContext.ProjectID, existingRouter.ID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	require.NoError(t, err)
	response, err := globalTestContext.httpClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	responseBody, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	require.NoError(t, err)
	t.Log("Undeploy Response:", string(responseBody))

	// Wait for delete timeout
	t.Log("Wait Undeploy router")
	time.Sleep(time.Second * deleteTimeoutSeconds)

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
	assert.Equal(t, models.RouterVersionStatusUndeployed, router.CurrRouterVersion.Status)
	assert.Equal(t, models.RouterStatusUndeployed, router.Status)
	assert.Equal(t, "", router.Endpoint)

	// Check router, ensembler deployments and configmap have been removed
	t.Log("Checking cluster state")
	assert.False(t, isDeploymentExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-router-%d-0-deployment", router.Name, 1)))
	assert.False(t, isDeploymentExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-enricher-%d-0-deployment", router.Name, 1)))
	assert.False(t, isDeploymentExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-ensembler-%d-0-deployment", router.Name, 1)))
	assert.False(t, isConfigMapExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-fiber-config", router.Name)))
	_, err = getRouterDownstream(globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-router", router.Name))
	assert.Equal(t, true, err != nil)
}
