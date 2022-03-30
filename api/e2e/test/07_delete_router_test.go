//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
Pre:
TestUpdateRouterStandardEnsemblerValidConfig run successfully.
Steps:
Invoke Undeploy Router.
a. Success response received immediately
b. Test cluster that all deployments removed
Invoke Delete Router.
a. Success response received immediately
b. Get Latest Router Version - not found
c. Get Router - not found
*/
func TestDeleteRouter(t *testing.T) {
	// Read existing router that MUST have already exists from previous create router e2e test
	// Router name is assumed to follow this format: e2e-experiment-{{.TestID}}
	routerName := "e2e-experiment-" + globalTestContext.TestID
	t.Log(fmt.Sprintf("Retrieving router with name '%s' created from previous test step", routerName))
	existingRouter, err := getRouterByName(
		globalTestContext.httpClient, globalTestContext.APIBasePath, globalTestContext.ProjectID, routerName)
	require.NoError(t, err)

	// Issue delete request
	url := fmt.Sprintf("%s/projects/%d/routers/%d/undeploy", globalTestContext.APIBasePath,
		globalTestContext.ProjectID, existingRouter.ID)
	t.Log("Undeploy router: POST " + url)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	require.NoError(t, err)
	response, err := globalTestContext.httpClient.Do(req)
	require.NoError(t, err)
	defer response.Body.Close()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	// Wait for delete timeout
	t.Log(fmt.Sprintf("Wait %d seconds to undeploy router", deleteTimeoutSeconds))
	time.Sleep(time.Second * deleteTimeoutSeconds)

	// Get router
	t.Log("Testing GET router")
	router, err := getRouter(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(existingRouter.ID),
	)
	assert.NoError(t, err)

	// Test cluster state
	t.Log("Checking cluster state")
	assert.False(t, isDeploymentExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-router-%d-0-deployment", router.Name, 3)))
	assert.False(t, isDeploymentExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-enricher-%d-0-deployment", router.Name, 3)))
	assert.False(t, isDeploymentExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-ensembler-%d-0-deployment", router.Name, 3)))
	assert.False(t, isDeploymentExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-fluentd-logger-0-%d", router.Name, 3)))
	assert.False(t, isConfigMapExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-fiber-config-%d", router.Name, 3)))
	assert.False(t, isPersistentVolumeClaimExists(globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-cache-volume-%d", router.Name, 3)))

	url = fmt.Sprintf("%s/projects/%d/routers/%d", globalTestContext.APIBasePath,
		globalTestContext.ProjectID, existingRouter.ID)
	t.Log("Delete router: DELETE " + url)
	req, err = http.NewRequestWithContext(context.Background(), http.MethodDelete, url, nil)
	require.NoError(t, err)
	response, err = globalTestContext.httpClient.Do(req)
	require.NoError(t, err)
	defer response.Body.Close()
	assert.Equal(t, http.StatusOK, response.StatusCode)
	// Get router
	t.Log("Testing GET router")
	router, err = getRouter(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(existingRouter.ID),
	)
	assert.NoError(t, err)
	assert.Equal(t, 0, int(router.ID))
	_, err = getRouterDownstream(globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-router", router.Name))
	assert.Equal(t, true, err != nil)
}
