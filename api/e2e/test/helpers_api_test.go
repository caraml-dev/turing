// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gojek/turing/api/turing/models"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getRouter(
	httpClient *http.Client,
	apiBasePath string,
	projectID int,
	routerID int,
) (*models.Router, error) {
	var router models.Router

	url := fmt.Sprintf("%s/projects/%d/routers/%d", apiBasePath, projectID, routerID)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBytes, &router)
	if err != nil {
		return nil, err
	}

	return &router, err
}

func getRouterByName(
	httpClient *http.Client,
	apiBasePath string,
	projectID int,
	routerName string,
) (*models.Router, error) {
	url := fmt.Sprintf("%s/projects/%d/routers", apiBasePath, projectID)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	routers := make([]*models.Router, 0)
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &routers); err != nil {
		return nil, err
	}

	for _, r := range routers {
		if r.Name == routerName {
			return r, nil
		}
	}

	return nil, fmt.Errorf("router with name '%s' not found", routerName)
}

func getRouterVersion(
	httpClient *http.Client,
	apiBasePath string,
	projectID int,
	routerID int,
	version int,
) (*models.RouterVersion, error) {
	var routerVersion models.RouterVersion

	url := fmt.Sprintf("%s/projects/%d/routers/%d/versions/%d",
		apiBasePath, projectID, routerID, version)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBytes, &routerVersion)
	if err != nil {
		return nil, err
	}

	return &routerVersion, err
}

// WaitDeployVersion gets the info for the given router version and checks that the
// its status != pending, until the deployment timeout
func waitDeployVersion(
	httpClient *http.Client,
	apiBasePath string,
	projectID int,
	routerID int,
	version int,
) error {
	timer := time.NewTimer(time.Second * deploymentWaitTimeoutSeconds)
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-timer.C:
			return errors.New("Timeout waiting for version to be deployed")
		case <-ticker.C:
			routerVersion, err := getRouterVersion(httpClient, apiBasePath,
				projectID, routerID, version)
			if err != nil {
				timer.Stop()
				return err
			}
			if routerVersion.Status != models.RouterVersionStatusPending {
				if routerVersion.Router.Status != models.RouterStatusPending {
					timer.Stop()
					return nil
				}
			}
		}
	}
}

func getPodLogs(t *testing.T, resp *http.Response) []service.PodLog {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	var podLogs []service.PodLog
	if err = json.Unmarshal(data, &podLogs); err != nil {
		t.Error(err)
	}
	return podLogs
}

// withDeployedRouter creates a new router from passed 'routerPayload' configuration,
// waits until this router is deployed and then asserts this router by using assertion
// function, that is passed as an argument
//
// cleanup - optional function, that can be used for cleaning up some resources from
//           the cluster, after assertion of the router is done
func withDeployedRouter(
	t *testing.T,
	routerPayload []byte,
	assertion func(router *models.Router),
	cleanup func(router *models.Router)) {

	createRouterAPI := fmt.Sprintf(
		"%s/projects/%d/routers",
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID)
	resp, err := http.Post(createRouterAPI, "application/json", bytes.NewReader(routerPayload))
	require.NoError(t, err)

	responsePayload, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	created := models.Router{}
	if err = json.Unmarshal(responsePayload, &created); err != nil {
		require.NoError(t, err)
	}
	t.Log(fmt.Sprintf("Created router with name: %s, ID: %d", created.Name, created.ID))

	t.Log("Ensure router has been created and current status is pending")
	router, err := getRouter(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(created.ID),
	)
	require.NoError(t, err)
	require.Nil(t, router.CurrRouterVersion)
	require.Equal(t, "", router.Endpoint)
	require.Equal(t, models.RouterStatusPending, router.Status)

	// Wait for the version status to to change to success/failed deployment
	t.Log("Waiting for router to deploy")
	err = waitDeployVersion(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(router.ID),
		1,
	)
	require.NoError(t, err)

	// Get router version with id 1
	t.Log("Testing GET router version and ensure the status is 'deployed'")
	routerVersion, err := getRouterVersion(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(router.ID),
		1, // version is 1, since it MUST be the only router version that exists, since we create a NEW router
	)
	require.NoError(t, err)
	assert.Equal(t, models.RouterVersionStatusDeployed, routerVersion.Status)

	// Get router with id 1 - check current version, status and endpoint
	t.Log("Testing GET router - new config")
	router, err = getRouter(
		globalTestContext.httpClient,
		globalTestContext.APIBasePath,
		globalTestContext.ProjectID,
		int(router.ID),
	)
	require.NoError(t, err)
	require.NotNil(t, router.CurrRouterVersion)

	assert.Equal(t, 1, int(router.CurrRouterVersion.Version))
	assert.Equal(t, models.RouterStatusDeployed, router.Status)

	expectedEndpoint := fmt.Sprintf(
		"http://%s-turing-router.%s.%s/v1/predict",
		router.Name,
		globalTestContext.ProjectName,
		globalTestContext.KServiceDomain,
	)
	assert.Equal(t, expectedEndpoint, router.Endpoint)

	t.Log("Ensure Istio virtual services are created successfully")
	downstream, err := getRouterDownstream(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-router", router.Name),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		fmt.Sprintf(
			"%s-turing-router-%d.%s.%s",
			router.Name,
			routerVersion.Version,
			globalTestContext.ProjectName,
			globalTestContext.KServiceDomain,
		),
		downstream,
	)

	t.Log(
		"Ensure Kubernetes deployment and ConfigMap objects for router, " +
			"enricher and ensembler are created successfully")
	require.True(t, isDeploymentExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-router-%d-0-deployment", router.Name, routerVersion.Version)))
	if router.CurrRouterVersion.Enricher != nil {
		require.True(t, isDeploymentExists(
			globalTestContext.clusterClients,
			globalTestContext.ProjectName,
			fmt.Sprintf("%s-turing-enricher-%d-0-deployment", router.Name, routerVersion.Version)))
	}
	if router.CurrRouterVersion.Ensembler != nil {
		require.True(t, isDeploymentExists(
			globalTestContext.clusterClients,
			globalTestContext.ProjectName,
			fmt.Sprintf("%s-turing-ensembler-%d-0-deployment", router.Name, routerVersion.Version)))
	}
	require.True(t, isConfigMapExists(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		fmt.Sprintf("%s-turing-fiber-config-%d", router.Name, routerVersion.Version)))

	defer func() {
		if cleanup != nil {
			cleanup(router)
		}
	}()

	assertion(router)
}
