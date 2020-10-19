// +build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gojek/turing/api/turing/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
Pre:
testUpdateRouterInvalidConfig run successfully.
Steps:
Invoke Router Version Deploy for the failed version
a. Accepted response received immediately
b. GET router version > status is failed
c. GET router > config still points to previously deployed version, currently undeployed
*/
func TestDeployRouterInvalidConfig(t *testing.T) {
	// Read existing router that MUST have already exists from previous create router e2e test
	// Router name is assumed to follow this format: e2e-experiment-{{.TestID}}
	routerName := "e2e-experiment-" + globalTestContext.TestID
	t.Log(fmt.Sprintf("Retrieving router with name '%s' created from previous test step", routerName))
	existingRouter, err := getRouterByName(globalTestContext.httpClient, globalTestContext.APIBasePath, globalTestContext.ProjectID, routerName)
	require.NoError(t, err)

	// Deploy router version
	url := fmt.Sprintf(
		"%s/projects/%d/routers/%d/versions/2/deploy",
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID, existingRouter.ID,
	)
	t.Log("Deploying router: POST " + url)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	require.NoError(t, err)
	response, err := globalTestContext.httpClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, response.StatusCode)

	// Wait for the version status to to change to success/failed deployment
	t.Log("Waiting for router to deploy")
	err = waitDeployVersion(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(existingRouter.ID),
		2,
	)
	require.NoError(t, err)

	// Test router version configuration
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
	// the expected version 1 is the valid version that the deployment fallback to due to invalid config
	assert.Equal(t, uint(1), router.CurrRouterVersion.Version)
	assert.Equal(t, models.RouterVersionStatusUndeployed, router.CurrRouterVersion.Status)
	assert.Equal(t, models.RouterStatusUndeployed, router.Status)
}
